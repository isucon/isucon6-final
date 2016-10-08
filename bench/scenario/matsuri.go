package scenario

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"

	"github.com/catatsuy/isucon6-final/bench/fails"
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
	s := newSession(origins)

	token, ok := fetchCSRFToken(s, "/")
	if !ok {
		return
	}

	roomID, ok := makeRoom(s, token)
	if !ok {
		return
	}

	seedStroke := seed.GetStroke("main001")

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
		watchers = append(watchers, NewRoomWatcher(origins[rand.Intn(len(origins))], roomID))
	}

	numToIncreaseWatcher := (timeout - watcherIncreaseInterval) / watcherIncreaseInterval
	for k := 0; k < numToIncreaseWatcher; k++ {
		// watcherIncreaseIntervalごとにその時点でまだ退室していない参加人数の数と同じ人数が入ってくる
		time.Sleep(time.Duration(watcherIncreaseInterval) * time.Second)

		for _, w := range watchers {
			if len(w.EndCh) == 0 {
				//fmt.Println("watcher", len(watchers)+1)
				watchers = append(watchers, NewRoomWatcher(origins[rand.Intn(len(origins))], roomID))
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
}

func makeRoom(s *session.Session, token string) (int64, bool) {
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

	var roomID int64

	ok := s.Post("/api/rooms", postBody, headers, func(body io.Reader, l *fails.Logger) bool {
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
		roomID = res.Room.ID

		return true
	})

	return roomID, ok
}

func drawStroke(s *session.Session, token string, roomID int64, stroke seed.Stroke) (int64, bool) {
	postBody, _ := json.Marshal(struct {
		RoomID int64 `json:"room_id"`
		seed.Stroke
	}{
		RoomID: roomID,
		Stroke: stroke,
	})

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-csrf-token": token,
	}

	var strokeID int64

	u := "/api/strokes/rooms/" + strconv.FormatInt(roomID, 10)
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

		strokeID = res.Stroke.ID

		return true
	})

	return strokeID, ok
}
