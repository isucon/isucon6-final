package scenario

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"fmt"

	"math/rand"

	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/score"
	"github.com/catatsuy/isucon6-final/bench/seed"
)

const (
	initialWatcherNum       = 5
	watcherIncreaseInterval = 5
)

// 一人がroomを作る→大勢がそのroomをwatchする
func Matsuri(origins []string, timeoutCh chan struct{}) {
	s := newSession(origins)

	var token string

	ok := s.Get("/", func(body io.Reader, l *fails.Logger) bool {
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}

		token, ok = extractCsrfToken(doc, l)
		if !ok {
			return false
		}

		return true
	})
	if !ok {
		return
	}

	postBody, _ := json.Marshal(struct {
		Name         string `json:"name"`
		CanvasWidth  int    `json:"canvas_width"`
		CanvasHeight int    `json:"canvas_height"`
	}{
		Name:         "ひたすら椅子を描く部屋",
		CanvasWidth:  1024,
		CanvasHeight: 768,
	})

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-csrf-token": token,
	}

	var RoomID int64

	ok = s.Post("/api/rooms", postBody, headers, func(body io.Reader, l *fails.Logger) bool {
		b, err := ioutil.ReadAll(body)
		if err != nil {
			l.Add("レスポンス内容が読み込めませんでした", err)
			return false
		}
		var res Response
		err = json.Unmarshal(b, &res)
		if err != nil {
			l.Add("レスポンス内容が正しくありません"+string(b[:20]), err)
			return false
		}
		if res.Room == nil || res.Room.ID <= 0 {
			l.Add("レスポンス内容が正しくありません"+string(b[:20]), nil)
			return false
		}
		RoomID = res.Room.ID

		return true
	})

	if !ok {
		return
	}

	seedStroke := seed.GetStroke("main001")

	postTimes := make(map[int64]time.Time)

	end := make(chan struct{})

	go func() {
		for {
			for _, stroke := range seedStroke {
				postBody, _ := json.Marshal(struct {
					RoomID int64 `json:"room_id"`
					seed.Stroke
				}{
					RoomID: RoomID,
					Stroke: stroke,
				})

				postTime := time.Now()

				u := "/api/strokes/rooms/" + strconv.FormatInt(RoomID, 10)
				ok := s.Post(u, postBody, headers, func(body io.Reader, l *fails.Logger) bool {

					b, err := ioutil.ReadAll(body)
					if err != nil {
						l.Add("レスポンス内容が読み込めませんでした", err)
						return false
					}

					var res Response
					err = json.Unmarshal(b, &res)
					if err != nil {
						l.Add("レスポンス内容が正しくありません"+string(b[:20]), err)
						return false
					}
					if res.Stroke == nil || res.Stroke.ID <= 0 {
						l.Add("レスポンス内容が正しくありません"+string(b[:20]), nil)
						return false
					}

					postTimes[res.Stroke.ID] = postTime

					return true
				})
				if !ok || len(timeoutCh) > 0 {
					end <- struct{}{}
				}
			}
		}
	}()
	/*
		v := url.Values{}
		v.Set("scheme", s.Scheme)
		v.Set("host", s.Host)
		v.Set("room", strconv.FormatInt(RoomID, 10))

		resp, err := http.Get(aud + "?" + v.Encode())
		if err != nil {
			fails.Add("予期せぬエラー (主催者に連絡してください)", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			_, err := ioutil.ReadAll(resp.Body)
			fails.Add("予期せぬエラー (主催者に連絡してください)", err)
			return
		}

		var audRes AudienceResponse
		err = json.NewDecoder(resp.Body).Decode(&audRes)
		if err != nil {
			fails.Add("予期せぬエラー (主催者に連絡してください)", err)
			return
		}
		for _, msg := range audRes.Errors {
			fails.Add(msg, nil)
		}*/
	watchers := make([]*RoomWatcher, 0)

	// まず最初にinitialWatcherNum人が入室する
	for i := 0; i < initialWatcherNum; i++ {
		fmt.Println("watcher", len(watchers)+1)
		watchers = append(watchers, NewRoomWatcher(origins[rand.Intn(len(origins))], RoomID))
	}

	numToIncreaseWatcher := (55 - watcherIncreaseInterval) / watcherIncreaseInterval // TODO: マジックナンバー
	for k := 0; k < numToIncreaseWatcher; k++ {
		// watcherIncreaseIntervalごとにその時点でまだ退室していない参加人数の数と同じ人数が入ってくる
		time.Sleep(time.Duration(watcherIncreaseInterval) * time.Second)

		for _, w := range watchers {
			if len(w.EndCh) == 0 {
				fmt.Println("watcher", len(watchers)+1)
				watchers = append(watchers, NewRoomWatcher(origins[rand.Intn(len(origins))], RoomID))
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

	StrokeLogs := []StrokeLog{}
	for _, w := range watchers {
		StrokeLogs = append(StrokeLogs, w.Logs...)
	}

	for _, strokeLog := range StrokeLogs {
		postTime := postTimes[strokeLog.StrokeID]
		timeTaken := strokeLog.ReceivedTime.Sub(postTime).Seconds()
		if timeTaken < 1 { // TODO: この時間は要調整
			score.Increment(StrokeReceiveScore * 2)
		} else if timeTaken < 3 {
			score.Increment(StrokeReceiveScore)
		}
	}

	<-end
}
