package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

var (
	databaseDSN = flag.String("database-dsn", "root:root@/isu6fportal_day0", "database `dsn`")
	debugMode   = flag.Bool("debug", false, "enable debug mode")
)

var db *sql.DB
var templates = map[string]*template.Template{}
var sessionStore sessions.Store
var locJST *time.Location

const (
	sessionName      = "isu6f"
	sessionKeyTeamID = "team-id"
)

func parseTemplateAsset(t *template.Template, name string) error {
	content, err := Asset(name)
	if err != nil {
		return err
	}

	_, err = t.Parse(string(content))
	return err
}

func initWeb() error {
	var err error

	dsn := *databaseDSN + "?charset=utf8mb4&parseTime=true&loc=Asia%2FTokyo&time_zone='Asia%2FTokyo'"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return errors.Wrapf(err, "sql.Open %q", dsn)
	}

	err = db.Ping()
	if err != nil {
		return errors.Wrapf(err, "db.Ping %q", dsn)
	}

	const templatesRoot = "views/"

	for _, file := range []string{
		"index.tmpl", "login.tmpl", "debug-queue.tmpl", "debug-leaderboard.tmpl", "debug-proxies.tmpl", "messages.tmpl",
	} {
		t := template.New(file).Funcs(template.FuncMap{
			"contestEnded": func() bool {
				return getContestStatus() == contestStatusEnded
			},
			"plusOne": func(i int) int {
				return i + 1
			},
		})

		if err := parseTemplateAsset(t, templatesRoot+"layout.tmpl"); err != nil {
			return err
		}

		if err := parseTemplateAsset(t, templatesRoot+file); err != nil {
			return err
		}

		templates[file] = t
	}

	locJST, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return err
	}

	sessionStore = sessions.NewCookieStore([]byte(":beers:"))

	return nil
}

type Team struct {
	ID           int
	Name         string
	IPAddr       string
	InstanceName string
}

func loadTeam(id uint64) (*Team, error) {
	var team Team

	row := db.QueryRow("SELECT id,name,IFNULL(ip_address, ''),IFNULL(instance_name, '') FROM teams WHERE id = ?", id)
	err := row.Scan(&team.ID, &team.Name, &team.IPAddr, &team.InstanceName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &team, err
}

func loadTeamFromSession(req *http.Request) (*Team, error) {
	if *debugMode {
		c, _ := req.Cookie("debug_team")
		if c != nil {
			n, _ := strconv.ParseUint(c.Value, 10, 0)
			if n != 0 {
				return loadTeam(n)
			}
		}
	}

	sess, err := sessionStore.New(req, sessionName)
	if err != nil {
		if cerr, ok := err.(securecookie.Error); ok && cerr.IsDecode() {
			// 違う session secret でアクセスしにくるとこれなので無視
		} else {
			return nil, errors.Wrap(err, "sessionStore.New()")
		}
	}

	v, ok := sess.Values[sessionKeyTeamID]
	if !ok {
		return nil, nil
	}

	teamID, ok := v.(uint64)
	if !ok {
		return nil, nil
	}

	team, err := loadTeam(teamID)
	return team, errors.Wrapf(err, "loadTeam(id=%#v)", teamID)
}

type viewParamsLayout struct {
	Team *Team
}

func serveIndex(w http.ResponseWriter, req *http.Request) error {
	return serveIndexWithMessage(w, req, "")
}

func serveIndexWithMessage(w http.ResponseWriter, req *http.Request, message string) error {
	if getContestStatus() == contestStatusEnded {
		http.Error(w, "Today's final has ended", http.StatusForbidden)
		return nil
	}

	team, err := loadTeamFromSession(req)
	if err != nil {
		return err
	}

	teamID := 0
	if team != nil {
		teamID = team.ID
	}

	plotLines, latestScores, err := getResults(db, teamID, 10, getRankingFixedAt())
	if err != nil {
		return err
	}

	teamResults, err := getTeamResults(db, teamID)
	if err != nil {
		return err
	}

	// キューをゲット
	jobs, err := getQueuedJobs(db)
	if err != nil {
		return err
	}

	messages, err := getMessages()
	if err != nil {
		return err
	}
	if message != "" {
		messages = append(messages, Message{Message: message, Kind: "danger"})
	}

	return templates["index.tmpl"].Execute(
		w, struct {
			viewParamsLayout
			PlotLines      []PlotLine
			LatestScores   []LatestScore
			IsRankingFixed bool
			TeamResults    []TeamResult
			Jobs           []QueuedJob
			Messages       []Message
		}{
			viewParamsLayout{team},
			plotLines,
			latestScores,
			getRankingFixedAt().Before(time.Now()),
			teamResults,
			jobs,
			messages,
		},
	)
}

type viewParamsLogin struct {
	viewParamsLayout
	ErrorMessage string
}

func serveLogin(w http.ResponseWriter, req *http.Request) error {
	if getContestStatus() == contestStatusEnded {
		http.Error(w, "Today's final has ended", http.StatusForbidden)
		return nil
	}

	team, err := loadTeamFromSession(req)
	if err != nil {
		return err
	}

	if req.Method == "GET" {
		return templates["login.tmpl"].Execute(
			w, viewParamsLogin{viewParamsLayout{team}, ""})
	}

	var (
		id       = req.FormValue("team_id")
		password = req.FormValue("password")
	)

	var teamID uint64
	row := db.QueryRow("SELECT id FROM teams WHERE id = ? AND password = ? LIMIT 1", id, password)
	err = row.Scan(&teamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return templates["login.tmpl"].Execute(
				w, viewParamsLogin{viewParamsLayout{team}, "Wrong id/password pair"})
		} else {
			return err
		}
	}

	sess, err := sessionStore.New(req, sessionName)
	if err != nil {
		if cerr, ok := err.(securecookie.Error); ok && cerr.IsDecode() {
			// 違う session secret でアクセスしにくるとこれなので無視
		} else {
			return errors.Wrap(err, "sessionStore.New()")
		}
	}

	sess.Values[sessionKeyTeamID] = teamID

	err = sess.Save(req, w)
	if err != nil {
		return err
	}

	http.Redirect(w, req, "/", 302)

	return nil
}

