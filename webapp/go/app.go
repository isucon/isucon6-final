package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
)

var (
	db *sqlx.DB
)

type token struct {
	ID        int64
	CSRFToken string
	CreatedAt time.Time
}

type point struct {
	ID       int64   `json:"id" db:"id"`
	StrokeID int64   `json:"stroke_id" db:"stroke_id"`
	X        float64 `json:"x" db:"x"`
	Y        float64 `json:"y" db:"y"`
}

type stroke struct {
	ID        int64     `json:"id" db:"id"`
	RoomID    int64     `json:"room_id" db:"room_id"`
	Width     int       `json:"width" db:"width"`
	Red       int       `json:"red" db:"red"`
	Green     int       `json:"green" db:"green"`
	Blue      int       `json:"blue" db:"blue"`
	Alpha     float64   `json:"alpha" db:"alpha"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Points    []point   `json:"points" db:"points"`
}

type room struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	CanvasWidth  int       `json:"canvas_width"`
	CanvasHeight int       `json:"canvas_height"`
	CreatedAt    time.Time `json:"created_at"`
	Strokes      []stroke  `json:"strokes"`
	StrokeCount  int       `json:"stroke_count"`
}

func checkToken(csrfToken string) bool {
	if csrfToken == "" {
		return false
	}

	query := "SELECT `id`, `csrf_token`, `created_at` FROM `tokens`"
	query += " WHERE `csrf_token` = :csrf_token AND `created_at` > CURRENT_TIMESTAMP(6) - INTERVAL 1 DAY"
	t := token{}
	err := db.QueryRow(query, csrfToken).Scan(&t.ID, &t.CSRFToken, &t.CreatedAt)

	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	}

	return true
}

func getStrokePoints(strokeID int) ([]point, error) {
	query := "SELECT `id`, `stroke_id`, `x`, `y` FROM `points` WHERE `stroke_id` = ? ORDER BY `id` ASC"
	ps := []point{}
	err := db.Select(&ps, query, strokeID)
	if err != nil {
		return nil, err
	}
	return ps, nil
}

func getStrokes(roomID int, greaterThanID int) ([]stroke, error) {
	query := "SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`"
	query += " WHERE `room_id` = ? AND `id` > ? ORDER BY `id` ASC"
	strokes := []stroke{}
	err := db.Select(&strokes, query, roomID, greaterThanID)
	if err != nil {
		return nil, err
	}
	return strokes, nil
}

func getRoom(roomID int) {
	query := "SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at` FROM `rooms` WHERE `id` = ?"
	r := room{}
	err := db.QueryRow(query, roomID).Scan(&r.ID, &r.Name, &r.CanvasWidth, &r.CanvasHeight, &r.CreatedAt)
	if err != nil {

	}
}

func outputErrorMsg(w http.ResponseWriter, msg string) {
	b, jerr := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: msg})

	if jerr != nil {
		log.Fatal(jerr)
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Write(b)
}

