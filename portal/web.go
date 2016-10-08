package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/catatsuy/isucon6-final/portal/score"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

var (
	databaseDSN = flag.String("database-dsn", "root:@/isu6fportal_day0", "database `dsn`")
	debugMode   = flag.Bool("debug", false, "enable debug mode")
)

var db *sql.DB
var day int
var templates = map[string]*template.Template{}
var sessionStore sessions.Store
var locJST *time.Location

const (
	sessionName      = "isu6f"
	sessionKeyTeamID = "team-id"
)

const (
	rankingPickLatest = 20
	rankingPickBest   = 20
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

	dsn := *databaseDSN + "?parseTime=true&loc=Asia%2FTokyo&time_zone='Asia%2FTokyo'"
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
		"index.tmpl", "login.tmpl", "debug-queue.tmpl", "debug-leaderboard.tmpl", "debug-proxies.tmpl",
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

	// ホントはJSONだけど整数値しか入ってないことを知ってるのでショートカット
	err = db.QueryRow("SELECT CONVERT(json,SIGNED) FROM setting WHERE name = 'day'").Scan(&day)
	if err != nil {
		return errors.Wrap(err, "SELECT CONVERT(json,SIGNED) FROM setting WHERE name = 'day'")
	}

	// 日によってDBを分けるので、万一 teams.id が被ってたら
	// 前日のセッションでログイン状態になってしまう
	sessionStore = sessions.NewCookieStore([]byte(fmt.Sprintf(":beers:%d", day)))

	return nil
}

type Team struct {
	ID           int
	Name         string
	IPAddr       string
	InstanceName string
}

type Score struct {
	Team   Team
	Latest int64
	Best   int64
	At     time.Time
}

type PlotLine struct {
	Name string         `json:"name"`
	Data map[string]int `json:"data"`
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

type byLatest []*Score

func (ss byLatest) Len() int           { return len(ss) }
func (ss byLatest) Less(i, j int) bool { return ss[i].Latest > ss[j].Latest }
func (ss byLatest) Swap(i, j int)      { ss[i], ss[j] = ss[j], ss[i] }

type byBest []*Score

func (ss byBest) Len() int           { return len(ss) }
func (ss byBest) Less(i, j int) bool { return ss[i].Best > ss[j].Best }
func (ss byBest) Swap(i, j int)      { ss[i], ss[j] = ss[j], ss[i] }

type queuedJob struct {
	TeamID int
	Status string
}

type viewParamsLayout struct {
	Team *Team
	Day  int
}

type latestResult struct {
	Output *score.Output
	At     time.Time
	Score  *Score
}

type viewParamsIndex struct {
	viewParamsLayout
	Ranking        []*Score
	RankingIsFixed bool
	PlotData       []PlotLine
	Jobs           []queuedJob
	LatestResult   latestResult
	Message        string
}

type viewParamsLogin struct {
	viewParamsLayout
	ErrorMessage string
}

func serveIndex(w http.ResponseWriter, req *http.Request) error {
	return serveIndexWithMessage(w, req, "")
}

func buildLeaderboard(team *Team) ([]*Score, *Score, bool, error) {
	// team_scores_snapshot にデータが入ってたらそっちを使う
	// ラスト1時間でランキングの更新を止めるための措置
	// データは手動でいれる :P
	ranking, myScore, err := buildLeaderboardFromTable(team, true)
	if err == nil && ranking != nil && len(ranking) > 0 {
		return ranking, myScore, true, nil
	} else if err != nil {
		log.Printf("buildLeaderboardFromTable: %v", err)
	}

	ranking, myScore, err = buildLeaderboardFromTable(team, false)
	return ranking, myScore, false, nil
}

func buildLeaderboardFromTable(team *Team, useSnapshot bool) ([]*Score, *Score, error) {
	// ランキングを作る。
	// 現在のスコアのトップ rankingPickLatest と最高スコアのトップ rankingPickBest と自チーム
	table := "team_scores"
	if useSnapshot {
		table = "team_scores_snapshot"

	}

	var (
		allScores     = []*Score{}
		scoreByTeamID = map[int]*Score{}
	)

	rows, err := db.Query(`
		SELECT teams.id,teams.name,team_scores.latest_score,team_scores.best_score,team_scores.updated_at
		FROM ` + table + ` AS team_scores
		  JOIN teams
		  ON team_scores.team_id = teams.id
		WHERE teams.category <> 'official'
	`)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var score Score
		err := rows.Scan(&score.Team.ID, &score.Team.Name, &score.Latest, &score.Best, &score.At)
		if err != nil {
			return nil, nil, err
		}
		allScores = append(allScores, &score)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// まず自チームのスコアのみを追加。17時を超えていても自チームだけは常に最新にする
	if useSnapshot {
		if len(allScores) == 0 {
			// スナップショットテーブルが空のときはさっさと処理を終わる
			return []*Score{}, nil, nil
		}

		// スナップショットの場合も自分のスコアだけは最新を使う
		if team.ID == 9999 {
			// 運営チームがランキングに入ってたら混乱するので入れない
		} else {
			var score Score
			err := db.QueryRow(`
				SELECT teams.id,teams.name,team_scores.latest_score,team_scores.best_score,team_scores.updated_at
				FROM team_scores
				  JOIN teams
				  ON team_scores.team_id = teams.id
				WHERE teams.id = ?
			`, team.ID).Scan(&score.Team.ID, &score.Team.Name, &score.Latest, &score.Best, &score.At)
			if err != nil {
				return nil, nil, err
			}

			scoreByTeamID[score.Team.ID] = &score
		}
	} else {
		for _, score := range allScores {
			if score.Team.ID == team.ID {
				scoreByTeamID[score.Team.ID] = score
			}
		}
	}

	// 次に自チーム以外のスコアを追加
	sort.Sort(byLatest(allScores))
	for _, s := range allScores {
		//if i >= rankingPickLatest {
		//	break
		//}
		if s.Team.ID != team.ID {
			scoreByTeamID[s.Team.ID] = s
		}
	}

	// sort.Sort(byBest(allScores))
	// for i, s := range allScores {
	// 	if i >= rankingPickBest {
	// 		break
	// 	}
	// 	scoreByTeamID[s.Team.ID] = s
	// }

	ranking := make([]*Score, 0, len(scoreByTeamID))
	for _, s := range scoreByTeamID {
		ranking = append(ranking, s)
	}

	// 最後に、最新のスコアでソート
	sort.Sort(byLatest(ranking))

	return ranking, scoreByTeamID[team.ID], nil
}

func buildPlotLine(score *Score) (*PlotLine, error) {
	plotLine := PlotLine{}
	plotLine.Name = score.Team.Name
	plotLine.Data = make(map[string]int)

	// 17時になったらsnapshotを使うが、自チームのscoreには最新が入ってる
	// 自チーム以外の最新scoreがプロットに出てしまわないように、created_atでも絞り込む
	rows, err := db.Query(`
		SELECT score, created_at FROM scores
		WHERE team_id = ? AND created_at <= ?
	  ORDER BY id ASC
	`, score.Team.ID, score.At)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			T time.Time
			S int
		)

		err := rows.Scan(&S, &T)
		if err != nil {
			return nil, err
		}

		plotLine.Data[T.Format("2006-01-02T15:04:05")] = S
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &plotLine, nil
}

