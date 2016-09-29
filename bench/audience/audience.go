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

var initialWatcherNum = 5

var watcherIncreaseInterval = 5 * time.Second

var watcherIncreaseTimes = 5

func handler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	room := r.URL.Query().Get("room")
	roomID, err := strconv.Atoi(room)
	if err != nil || target == "" {
		w.WriteHeader(400)
		w.Write([]byte("引数が間違っています (例: /?target=https%3A%2F%2F127.0.0.1&room=1)"))
	}

	watchers := make([]*RoomWatcher, 0)

	// まず最初にinitialWatcherNum人が入室する
	for i := 0; i < initialWatcherNum; i++ {
		fmt.Println("watcher", len(watchers)+1)
		watchers = append(watchers, NewRoomWatcher(target, roomID))
	}
	fmt.Println("start")

	for k := 0; k < watcherIncreaseTimes; k++ {
		// watcherIncreaseIntervalごとにその時点でまだ退室していない参加人数の数と同じ人数が入ってくる
		time.Sleep(watcherIncreaseInterval)

		for _, w := range watchers {
			if len(w.EndCh) == 0 {
				fmt.Println("watcher", len(watchers)+1)
				watchers = append(watchers, NewRoomWatcher(target, roomID))
			}
		}
	}

	time.Sleep(watcherIncreaseInterval)

	// ここまでで合計 watcherIncreaseInterval * watcherIncreaseTimes 秒かかり、
	// 最大で initialWatcherNum * 2 ^ watcherIncreaseTimes 人が入室してる

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
		res.Logs = append(res.Logs, w.Logs...)
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
