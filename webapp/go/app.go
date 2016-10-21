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
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)

		b, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{Error: "Streaming unsupported!"})

		w.Write(b)
		fmt.Fprintln(os.Stderr, "Streaming unsupported!")
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
	// 空スライスを入れてJSONでnullを返さないように
	for i := range strokes {
		strokes[i].Points = []Point{}
	}
	return strokes, nil
}

func getRoom(roomID int64) (*Room, error) {
	query := "SELECT `id`, `name`, `canvas_width`, `canvas_height`, `created_at` FROM `rooms` WHERE `id` = ?"
	r := &Room{}
	err := dbx.Get(r, query, roomID)
	if err != nil {
		return nil, err
	}
	// 空スライスを入れてJSONでnullを返さないように
	r.Strokes = []Stroke{}
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

func postAPICsrfToken(w http.ResponseWriter, r *http.Request) {
	query := "INSERT INTO `tokens` (`csrf_token`) VALUES"
	query += " (SHA2(CONCAT(RAND(), UUID_SHORT()), 256))"

	result, err := dbx.Exec(query)
	if err != nil {
		outputError(w, err)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		outputError(w, err)
		return
	}

	t := Token{}
	query = "SELECT `id`, `csrf_token`, `created_at` FROM `tokens` WHERE id = ?"
	err = dbx.Get(&t, query, id)
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

func getAPIRooms(w http.ResponseWriter, r *http.Request) {
	query := "SELECT `room_id`, MAX(`id`) AS `max_id` FROM `strokes`"
	query += " GROUP BY `room_id` ORDER BY `max_id` DESC LIMIT 100"

	type result struct {
		RoomID int64 `db:"room_id"`
		MaxID  int64 `db:"max_id"`
	}

	results := []result{}

	err := dbx.Select(&results, query)
	if err != nil {
		outputError(w, err)
		return
	}

	rooms := []*Room{}

	for _, r := range results {
		room, err := getRoom(r.RoomID)
		if err != nil {
			outputError(w, err)
			return
		}
		s, err := getStrokes(room.ID, 0)
		if err != nil {
			outputError(w, err)
			return
		}
		room.StrokeCount = len(s)
		rooms = append(rooms, room)
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	b, _ := json.Marshal(struct {
		Rooms []*Room `json:"rooms"`
	}{Rooms: rooms})

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func postAPIRooms(w http.ResponseWriter, r *http.Request) {
	t, err := checkToken(r.Header.Get("x-csrf-token"))

	if err != nil {
		outputError(w, err)
		return
	}

	if t == nil {
		outputErrorMsg(w, http.StatusBadRequest, "トークンエラー。ページを再読み込みしてください。")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		outputError(w, err)
		return
	}

	postedRoom := Room{}
	err = json.Unmarshal(body, &postedRoom)
	if err != nil {
		outputError(w, err)
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
	roomID, err := result.LastInsertId()
	if err != nil {
		outputError(w, err)
		return
	}

	query = "INSERT INTO `room_owners` (`room_id`, `token_id`) VALUES (?, ?)"
	tx.MustExec(query, roomID, t.ID)

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		outputError(w, err)
		return
	}

	room, err := getRoom(roomID)
	if err != nil {
		outputError(w, err)
		return
	}

	b, _ := json.Marshal(struct {
		Room *Room `json:"room"`
	}{Room: room})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getAPIRoomsID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	idStr := pat.Param(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		outputErrorMsg(w, http.StatusNotFound, "この部屋は存在しません。")
		return
	}

	room, err := getRoom(id)
	if err != nil {
		if err == sql.ErrNoRows {
			outputErrorMsg(w, http.StatusNotFound, "この部屋は存在しません。")
		} else {
			outputError(w, err)
		}
		return
	}

	strokes, err := getStrokes(room.ID, 0)
	if err != nil {
		outputError(w, err)
		return
	}

	for i, s := range strokes {
		p, err := getStrokePoints(s.ID)
		if err != nil {
			outputError(w, err)
			return
		}
		strokes[i].Points = p
	}

	room.Strokes = strokes
	room.WatcherCount, err = getWatcherCount(room.ID)
	if err != nil {
		outputError(w, err)
		return
	}

	b, _ := json.Marshal(struct {
		Room *Room `json:"room"`
	}{Room: room})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getAPIStreamRoomsID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")

	idStr := pat.Param(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return
	}

	t, err := checkToken(r.URL.Query().Get("csrf_token"))

	if err != nil {
		outputError(w, err)
		return
	}
	if t == nil {
		printAndFlush(w, "event:bad_request\n"+"data:トークンエラー。ページを再読み込みしてください。\n\n")
		return
	}

	room, err := getRoom(id)
	if err != nil {
		if err == sql.ErrNoRows {
			printAndFlush(w, "event:bad_request\n"+"data:この部屋は存在しません\n\n")
		} else {
			outputError(w, err)
		}
		return
	}

	err = updateRoomWatcher(room.ID, t.ID)
	if err != nil {
		outputError(w, err)
		return
	}

	watcherCount, err := getWatcherCount(room.ID)
	if err != nil {
		outputError(w, err)
		return
	}

	printAndFlush(w, "retry:500\n\n"+"event:watcher_count\n"+"data:"+strconv.Itoa(watcherCount)+"\n\n")

	var lastStrokeID int64
	lastEventIDStr := r.Header.Get("Last-Event-ID")
	if lastEventIDStr != "" {
		lastEventID, err := strconv.ParseInt(lastEventIDStr, 10, 64)
		if err != nil {
			outputError(w, err)
			return
		}
		lastStrokeID = lastEventID
	}

	loop := 6
	for loop > 0 {
		loop--
		time.Sleep(500 * time.Millisecond)

		strokes, err := getStrokes(room.ID, int64(lastStrokeID))
		if err != nil {
			outputError(w, err)
			return
		}

		for _, s := range strokes {
			s.Points, err = getStrokePoints(s.ID)
			if err != nil {
				outputError(w, err)
				return
			}
			d, _ := json.Marshal(s)
			printAndFlush(w, "id:"+strconv.FormatInt(s.ID, 10)+"\n\n"+"event:stroke\n"+"data:"+string(d)+"\n\n")
			lastStrokeID = s.ID
		}

		err = updateRoomWatcher(room.ID, t.ID)
		if err != nil {
			outputError(w, err)
			return
		}

		newWatcherCount, err := getWatcherCount(room.ID)
		if err != nil {
			outputError(w, err)
			return
		}
		if newWatcherCount != watcherCount {
			watcherCount = newWatcherCount
			printAndFlush(w, "event:watcher_count\n"+"data:"+strconv.Itoa(watcherCount)+"\n\n")
		}
	}
}

func postAPIStrokesRoomsID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	t, err := checkToken(r.Header.Get("x-csrf-token"))

	if err != nil {
		outputError(w, err)
		return
	}
	if t == nil {
		outputErrorMsg(w, http.StatusBadRequest, "トークンエラー。ページを再読み込みしてください。")
		return
	}

	idStr := pat.Param(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		outputErrorMsg(w, http.StatusNotFound, "この部屋は存在しません。")
		return
	}

	room, err := getRoom(id)
	if err != nil {
		if err == sql.ErrNoRows {
			outputErrorMsg(w, http.StatusNotFound, "この部屋は存在しません。")
		} else {
			outputError(w, err)
		}
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		outputError(w, err)
		return
	}
	postedStroke := Stroke{}
	err = json.Unmarshal(body, &postedStroke)
	if err != nil {
		outputError(w, err)
		return
	}

	if postedStroke.Width == 0 || len(postedStroke.Points) == 0 {
		outputErrorMsg(w, http.StatusBadRequest, "リクエストが正しくありません。")
		return
	}

	strokes, err := getStrokes(room.ID, 0)
	if err != nil {
		outputError(w, err)
		return
	}
	if len(strokes) == 0 {
		query := "SELECT COUNT(*) AS cnt FROM `room_owners` WHERE `room_id` = ? AND `token_id` = ?"
		cnt := 0
		err = dbx.QueryRow(query, room.ID, t.ID).Scan(&cnt)
		if err != nil {
			outputError(w, err)
			return
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
	strokeID, err := result.LastInsertId()
	if err != nil {
		outputError(w, err)
		return
	}

	query = "INSERT INTO `points` (`stroke_id`, `x`, `y`) VALUES (?, ?, ?)"
	for _, p := range postedStroke.Points {
		tx.MustExec(query, strokeID, p.X, p.Y)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		outputError(w, err)
		return
	}

	query = "SELECT `id`, `room_id`, `width`, `red`, `green`, `blue`, `alpha`, `created_at` FROM `strokes`"
	query += " WHERE `id` = ?"
	s := Stroke{}
	err = dbx.Get(&s, query, strokeID)
	if err != nil {
		outputError(w, err)
		return
	}

	s.Points, err = getStrokePoints(strokeID)
	if err != nil {
		outputError(w, err)
		return
	}

	b, _ := json.Marshal(struct {
		Stroke Stroke `json:"stroke"`
	}{Stroke: s})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
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
	dbname := "isuketch"

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
	mux.HandleFunc(pat.Post("/api/csrf_token"), postAPICsrfToken)
	mux.HandleFunc(pat.Get("/api/rooms"), getAPIRooms)
	mux.HandleFunc(pat.Post("/api/rooms"), postAPIRooms)
	mux.HandleFuncC(pat.Get("/api/rooms/:id"), getAPIRoomsID)
	mux.HandleFuncC(pat.Get("/api/stream/rooms/:id"), getAPIStreamRoomsID)
	mux.HandleFuncC(pat.Post("/api/strokes/rooms/:id"), postAPIStrokesRoomsID)

	log.Fatal(http.ListenAndServe("0.0.0.0:80", mux))
}
