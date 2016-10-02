package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/catatsuy/isucon6-final/bench/audience"
	"github.com/catatsuy/isucon6-final/bench/http"
)

var initialWatcherNum int

var watcherIncreaseInterval int

var timeout int

var listen string

func main() {
	flag.IntVar(&initialWatcherNum, "initialWatcherNum", 5, "最初に入室するクライアント数")
	flag.IntVar(&watcherIncreaseInterval, "watcherIncreaseInterval", 5, "何秒ごとにクライアントを増やすか")
	flag.IntVar(&timeout, "timeout", 55, "何秒でクライアントを増やし続けるのをやめてタイムアウトとするか")
	flag.StringVar(&listen, "listen", "0.0.0.0:10080", "listenするIPとport (例: 0.0.0.0:10080)")
	flag.Parse()

	fmt.Println("listening on " + listen)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	scheme := r.URL.Query().Get("scheme")
	host := r.URL.Query().Get("host")
	room := r.URL.Query().Get("room")
	roomID, err := strconv.ParseInt(room, 10, 64)
	if err != nil || scheme == "" || host == "" {
		w.WriteHeader(400)
		w.Write([]byte("引数が間違っています (例: /?scheme=https&host=127.0.0.1&room=1)"))
		return
	}

	baseURL := scheme + "://" + host
	watchers := make([]*audience.RoomWatcher, 0)

	fmt.Println("start")

	// まず最初にinitialWatcherNum人が入室する
	for i := 0; i < initialWatcherNum; i++ {
		fmt.Println("watcher", len(watchers)+1)
		watchers = append(watchers, audience.NewRoomWatcher(baseURL, roomID))
	}

	numToIncreaseWatcher := (timeout - watcherIncreaseInterval) / watcherIncreaseInterval
	for k := 0; k < numToIncreaseWatcher; k++ {
		// watcherIncreaseIntervalごとにその時点でまだ退室していない参加人数の数と同じ人数が入ってくる
		time.Sleep(time.Duration(watcherIncreaseInterval) * time.Second)

		for _, w := range watchers {
			if len(w.EndCh) == 0 {
				fmt.Println("watcher", len(watchers)+1)
				watchers = append(watchers, audience.NewRoomWatcher(baseURL, roomID))
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

	res := &audience.AudienceResponse{
		Errors:     make([]string, 0),
		StrokeLogs: make([]audience.StrokeLog, 0),
	}
	for _, w := range watchers {
		res.Errors = append(res.Errors, w.Errors...)
		res.StrokeLogs = append(res.StrokeLogs, w.Logs...)
	}

	b, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}