type httpError interface {
	httpStatus() int
	error
}

type errHTTP int

func (s errHTTP) Error() string   { return http.StatusText(int(s)) }
func (s errHTTP) httpStatus() int { return int(s) }

type errHTTPMessage struct {
	status  int
	message string
}

func (m errHTTPMessage) Error() string   { return m.message }
func (m errHTTPMessage) httpStatus() int { return m.status }

func serveStatic(w http.ResponseWriter, req *http.Request) error {
	path := req.URL.Path[1:]
	content, err := Asset(path)
	if err != nil {
		return errHTTP(http.StatusNotFound)
	}
	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}
	w.Write(content)

	return nil
}

func serveUpdateTeam(w http.ResponseWriter, req *http.Request) error {
	if getContestStatus() == contestStatusEnded {
		http.Error(w, "Today's final has ended", http.StatusForbidden)
		return nil
	}

	if req.Method != http.MethodPost {
		return errHTTP(http.StatusMethodNotAllowed)
	}

	team, err := loadTeamFromSession(req)
	if err != nil {
		return err
	}
	if team == nil {
		return errHTTP(http.StatusForbidden)
	}

	instanceName := req.FormValue("instance_name")
	ipAddress := req.FormValue("ip_address")

	if ipAddress != "" {
		ip := net.ParseIP(ipAddress)
		if ip == nil || ip.To4() == nil {
			return errHTTP(http.StatusBadRequest)
		}
	}

	_, err = db.Exec("UPDATE teams SET instance_name = ?, ip_address = ? WHERE id = ?", instanceName, ipAddress, team.ID)
	if err != nil {
		return err
	}

	// proxyにチームのIPアドレスを通知する
	err = exec.Command(`/usr/local/bin/consul`, `event`, `-name`, `nginx_reload`, `-node`, `proxy`).Run()
	if err != nil {
		log.Printf("consul: %v", err)
		return err
	}

	// IPアドレスの反映に時間がかかることを考えてこのへんで3秒待つ
	time.Sleep(3 * time.Second)

	http.Redirect(w, req, "/", http.StatusFound)
	return nil
}

func serveDebugLeaderboard(w http.ResponseWriter, req *http.Request) error {
	plotLines, latestScores, err := getResults(db, 0, 26, time.Now()) // ここは常に最新のを使う
	if err != nil {
		return err
	}

	type viewParamsDebugLeaderboard struct {
		viewParamsLayout
		PlotLines    []PlotLine
		LatestScores []LatestScore
	}

	return templates["debug-leaderboard.tmpl"].Execute(
		w, viewParamsDebugLeaderboard{
			viewParamsLayout{nil},
			plotLines,
			latestScores,
		},
	)
}
