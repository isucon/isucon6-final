package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	dbx *sqlx.DB
)

type Token struct {
	ID        int64     `db:"id"`
	CSRFToken string    `db:"csrf_token"`
	CreatedAt time.Time `db:"created_at"`
}

type Point struct {
	ID       int64   `json:"id" db:"id"`
	StrokeID int64   `json:"stroke_id" db:"stroke_id"`
	X        float64 `json:"x" db:"x"`
	Y        float64 `json:"y" db:"y"`
}

type Stroke struct {
	ID        int64     `json:"id" db:"id"`
	RoomID    int64     `json:"room_id" db:"room_id"`
	Width     int       `json:"width" db:"width"`
	Red       int       `json:"red" db:"red"`
	Green     int       `json:"green" db:"green"`
	Blue      int       `json:"blue" db:"blue"`
	Alpha     float64   `json:"alpha" db:"alpha"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Points    []Point   `json:"points" db:"points"`
}

type Room struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	CanvasWidth  int       `json:"canvas_width" db:"canvas_width"`
	CanvasHeight int       `json:"canvas_height" db:"canvas_height"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	Strokes      []Stroke  `json:"strokes"`
	StrokeCount  int       `json:"stroke_count"`
	WatcherCount int       `json:"watcher_count"`
}

func printAndFlush(w http.ResponseWriter, content string) {
	fmt.Fprint(w, content)

	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	f.Flush()
}

func checkToken(csrfToken string) (*Token, error) {
	if csrfToken == "" {
		return nil, nil
	}

	query := "SELECT `id`, `csrf_token`, `created_at` FROM `tokens`"
	query += " WHERE `csrf_token` = ? AND `created_at` > CURRENT_TIMESTAMP(6) - INTERVAL 1 DAY"

	t := &Token{}
	err := dbx.Get(t, query, csrfToken)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return t, nil
}

func getStrokePoints(strokeID int64) ([]Point, error) {
	query := "SELECT `id`, `stroke_id`, `x`, `y` FROM `points` WHERE `stroke_id` = ? ORDER BY `id` ASC"
	ps := []Point{}
	err := dbx.Select(&ps, query, strokeID)
	if err != nil {
		return nil, err
	}
	return ps, nil
}

func getStrokes(roomID int64, greaterThanID int64) ([]Stroke, error) {
	query := "SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`"
	query += " WHERE `room_id` = ? AND `id` > ? ORDER BY `id` ASC"
	strokes := []Stroke{}
	err := dbx.Select(&strokes, query, roomID, greaterThanID)
	if err != nil {
		return nil, err
	}
	return strokes, nil
}

func getRoom(roomID int64) (*Room, error) {
	query := "SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at` FROM `rooms` WHERE `id` = ?"
	r := &Room{}
	err := dbx.Get(r, query, roomID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return r, nil
}

func getWatcherCount(roomID int64) (int, error) {
	query := "SELECT COUNT(*) AS `watcher_count` FROM `room_watchers`"
	query += " WHERE `room_id` = ? AND `updated_at` > CURRENT_TIMESTAMP(6) - INTERVAL 3 SECOND"

	var watcherCount int
	err := dbx.QueryRow(query, roomID).Scan(&watcherCount)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return watcherCount, nil
}

func updateRoomWatcher(roomID int64, tokenID int64) error {
	query := "INSERT INTO `room_watchers` (`room_id`, `token_id`) VALUES (?, ?)"
	query += " ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(6)"

	_, err := dbx.Exec(query, roomID, tokenID)
	return err
}

func outputErrorMsg(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	b, _ := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: msg})

	w.WriteHeader(status)
	w.Write(b)
}

func outputError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	b, _ := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: "InternalServerError"})

	w.Write(b)
	fmt.Fprintln(os.Stderr, err.Error())
}

