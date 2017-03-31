package scenario

import (
	"time"

	"github.com/sesta/isucon6-final/bench/fails"
	"github.com/sesta/isucon6-final/bench/score"
	"github.com/sesta/isucon6-final/bench/seed"
	"github.com/sesta/isucon6-final/bench/session"
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

	postedStrokes := make(map[int64]Stroke)

	go func() {
		// 2秒おきにstrokeをPOSTする
		for {
			for _, seedStroke := range seedStrokes {
				postTime := time.Now()

				stroke, ok := drawStroke(s, token, room.ID, seed.FluctuateStroke(seedStroke))
				if ok {
					postTimes[stroke.ID] = postTime
					postedStrokes[stroke.ID] = *stroke
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
		// watcherIncreaseInterval秒おきに、 (まだ退室していないwatcherの数 - 既に退室したwatcherの数) の人数が入室する

		n := 0
		for _, w := range watchers {
			if len(w.StrokeLogs) > 0 && len(w.EndCh) == 0 { // 既にStrokeLogを1つ以上受け取ってる、かつ、まだ退室してないwatcherと同数のwatcherが入室する
				n++
			} else { // ただし、既に退室した人数をペナルティとする
				n--
			}
		}

		if n <= 0 { // ゼロならinitialWatcherNum人が入室する（特に初回）
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

	for _, w := range watchers {
		for _, strokeLog := range w.StrokeLogs {
			if postedStroke, ok := postedStrokes[strokeLog.Stroke.ID]; ok {
				if len(postedStroke.Points) != len(strokeLog.Points) {
					fails.Add("streamされたstrokeが間違っています", nil)
				} else {
					if postTime, ok := postTimes[strokeLog.Stroke.ID]; ok {
						timeTaken := strokeLog.ReceivedTime.Sub(postTime).Seconds()

						if timeTaken < 2 {
							score.Increment(StrokeReceiveScore)
						} else {
							// 2秒以上かかって届いた。スコア増えない。5秒以上かかった場合はそこで退室したはず。
						}
					} else {
						// POSTしてないstrokeが届いた。普通はありえない
					}
				}
			} else {
				// POSTしてないstrokeが届いた。普通はありえない
			}
		}
	}
}
