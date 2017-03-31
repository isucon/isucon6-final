package scenario

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/sesta/isucon6-final/bench/action"
	"github.com/sesta/isucon6-final/bench/fails"
	"github.com/sesta/isucon6-final/bench/session"
	"github.com/sesta/isucon6-final/bench/sse"
)

type StrokeLog struct {
	ReceivedTime time.Time
	Stroke
}

type WatcherCountLog struct {
	ReceivedTime time.Time
	Count        int
}

type RoomWatcher struct {
	EndCh            chan struct{}
	StrokeLogs       []StrokeLog
	WatcherCountLogs []WatcherCountLog

	s      *session.Session
	es     *sse.EventSource
	isLeft bool
}

func NewRoomWatcher(target string, roomID int64) *RoomWatcher {
	w := &RoomWatcher{
		EndCh:            make(chan struct{}, 1),
		StrokeLogs:       make([]StrokeLog, 0),
		WatcherCountLogs: make([]WatcherCountLog, 0),
		isLeft:           false,
		s:                session.New(target),
	}

	go w.watch(roomID)

	return w
}

// 描いたstrokeがこの時間以上経ってから届いたら、ユーザーがストレスに感じてタブを閉じる、という設定にした。
const thresholdResponseTime = 5 * time.Second

func (w *RoomWatcher) watch(roomID int64) {

	path := fmt.Sprintf("/rooms/%d", roomID)
	token, ok := fetchCSRFToken(w.s, path)
	if !ok || w.isLeft {
		w.finalize()
		return
	}

	path = "/api/stream" + path
	l := &fails.Logger{Prefix: "[" + path + "] "}

	values := url.Values{}
	values.Add("csrf_token", token)

	startTime := time.Now()
	w.es, ok = action.SSE(w.s, path+"?"+values.Encode())
	if !ok {
		w.finalize()
		return
	}

	w.es.On("stroke", func(data string) {
		now := time.Now()
		var stroke Stroke
		err := json.Unmarshal([]byte(data), &stroke)
		if err != nil {
			l.Add("jsonのデコードに失敗しました", err)
			w.es.Close()
		}
		// strokes APIには最初はLast-Event-IDをつけずに送るので、これまでに描かれたstrokeが全部降ってくるが、それは無視する。
		if stroke.CreatedAt.After(startTime) && now.Sub(stroke.CreatedAt) > thresholdResponseTime {
			l.Add("strokeが届くまでに時間がかかりすぎています", nil)
			w.es.Close()
		}
		w.StrokeLogs = append(w.StrokeLogs, StrokeLog{
			ReceivedTime: now,
			Stroke:       stroke,
		})
	})
	w.es.On("bad_request", func(data string) {
		l.Add("bad_request: "+data, nil)
		w.es.Close()
	})
	w.es.On("watcher_count", func(data string) {
		now := time.Now()
		count, err := strconv.Atoi(data)
		if err != nil {
			l.Add("watcher_countがパースできませんでした "+data, err)
		}
		w.WatcherCountLogs = append(w.WatcherCountLogs, WatcherCountLog{
			ReceivedTime: now,
			Count:        count,
		})
	})
	w.es.OnError(func(err error) {
		if e, ok := err.(*sse.BadContentType); ok {
			l.Add("Content-Typeが正しくありません: "+e.ContentType, err)
			return
		}
		if e, ok := err.(*sse.BadStatusCode); ok {
			l.Add(fmt.Sprintf("ステータスコードが正しくありません: %d", e.StatusCode), err)
			w.es.Close()
			return
		}
		l.Add("リクエストに失敗しました", err)
	})
	w.es.OnEnd(func() {
		w.finalize()
	})

	w.es.Open()
}

// Watcherを部屋から退出させるために呼ぶ。Leaveを呼ばれたらWatcher内部でクリーンアップ処理などをし、EndChに通知が行く
func (w *RoomWatcher) Leave() {
	w.isLeft = true
	if w.es != nil {
		w.es.Close()
	}
}

func (w *RoomWatcher) finalize() {
	w.s.Bye()
	w.EndCh <- struct{}{}
}
