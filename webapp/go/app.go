package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"net/http"
	"strconv"
	"time"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var staticPath string

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
	flag.StringVar(&staticPath, "staticpath", "", "static file directory")
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
	// if lastID == "" {
	// 	lastID = r.URL.Query().Get("last-id")
	// }

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
		time.Sleep(1 * time.Second)
		// TODO: break loop when connection is closed
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
	// TODO: save
	w.WriteHeader(http.StatusOK)
	// TODO: broadcast

	stroke.ID = randomString()
	strokes = append(strokes, stroke)
	// fmt.Println(stroke)
}

func randomString() string {
	var n uint64
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	return strconv.FormatUint(n, 36)
}

func main() {
	flag.Parse()

	goji.Post("/api/stroke", stroke)
	goji.Get("/api/events", events)
	goji.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	goji.Serve()
}