func postApiCsrfToken(w http.ResponseWriter, r *http.Request) {
	sql := "INSERT INTO `tokens` (`csrf_token`) VALUES"
	sql += " (SHA2(RAND(), 256))"

	result, derr := db.Exec(sql)
	if derr != nil {
		log.Fatal(derr.Error())
	}

	id, lerr := result.LastInsertId()
	if lerr != nil {
		log.Fatal(lerr)
	}

	t := token{}
	sql = "SELECT `id`, `csrf_token`, `created_at` FROM `tokens` WHERE id = ?"
	err := db.QueryRow(sql, id).Scan(&t.ID, &t.CSRFToken, &t.CreatedAt)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	b, jerr := json.Marshal(struct {
		Token string `json:"token"`
	}{Token: t.CSRFToken})

	if jerr != nil {
		log.Fatal(jerr)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getApiRooms(w http.ResponseWriter, r *http.Request) {
	sql := "SELECT `room_id`, MAX(`id`) AS `max_id` FROM `strokes`"
	sql += " GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100"

	rows, derr := db.Query(sql)
	if derr != nil {
		log.Fatal(derr)
	}
	defer rows.Close()

	rooms := []room{}

	for rows.Next() {
		r := room{}
		rerr := rows.Scan(&r.ID, &r.Name, &r.CanvasWidth, &r.CanvasHeight, &r.CreatedAt)
		if rerr != nil {
			log.Fatal(rerr)
		}
		rooms = append(rooms, r)
	}

	for i, r := range rooms {
		sql = "SELECT COUNT(*) AS stroke_count FROM `strokes` WHERE `room_id` = ?"
		var count int
		err := db.QueryRow(sql, r.ID).Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
		rooms[i].StrokeCount = count
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	b, jerr := json.Marshal(rooms)

	if jerr != nil {
		log.Fatal(jerr)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func postApiRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	if !checkToken(r.Header.Get("x-csrf-token")) {
		outputErrorMsg(w, "トークンエラー。ページを再読み込みしてください。")
		return
	}

	if r.FormValue("name") == "" || r.FormValue("canvas_width") == "" || r.FormValue("canvas_height") == "" {
		outputErrorMsg(w, "トークンエラー。ページを再読み込みしてください。")
		return
	}

	query := "INSERT INTO `rooms` (`name`, `canvas_width`, `canvas_height`)"
	query += " VALUES (?, ?, ?)"
	result, derr := db.Exec(query, r.FormValue("name") == "", r.FormValue("canvas_width"), r.FormValue("canvas_height"))
	if derr != nil {
		log.Fatal(derr)
	}
	id, lerr := result.LastInsertId()
	if lerr != nil {
		log.Fatal(lerr)
	}

	rm := room{}
	query = "SELECT * FROM `rooms` WHERE `id` = ?"
	err := db.QueryRow(query, id).Scan(&rm.ID, &rm.Name, &rm.CanvasWidth, &rm.CanvasHeight, &rm.CreatedAt)
	if err != nil {
		log.Fatal(err)
	}

	b, jerr := json.Marshal(rm)
	if jerr != nil {
		log.Fatal(jerr)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getApiRoomsID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	idStr := pat.Param(ctx, "id")
	id, aerr := strconv.Atoi(idStr)
	if aerr != nil {
		outputErrorMsg(w, "この部屋は存在しません。")
		return
	}

	rm := room{}
	query := "SELECT * FROM `rooms` WHERE `id` = ?"
	derr := db.QueryRow(query, id).Scan(&rm.ID, &rm.Name, &rm.CanvasWidth, &rm.CanvasHeight, &rm.CreatedAt)
	if derr == sql.ErrNoRows {
		outputErrorMsg(w, "この部屋は存在しません。")
		return
	} else if derr != nil {
		log.Fatal(derr)
	}

	query = "SELECT * FROM `strokes` WHERE `room_id` = ? ORDER BY `id` ASC"
	rows, qerr := db.Query(query, id)
	if qerr != nil {
		log.Fatal(qerr)
	}
	defer rows.Close()
	strokes := []stroke{}
	for rows.Next() {
		s := stroke{}
		rerr := rows.Scan(&s.ID, &s.RoomID, &s.Width, &s.Red, &s.Green, &s.Blue, &s.Alpha, &s.CreatedAt)
		if rerr != nil {
			log.Fatal(rerr)
		}
		strokes = append(strokes, s)
	}

	for i, s := range strokes {
		query := "SELECT * FROM `points` WHERE `stroke_id` = ? ORDER BY `id` ASC"
		rows, pderr := db.Query(query, s.ID)
		if pderr != nil {
			log.Fatal(pderr)
		}
		defer rows.Close()
		ps := []point{}
		for rows.Next() {
			p := point{}
			rows.Scan(&p.ID, &p.StrokeID, &p.X, &p.Y)
			ps = append(ps, p)
		}
		strokes[i].Points = ps
	}

	rm.Strokes = strokes

	b, jerr := json.Marshal(struct {
		Room room `json:"room"`
	}{Room: rm})

	if jerr != nil {
		log.Fatal(jerr)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func main() {
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Failed to read DB port number from an environment variable MYSQL_PORT.\nError: %s", err.Error())
	}
	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("MYSQL_PASS")
	dbname := "isuchannel"

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user,
		password,
		host,
		port,
		dbname,
	)

	db, err = sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s.", err.Error())
	}
	defer db.Close()

	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/api/csrf_token"), postApiCsrfToken)
	mux.HandleFunc(pat.Get("/api/rooms"), getApiRooms)
	mux.HandleFunc(pat.Post("/api/rooms"), postApiRooms)
	mux.HandleFuncC(pat.Get("/api/rooms/:id"), getApiRoomsID)

	http.ListenAndServe("localhost:8000", mux)
}
