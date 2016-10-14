package scenario

import (
	"time"

	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/score"
	"github.com/catatsuy/isucon6-final/bench/seed"
	"github.com/catatsuy/isucon6-final/bench/session"
)

const (
	initialWatcherNum             = 10
	watcherIncreaseInterval       = 5
	StrokeReceiveScore      int64 = 1
)

// 一人がroomを作る→大勢がそのroomをwatchする
func Matsuri(origins []string, timeout int) {
	s := session.New(randomOrigin(origins))
	defer s.Bye()

	token, ok := fetchCSRFToken(s, "/")
	if !ok {
		return
	}

	room, ok := makeRoom(s, token)
	if !ok {
		return
	}

	seedStrokes := seed.GetStrokes("isu")

	postTimes := make(map[int64]time.Time)

	start := time.Now()

	postedStrokes := make([]Stroke, 0)

	go func() {
		// 2秒おきにstrokeをPOSTする
		for {
			for _, seedStroke := range seedStrokes {
				postTime := time.Now()

				stroke, ok := drawStroke(s, token, room.ID, seed.FluctuateStroke(seedStroke))
				if ok {
					postTimes[stroke.ID] = postTime
					postedStrokes = append(postedStrokes, *stroke)
				}
				time.Sleep(2 * time.Second)
				if time.Now().Sub(start).Seconds() > float64(timeout) {
					return
				}
			}
		}
	}()

	watchers := make([]*RoomWatcher, 0)

	for {
		// watcherIncreaseInterval秒おきに、まだ退室していないwatcherの数と同数の人数が入室する

		n := 0
		for _, w := range watchers {
			if len(w.StrokeLogs) > 0 && len(w.EndCh) == 0 { // 既に最初のStrokeLogを1つ以上受け取ってる、かつ、まだ退室してない
				n++
			}
		}
		if n == 0 { // ゼロならinitialWatcherNum人が入室する（特に初回）
			n = initialWatcherNum
		}
		for i := 0; i < n; i++ {
			watchers = append(watchers, NewRoomWatcher(randomOrigin(origins), room.ID))
		}

		time.Sleep(time.Duration(watcherIncreaseInterval) * time.Second)
		if time.Now().Sub(start).Seconds() > float64(timeout-watcherIncreaseInterval) {
			break
		}
	}

	// ここまでで最大 initialWatcherNum * 2 ^ ((timeout - watcherIncreaseInterval) / watcherIncreaseInterval) 人が入室してるはず
	// 例えば initialWatcherNum=10, timeout=55, watcherIncreaseInterval=5 なら 10 * 2 ^ ((55-5)/5) = 10240 人

	//fmt.Println("stop")
	for _, w := range watchers {
		w.Leave()
	}
	//fmt.Println("wait")
	for _, w := range watchers {
		<-w.EndCh
	}
	//fmt.Println("done")

	// TODO: watcher_countが正しいか

	for _, w := range watchers {
		for i, strokeLog := range w.StrokeLogs {
			postTime := postTimes[strokeLog.Stroke.ID]
			timeTaken := strokeLog.ReceivedTime.Sub(postTime).Seconds()

			if timeTaken < 2 {
				score.Increment(StrokeReceiveScore)
			}

			if i >= len(postedStrokes) {
				// 普通は起こらないはず
				break
			}
			if strokeLog.ID != postedStrokes[i].ID {
				fails.Critical("streamされたstrokeに抜け・狂いがあります", nil)
				break
			}
			if len(strokeLog.Points) != len(postedStrokes[i].Points) {
				fails.Critical("streamされたstrokeが間違っています", nil)
				break
			}
		}
	}
}