func postApiCsrfToken(w http.ResponseWriter, r *http.Request) {
	query := "INSERT INTO `tokens` (`csrf_token`) VALUES"
	query += " (SHA2(RAND(), 256))"

	result, derr := dbx.Exec(query)
	if derr != nil {
		outputError(w, derr)
		return
	}

	id, lerr := result.LastInsertId()
	if lerr != nil {
		outputError(w, lerr)
		return
	}

	t := Token{}
	query = "SELECT `id`, `csrf_token`, `created_at` FROM `tokens` WHERE id = ?"
	err := dbx.Get(&t, query, id)
	if err != nil {
		outputError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	b, _ := json.Marshal(struct {
		Token string `json:"token"`
	}{Token: t.CSRFToken})

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getApiRooms(w http.ResponseWriter, r *http.Request) {
	query := "SELECT `room_id`, MAX(`id`) AS `max_id` FROM `strokes`"
	query += " GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100"

	type result struct {
		RoomID int64 `db:"room_id"`
		MaxID  int64 `db:"max_id"`
	}

	results := []result{}

	derr := dbx.Select(&results, query)
	if derr != nil {
		outputError(w, derr)
		return
	}

	rooms := []*Room{}

	for _, r := range results {
		room, rerr := getRoom(r.RoomID)
		if rerr != nil {
			outputError(w, rerr)
			return
		}
		s, serr := getStrokes(room.ID, 0)
		if serr != nil {
			outputError(w, serr)
			return
		}
		room.StrokeCount = len(s)
		rooms = append(rooms, room)
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	b, _ := json.Marshal(rooms)

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func postApiRooms(w http.ResponseWriter, r *http.Request) {
	t, terr := checkToken(r.Header.Get("x-csrf-token"))

	if terr != nil {
		outputError(w, terr)
		return
	}

	if t == nil {
		outputErrorMsg(w, http.StatusBadRequest, "トークンエラー。ページを再読み込みしてください。")
		return
	}

	body, _ := ioutil.ReadAll(r.Body)
	var postedRoom Room
	jerr := json.Unmarshal(body, &postedRoom)
	if jerr != nil {
		outputError(w, jerr)
		return
	}

	if postedRoom.Name == "" || postedRoom.CanvasWidth == 0 || postedRoom.CanvasHeight == 0 {
		outputErrorMsg(w, http.StatusBadRequest, "リクエストが正しくありません。")
		return
	}

	tx := dbx.MustBegin()
	query := "INSERT INTO `rooms` (`name`, `canvas_width`, `canvas_height`)"
	query += " VALUES (?, ?, ?)"

	result := tx.MustExec(query, postedRoom.Name, postedRoom.CanvasWidth, postedRoom.CanvasHeight)
	roomID, lerr := result.LastInsertId()
	if lerr != nil {
		outputError(w, lerr)
		return
	}

	query = "INSERT INTO `room_owners` (`room_id`, `token_id`) VALUES (?, ?)"
	tx.MustExec(query, roomID, t.ID)

	cerr := tx.Commit()
	if cerr != nil {
		tx.Rollback()
		outputErrorMsg(w, http.StatusInternalServerError, "エラーが発生しました。")
		return
	}

	room, rerr := getRoom(roomID)
	if rerr != nil {
		outputError(w, rerr)
		return
	}

	b, _ := json.Marshal(struct {
		Room *Room `json:"room"`
	}{Room: room})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getApiRoomsID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	idStr := pat.Param(ctx, "id")
	id, aerr := strconv.Atoi(idStr)
	if aerr != nil {
		outputErrorMsg(w, http.StatusNotFound, "この部屋は存在しません。")
		return
	}

	room, rerr := getRoom(int64(id))
	if rerr != nil {
		outputError(w, rerr)
		return
	}

	if room == nil {
		outputErrorMsg(w, http.StatusNotFound, "この部屋は存在しません。")
		return
	}

	strokes, _ := getStrokes(room.ID, 0)

	for i, s := range strokes {
		p, _ := getStrokePoints(s.ID)
		strokes[i].Points = p
	}

	room.Strokes = strokes
	var wcerr error
	room.WatcherCount, wcerr = getWatcherCount(room.ID)
	if wcerr != nil {
		outputError(w, wcerr)
		return
	}

	b, _ := json.Marshal(struct {
		Room *Room `json:"room"`
	}{Room: room})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getApiStrokesRoomsID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")

	idStr := pat.Param(ctx, "id")
	id, aerr := strconv.Atoi(idStr)
	if aerr != nil {
		return
	}

	t, terr := checkToken(r.URL.Query().Get("csrf_token"))

	if terr != nil {
		outputError(w, terr)
		return
	}
	if t == nil {
		printAndFlush(w, "event:bad_request\n"+"data:トークンエラー。ページを再読み込みしてください。\n\n")
		return
	}

	room, rerr := getRoom(int64(id))
	if rerr != nil {
		outputError(w, rerr)
		return
	}
	if room == nil {
		printAndFlush(w, "event:bad_request\n"+"data:この部屋は存在しません\n\n")
		return
	}

	urwerr := updateRoomWatcher(room.ID, t.ID)
	if urwerr != nil {
		outputError(w, urwerr)
		return
	}

	var wcerr error
	watcherCount, wcerr := getWatcherCount(room.ID)
	if wcerr != nil {
		outputError(w, wcerr)
		return
	}

	printAndFlush(w, "retry:500\n\n"+"event:watcher_count\n"+"data:"+strconv.Itoa(watcherCount)+"\n\n")

	var lastStrokeID int64
	lastEventIDStr := r.Header.Get("Last-Event-ID")
	if lastEventIDStr != "" {
		lastEventID, _ := strconv.Atoi(lastEventIDStr)
		lastStrokeID = int64(lastEventID)
	}

	loop := 6
	for loop > 0 {
		loop--
		time.Sleep(500 * 1000 * time.Microsecond)

		strokes, _ := getStrokes(room.ID, int64(lastStrokeID))

		for _, s := range strokes {
			s.Points, _ = getStrokePoints(s.ID)
			d, _ := json.Marshal(s)
			printAndFlush(w, "id:"+strconv.FormatInt(s.ID, 10)+"\n\n"+"event:stroke\n"+"data:"+string(d)+"\n\n")
			lastStrokeID = s.ID
		}

		urwerr = updateRoomWatcher(room.ID, t.ID)
		if urwerr != nil {
			outputError(w, urwerr)
			return
		}

		newWatcherCount, wcerr := getWatcherCount(room.ID)
		if wcerr != nil {
			outputError(w, wcerr)
			return
		}
		if newWatcherCount != watcherCount {
			printAndFlush(w, "event:watcher_count\n"+"data:"+strconv.Itoa(watcherCount)+"\n\n")
			watcherCount = newWatcherCount
		}
	}
}

func postApiStrokesRoomsID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	t, terr := checkToken(r.Header.Get("x-csrf-token"))

	if terr != nil {
		outputError(w, terr)
		return
	}
	if t == nil {
		outputErrorMsg(w, http.StatusBadRequest, "トークンエラー。ページを再読み込みしてください。")
		return
	}

	idStr := pat.Param(ctx, "id")
	id, aerr := strconv.Atoi(idStr)
	if aerr != nil {
		outputErrorMsg(w, http.StatusNotFound, "この部屋は存在しません。")
		return
	}

	room, rerr := getRoom(int64(id))
	if rerr != nil {
		outputError(w, rerr)
		return
	}
	if room == nil {
		outputErrorMsg(w, http.StatusNotFound, "この部屋は存在しません。")
		return
	}

	body, _ := ioutil.ReadAll(r.Body)
	var postedStroke Stroke
	jerr := json.Unmarshal(body, &postedStroke)
	if jerr != nil {
		outputError(w, jerr)
		return
	}

	if postedStroke.Width == 0 || len(postedStroke.Points) == 0 {
		outputErrorMsg(w, http.StatusBadRequest, "リクエストが正しくありません。")
		return
	}

	strokes, _ := getStrokes(room.ID, 0)
	strokeCount := len(strokes)
	if strokeCount > 1000 {
		outputErrorMsg(w, http.StatusBadRequest, "1000画を超えました。これ以上描くことはできません。")
		return
	}
	if strokeCount == 0 {
		query := "SELECT COUNT(*) AS cnt FROM `room_owners` WHERE `room_id` = ? AND `token_id` = ?"
		cnt := 0
		err := dbx.QueryRow(query, room.ID, t.ID).Scan(&cnt)
		if err != nil {
			outputError(w, err)
		}
		if cnt == 0 {
			outputErrorMsg(w, http.StatusBadRequest, "他人の作成した部屋に1画目を描くことはできません")
			return
		}
	}

	tx := dbx.MustBegin()
	query := "INSERT INTO `strokes` (`room_id`, `width`, `red`, `green`, `blue`, `alpha`)"
	query += " VALUES(?, ?, ?, ?, ?, ?)"

	result := tx.MustExec(query,
		room.ID,
		postedStroke.Width,
		postedStroke.Red,
		postedStroke.Green,
		postedStroke.Blue,
		postedStroke.Alpha,
	)
	strokeID, lerr := result.LastInsertId()
	if lerr != nil {
		outputError(w, lerr)
		return
	}

	query = "INSERT INTO `points` (`stroke_id`, `x`, `y`) VALUES (?, ?, ?)"
	for _, p := range postedStroke.Points {
		tx.MustExec(query, strokeID, p.X, p.Y)
	}

	cerr := tx.Commit()
	if cerr != nil {
		tx.Rollback()
		outputErrorMsg(w, http.StatusInternalServerError, "エラーが発生しました。")
		return
	}

	query = "SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`"
	query += " WHERE `id` = ?"
	var s Stroke
	serr := dbx.Get(&s, query, strokeID)
	if serr != nil {
		outputError(w, serr)
		return
	}

	s.Points, _ = getStrokePoints(strokeID)

	b, _ := json.Marshal(struct {
		Stroke Stroke `json:"stroke"`
	}{Stroke: s})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getApiInitialize(w http.ResponseWriter, r *http.Request) {
	queries := []string{
		"DELETE FROM `points` WHERE `id` > 1443000",
		"DELETE FROM `strokes` WHERE `id` > 41000",
		"DELETE FROM `rooms` WHERE `id` > 1000",
		"DELETE FROM `tokens` WHERE `id` > 0",
	}

	for _, query := range queries {
		dbx.Exec(query)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
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

	dbx, err = sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s.", err.Error())
	}
	defer dbx.Close()

	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/api/csrf_token"), postApiCsrfToken)
	mux.HandleFunc(pat.Get("/api/rooms"), getApiRooms)
	mux.HandleFunc(pat.Post("/api/rooms"), postApiRooms)
	mux.HandleFuncC(pat.Get("/api/rooms/:id"), getApiRoomsID)
	mux.HandleFuncC(pat.Get("/api/strokes/rooms/:id"), getApiStrokesRoomsID)
	mux.HandleFuncC(pat.Post("/api/strokes/rooms/:id"), postApiStrokesRoomsID)
	mux.HandleFunc(pat.Get("/api/initialize"), getApiInitialize)

	log.Fatal(http.ListenAndServe("localhost:8000", mux))
}
