package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/catatsuy/isucon6-final/bench/scenario"
	"github.com/catatsuy/isucon6-final/bench/session"
	"github.com/catatsuy/isucon6-final/bench/sse"
)

func watch(target string, roomID int) {
	s := session.New(target)

	token, err := scenario.GetCSRFTokenFromRoom(s, roomID)
	if err != nil {
		return // 何を返す？
	}

	u := fmt.Sprintf("%s://%s/api/strokes/rooms/%d?csrf_token=%s", s.Scheme, s.Host, roomID, token)
	stream := sse.NewStream(s.Client, u)

	stream.On("stroke", func(data string) {
		fmt.Println("stroke")
		fmt.Println(data)
	})
	stream.On("bad_request", func(data string) {
		fmt.Println("bad_request")
		fmt.Println(data)
		stream.Close()
	})
	stream.On("watcher_count", func(data string) {
		fmt.Println("watcher_count")
		fmt.Println(data)
	})
	stream.OnError(func(err error) {
		fmt.Println("error")

		if e, ok := err.(*sse.BadContentType); ok {
			fmt.Println("bad content type " + e.ContentType)
		}
		if e, ok := err.(*sse.BadStatusCode); ok {
			fmt.Printf("bad status code %d\n", e.StatusCode)
			if 400 <= e.StatusCode && e.StatusCode < 500 {
				stream.Close()
			}
		}
		fmt.Println(err)
	})

	go stream.Start()

	time.Sleep(30 * time.Second)

	fmt.Println("close")
	stream.Close()
}

func handler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	room := r.URL.Query().Get("room")
	roomID, err := strconv.Atoi(room)
	if err != nil || target == "" {
		w.WriteHeader(400)
		w.Write([]byte("引数が間違っています (例: /?target=https%3A%2F%2F127.0.0.1&room=1)"))
	}

	watch(target, roomID)

	fmt.Fprintf(w, "Hello, World")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe("0.0.0.0:10080", nil)
}