func buildPlotData(ranking []*Score) ([]PlotLine, error) {
	plotData := make([]PlotLine, 0)

	if len(ranking) == 0 {
		return plotData, nil
	}

	for _, score := range ranking {
		plotLine, err := buildPlotLine(score)
		if err != nil {
			return nil, err
		}
		plotData = append(plotData, *plotLine)
	}

	return plotData, nil
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

	if team == nil {
		http.Redirect(w, req, "/login", http.StatusFound)
		return nil
	}

	ranking, myScore, rankingIsFixed, err := buildLeaderboard(team)
	if err != nil {
		return err
	}

	plotData, err := buildPlotData(ranking)
	if err != nil {
		return err
	}

	// キューをゲット
	jobs := []queuedJob{}
	if getContestStatus() == contestStatusStarted {
		rows, err := db.Query(`
			SELECT team_id, status
			FROM queues
			WHERE status IN ('waiting', 'running')
			  AND team_id <> 9999
			ORDER BY created_at ASC
		`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var job queuedJob
			err := rows.Scan(&job.TeamID, &job.Status)
			if err != nil {
				rows.Close()
				return err
			}
			jobs = append(jobs, job)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
	}

	// 自分チームの最新状況を取得
	var (
		latestScore     *score.Output
		latestScoreAt   time.Time
		latestScoreJSON string
	)
	err = db.QueryRow(`
		SELECT IFNULL(result_json, ''),updated_at FROM queues
		WHERE team_id = ?
		  AND status = 'done'
		ORDER BY updated_at DESC
		LIMIT 1
	`, team.ID).Scan(&latestScoreJSON, &latestScoreAt)
	switch err {
	case sql.ErrNoRows:
		// nop
	case nil:
		var res score.Output
		err := json.Unmarshal([]byte(latestScoreJSON), &res)
		if err != nil {
			return err
		}
		latestScore = &res
	default:
		return err
	}

	return templates["index.tmpl"].Execute(
		w, viewParamsIndex{
			viewParamsLayout{team, day},
			ranking,
			rankingIsFixed,
			plotData,
			jobs,
			latestResult{
				latestScore,
				latestScoreAt,
				myScore,
			},
			message,
		},
	)
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
		return templates["login.tmpl"].Execute(w, viewParamsLogin{viewParamsLayout{team, day}, ""})
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
			return templates["login.tmpl"].Execute(w, viewParamsLogin{viewParamsLayout{team, day}, "Wrong id/password pair"})
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

	// TODO: proxyにチームのIPアドレスを通知する

	_, err = db.Exec("UPDATE teams SET instance_name = ?, ip_address = ? WHERE id = ?", instanceName, ipAddress, team.ID)
	if err != nil {
		return err
	}

	// TODO: IPアドレスの反映に時間がかかることを考えてこのへんで3秒程度待つか？

	http.Redirect(w, req, "/", http.StatusFound)
	return nil
}

func serveDebugLeaderboard(w http.ResponseWriter, req *http.Request) error {
	// ここは常に最新のを使う
	ranking, _, err := buildLeaderboardFromTable(&Team{}, false)
	if err != nil {
		return err
	}

	plotData, err := buildPlotData(ranking)
	if err != nil {
		return err
	}

	type viewParamsDebugLeaderboard struct {
		viewParamsLayout
		Ranking  []*Score
		PlotData []PlotLine
	}

	return templates["debug-leaderboard.tmpl"].Execute(
		w, viewParamsDebugLeaderboard{
			viewParamsLayout{nil, day},
			ranking,
			plotData,
		},
	)
}
