package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Errors []string `json:"errors"`
	Logs   []Log    `json:"logs"`
}

var initialWatcherNum int

var watcherIncreaseInterval int

var timeout int

var listen string

var targetScheme string

func main() {
	flag.IntVar(&initialWatcherNum, "initialWatcherNum", 5, "最初に入室するクライアント数")
	flag.IntVar(&watcherIncreaseInterval, "watcherIncreaseInterval", 5, "何秒ごとにクライアントを増やすか")
	flag.IntVar(&timeout, "timeout", 55, "何秒でクライアントを増やし続けるのをやめてタイムアウトとするか")
	flag.StringVar(&listen, "listen", "0.0.0.0:10080", "listenするIPとport (例: 0.0.0.0:10080)")
	flag.StringVar(&targetScheme, "targetScheme", "https", "targetのURLスキーム")
	flag.Parse()

	fmt.Println("listening on " + listen)
	http.HandleFunc("/", handler)
	http.ListenAndServe(listen, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	targetHost := r.URL.Query().Get("host")
	room := r.URL.Query().Get("room")
	roomID, err := strconv.Atoi(room)
	if err != nil || targetHost == "" {
		w.WriteHeader(400)
		w.Write([]byte("引数が間違っています (例: /?host=127.0.0.1&room=1)"))
	}
	targetURL := targetScheme + "://" + targetHost

	watchers := make([]*RoomWatcher, 0)

	fmt.Println("start")

	// まず最初にinitialWatcherNum人が入室する
	for i := 0; i < initialWatcherNum; i++ {
		fmt.Println("watcher", len(watchers)+1)
		watchers = append(watchers, NewRoomWatcher(targetURL, roomID))
	}

	numToIncreaseWatcher := (timeout - watcherIncreaseInterval) / watcherIncreaseInterval
	for k := 0; k < numToIncreaseWatcher; k++ {
		// watcherIncreaseIntervalごとにその時点でまだ退室していない参加人数の数と同じ人数が入ってくる
		time.Sleep(time.Duration(watcherIncreaseInterval) * time.Second)

		for _, w := range watchers {
			if len(w.EndCh) == 0 {
				fmt.Println("watcher", len(watchers)+1)
				watchers = append(watchers, NewRoomWatcher(targetURL, roomID))
			}
		}
	}

	time.Sleep(time.Duration(watcherIncreaseInterval) * time.Second)

	// ここまでで合計 timeout 秒かかり、
	// 最大で initialWatcherNum * 2 ^ numToIncreaseWatcher 人が入室してる

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
