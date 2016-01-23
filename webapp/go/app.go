package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var staticPath string
var templatePath string

// Stroke is a single stroke of drawing
type Stroke struct {
	ID    string  `json:"id"`
	Width int     `json:"width"`
	Red   int     `json:"red"`
	Green int     `json:"green"`
	Blue  int     `json:"blue"`
	Alpha float32 `json:"alpha"`
	Xs    []int   `json:"xs"`
	Ys    []int   `json:"ys"`
}

func init() {
	flag.StringVar(&staticPath, "staticpath", "../static", "static file directory")
	flag.StringVar(&templatePath, "templatepath", "templates", "template file directory")
}

// TODO: save to DB
var strokes []Stroke

// GetStrokesSince returns strokes drawn after a given ID
func GetStrokesSince(id string) []Stroke {
	ret := []Stroke{}
	skip := true
	if id == "" {
		skip = false
	}
	for _, stroke := range strokes {
		if stroke.ID == id {
			skip = false
			continue
		}
		if skip {
			continue
		}
		ret = append(ret, stroke)
	}
	return ret
}

func events(c web.C, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	lastID := r.Header.Get("Last-Event-ID")
	// TODO: maybe needed for manual reconnect?
	// http://stackoverflow.com/questions/24564030/is-an-eventsource-sse-supposed-to-try-to-reconnect-indefinitely
	// if lastID == "" {
	// 	lastID = r.URL.Query().Get("last-id")
	// }

	cn, ok := w.(http.CloseNotifier)
	if !ok {
		panic("ResponseWriter does not support the CloseNotifier interface")
	}
	closed := false
	go func() {
		<-cn.CloseNotify()
		closed = true
	}()

	for {
		for _, stroke := range GetStrokesSince(lastID) {
			var b []byte
			buf := bytes.NewBuffer(b)

			// TODO: return on error
			_, _ = buf.WriteString("id: ")
			_, _ = buf.WriteString(stroke.ID)
			_, _ = buf.WriteString("\n")
			_, _ = buf.WriteString("data: ")
			j, _ := json.Marshal(stroke)
			_, _ = buf.Write(j)
			_, _ = buf.WriteString("\n\n")
			_, _ = buf.WriteTo(w)
			flusher, ok := w.(http.Flusher)
			if !ok {
				panic("ResponseWriter does not support the Flusher interface")
			}
			flusher.Flush()

			lastID = stroke.ID
		}
		if closed {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func stroke(c web.C, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var stroke Stroke
	err := decoder.Decode(&stroke)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stroke.ID = randomString(36)
	strokes = append(strokes, stroke)

	w.WriteHeader(http.StatusOK)
}

// RoomData is passed to the template
type RoomData struct {
	ID          string
	Title       string
	RoomURL     string
	ImageURL    string
	MemberCount int
}

var AllRooms []RoomData

func findRoomByID(id string) *RoomData {
	for _, room := range AllRooms {
		if id == room.ID {
			return &room
		}
	}
	return nil
}

func index(c web.C, w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(templatePath + "/index.html"))
	var rd [30]RoomData
	t.Execute(w, struct{ Rooms []RoomData }{rd[:]})
}

var store = sessions.NewCookieStore([]byte("isucon-nocusi"))

func getSession(r *http.Request) *sessions.Session {
	// safe to ignore the error because a new session is also returned
	// when the cookie could not be decoded
	session, _ := store.Get(r, "ISU-SESSION")
	return session
}

func room(c web.C, w http.ResponseWriter, r *http.Request) {
	roomID := c.URLParams["id"]

	room := findRoomByID(roomID)
	if room == nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	// room := &RoomData{}

	session := getSession(r)

	val := session.Values["csrf-token"] // session.Values is map[string]interface{}
	token, ok := val.(string)
	if !ok {
		token := randomString(128)
		session.Values["csrf-token"] = token
		session.Save(r, w)
	}

	t := template.Must(template.ParseFiles(templatePath + "/room.html"))
	t.Execute(w, struct {
		Room      RoomData
		CSRFToken string
		APIURL    string
	}{
		*room,
		token,
		"http://" + r.Host,
	})
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	flag.Parse()

	goji.Get("/", index)
	goji.Get("/rooms/:id", room)

	goji.Post("/api/stroke", stroke)
	goji.Get("/api/events", events)
	goji.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	goji.Serve()
}
