package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Errors []string `json:"errors"`
	Logs   []Log    `json:"logs"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	room := r.URL.Query().Get("room")
	roomID, err := strconv.Atoi(room)
	if err != nil || target == "" {
		w.WriteHeader(400)
		w.Write([]byte("引数が間違っています (例: /?target=https%3A%2F%2F127.0.0.1&room=1)"))
	}

	watchers := make([]*RoomWatcher, 0)
	for i := 0; i < 20; i++ {
		w := NewRoomWatcher(target, roomID)
		watchers = append(watchers, w)
	}
	fmt.Println("start")

	time.Sleep(30 * time.Second)
	fmt.Println("stop")

	for _, w := range watchers {
		w.Leave()
	}
	fmt.Println("wait")
	for _, w := range watchers {
		<-w.EndCh
	}
	fmt.Println("done")

	res := &Response{
		Errors: make([]string, 0),
		Logs:   make([]Log, 0),
	}
	for _, w := range watchers {
		res.Errors = append(res.Errors, w.Errors...)
		res.Logs = append(w.Logs, w.Logs...)
	}

	b, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe("0.0.0.0:10080", nil)
}
