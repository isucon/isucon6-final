package scenario

import (
	"time"

	"github.com/catatsuy/isucon6-final/bench/score"
	"github.com/catatsuy/isucon6-final/bench/seed"
	"github.com/catatsuy/isucon6-final/bench/session"
)

const (
	initialWatcherNum       = 5
	watcherIncreaseInterval = 5
)

// 一人がroomを作る→大勢がそのroomをwatchする
func Matsuri(origins []string, timeout int) {
	s := session.New(randomOrigin(origins))
	defer s.Bye()

	token, ok := fetchCSRFToken(s, "/")
	if !ok {
		return
	}

	roomID, ok := makeRoom(s, token)
	if !ok {
		return
	}

	seedStroke := seed.GetStrokes("isu")

	postTimes := make(map[int64]time.Time)

	go func() {
		for {
			for _, stroke := range seedStroke {
				postTime := time.Now()

				strokeID, _ := drawStroke(s, token, roomID, stroke)
				// 特に止める必要もない
				postTimes[strokeID] = postTime
			}
		}
	}()

	watchers := make([]*RoomWatcher, 0)

	// まず最初にinitialWatcherNum人が入室する
	for i := 0; i < initialWatcherNum; i++ {
		//fmt.Println("watcher", len(watchers)+1)
		watchers = append(watchers, NewRoomWatcher(randomOrigin(origins), roomID))
	}

	numToIncreaseWatcher := (timeout - watcherIncreaseInterval) / watcherIncreaseInterval
	for k := 0; k < numToIncreaseWatcher; k++ {
		// watcherIncreaseIntervalごとにその時点でまだ退室していない参加人数の数と同じ人数が入ってくる
		time.Sleep(time.Duration(watcherIncreaseInterval) * time.Second)

		for _, w := range watchers {
			if len(w.EndCh) == 0 {
				//fmt.Println("watcher", len(watchers)+1)
				watchers = append(watchers, NewRoomWatcher(randomOrigin(origins), roomID))
			}
		}
	}

	time.Sleep(time.Duration(watcherIncreaseInterval) * time.Second)

	// ここまでで合計 timeout 秒かかり、
	// 最大で initialWatcherNum * 2 ^ numToIncreaseWatcher 人が入室してる

	//fmt.Println("stop")

	for _, w := range watchers {
		w.Leave()
	}
	//fmt.Println("wait")
	for _, w := range watchers {
		<-w.EndCh
	}
	//fmt.Println("done")

	StrokeLogs := []StrokeLog{}
	for _, w := range watchers {
		StrokeLogs = append(StrokeLogs, w.StrokeLogs...)
	}

	for _, strokeLog := range StrokeLogs {
		postTime := postTimes[strokeLog.Stroke.ID]
		timeTaken := strokeLog.ReceivedTime.Sub(postTime).Seconds()
		if timeTaken < 1 { // TODO: この時間は要調整
			score.Increment(StrokeReceiveScore * 2)
		} else if timeTaken < 3 {
			score.Increment(StrokeReceiveScore)
		}
	}
}
