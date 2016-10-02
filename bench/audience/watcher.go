package audience

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/catatsuy/isucon6-final/bench/scenario"
	"github.com/catatsuy/isucon6-final/bench/session"
	"github.com/catatsuy/isucon6-final/bench/sse"
)

type RoomWatcher struct {
	EndCh  chan struct{}
	Logs   []StrokeLog
	Errors []string

	es     *sse.EventSource
	isLeft bool
}

func NewRoomWatcher(target string, roomID int64) *RoomWatcher {
	w := &RoomWatcher{
		EndCh:  make(chan struct{}, 1),
		Logs:   make([]StrokeLog, 0),
		Errors: make([]string, 0),
		isLeft: false,
	}

	go w.watch(target, roomID)

	return w
}

// 描いたstrokeがこの時間以上経ってから届いたら、ユーザーがストレスに感じてタブを閉じる、という設定にした。
const thresholdResponseTime = 5 * time.Second

func (w *RoomWatcher) watch(target string, roomID int64) {

	// TODO:用途がだいぶ特殊なので普通のベンチマークと同じsessionを使うべきか悩ましい
	s := session.New(target)
	s.Client.Timeout = 3 * time.Second

	path := fmt.Sprintf("/rooms/%d", roomID)
	token, err := scenario.GetCSRFToken(s, target+path)
	if err != nil {
		w.addError(err.Error())
		w.EndCh <- struct{}{}
		return
	}

	startTime := time.Now()
	path = "/api/stream" + path

	if w.isLeft {
		w.EndCh <- struct{}{}
		return
	}
	w.es = sse.NewEventSource(s.Client, target+path+"?csrf_token="+token)

	w.es.On("stroke", func(data string) {
		var stroke scenario.Stroke
		err := json.Unmarshal([]byte(data), &stroke)
		if err != nil {
			w.Errors = append(w.Errors, err.Error())
			fmt.Println(err)
			w.es.Close()
		}
		now := time.Now()
		// strokes APIには最初はLast-Event-IDをつけずに送るので、これまでに描かれたstrokeが全部降ってくるが、それは無視する。
		if stroke.CreatedAt.After(startTime) && now.Sub(stroke.CreatedAt) > thresholdResponseTime {
			fmt.Println("response too late")
			w.es.Close()
		}
		w.Logs = append(w.Logs, StrokeLog{
			ReceivedTime: now,
			RoomID:       roomID,
			StrokeID:     stroke.ID,
		})
	})
	w.es.On("bad_request", func(data string) {
		w.addError(path + " bad_request: " + data)
		w.es.Close()
	})
	//w.es.On("watcher_count", func(data string) {
	//	fmt.Println("watcher_count")
	//	fmt.Println(data)
	//})
	w.es.OnError(func(err error) {
		if e, ok := err.(*sse.BadContentType); ok {
			w.addError(path + " Content-Typeが正しくありません: " + e.ContentType)
			return
		}
		if e, ok := err.(*sse.BadStatusCode); ok {
			w.addError(fmt.Sprintf("%s ステータスコードが正しくありません: %d\n", path, e.StatusCode))
			w.es.Close()
			return
		}
		fmt.Println(err)
		w.addError(path + " 予期せぬエラー")
	})
	w.es.OnEnd(func() {
		w.EndCh <- struct{}{}
	})

	w.es.Start()
}

func (w *RoomWatcher) addError(msg string) {
	w.Errors = append(w.Errors, fmt.Sprintf("%s", msg))
}

func (w *RoomWatcher) Leave() {
	w.isLeft = true
	if w.es != nil {
		w.es.Close()
	}
}
