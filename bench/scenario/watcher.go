package scenario

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/session"
	"github.com/catatsuy/isucon6-final/bench/sse"
)

type RoomWatcher struct {
	EndCh chan struct{}
	Logs  []StrokeLog

	es     *sse.EventSource
	isLeft bool
}

func NewRoomWatcher(target string, roomID int64) *RoomWatcher {
	w := &RoomWatcher{
		EndCh:  make(chan struct{}, 1),
		Logs:   make([]StrokeLog, 0),
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
	l := &fails.Logger{Prefix: "[" + path + "] "}

	token, ok := fetchCSRFToken(s, path)
	if !ok {
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
	w.es.AddHeader("User-Agent", s.UserAgent)

	w.es.On("stroke", func(data string) {
		var stroke Stroke
		err := json.Unmarshal([]byte(data), &stroke)
		if err != nil {
			l.Add("jsonのデコードに失敗しました", err)
			w.es.Close()
		}
		now := time.Now()
		// strokes APIには最初はLast-Event-IDをつけずに送るので、これまでに描かれたstrokeが全部降ってくるが、それは無視する。
		if stroke.CreatedAt.After(startTime) && now.Sub(stroke.CreatedAt) > thresholdResponseTime {
			l.Add("strokeが届くまでに時間がかかりすぎています", nil)
			w.es.Close()
		}
		w.Logs = append(w.Logs, StrokeLog{
			ReceivedTime: now,
			RoomID:       roomID,
			StrokeID:     stroke.ID,
		})
	})
	w.es.On("bad_request", func(data string) {
		l.Add("bad_request: "+data, nil)
		w.es.Close()
	})
	//w.es.On("watcher_count", func(data string) {
	//	fmt.Println("watcher_count")
	//	fmt.Println(data)
	//})
	w.es.OnError(func(err error) {
		if e, ok := err.(*sse.BadContentType); ok {
			l.Add(path+" Content-Typeが正しくありません: "+e.ContentType, err)
			return
		}
		if e, ok := err.(*sse.BadStatusCode); ok {
			l.Add(fmt.Sprintf("ステータスコードが正しくありません: %d", e.StatusCode), err)
			w.es.Close()
			return
		}
		l.Add("予期せぬエラー（主催者に連絡してください）", err)
	})
	w.es.OnEnd(func() {
		w.EndCh <- struct{}{}
	})

	w.es.Start()
}

func (w *RoomWatcher) Leave() {
	w.isLeft = true
	if w.es != nil {
		w.es.Close()
	}
}
