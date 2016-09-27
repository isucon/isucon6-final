package main

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/catatsuy/isucon6-final/bench/scenario"
	"github.com/catatsuy/isucon6-final/bench/session"
	"github.com/catatsuy/isucon6-final/bench/sse"
)

type Log struct {
	Time       time.Time `json:"time"`
	RoomID     int       `json:"room_id"`
	StrokeID   int64     `json:"stroke_id"`
	StrokeTime time.Time `json:"stroke_time"`
}

type RoomWatcher struct {
	EndCh  chan struct{}
	Logs   []Log
	Errors []string

	es *sse.EventSource
}

func NewRoomWatcher(target string, roomID int) *RoomWatcher {
	w := &RoomWatcher{
		EndCh:  make(chan struct{}),
		Logs:   make([]Log, 0),
		Errors: make([]string, 0),
	}

	go w.watch(target, roomID)

	return w
}

// 描いたstrokeがこの時間以上経ってから届いたら、ユーザーがストレスに感じてタブを閉じる、という設定にした。
const thresholdResponseTime = 5 * time.Second

func (w *RoomWatcher) watch(target string, roomID int) {

	s := session.New(target)

	token, err := scenario.GetCSRFTokenFromRoom(s, roomID)
	if err != nil {
		w.Errors = append(w.Errors, err.Error()) // TODO: mutex
		fmt.Println(err)
		w.EndCh <- struct{}{}
		return
	}

	startTime := time.Now()
	u := fmt.Sprintf("%s://%s/api/strokes/rooms/%d?csrf_token=%s", s.Scheme, s.Host, roomID, token)
	w.es = sse.NewEventSource(s.Client, u)

	w.es.On("stroke", func(data string) {
		//fmt.Println("stroke")
		//fmt.Println(data)
		var stroke scenario.Stroke
		err := json.Unmarshal([]byte(data), &stroke)
		if err != nil {
			w.Errors = append(w.Errors, err.Error()) // TODO: mutex
			fmt.Println(err)
			w.Leave()
		}
		now := time.Now()
		// strokes APIには最初はLast-Event-IDをつけずに送るので、これまでに描かれたstrokeが全部降ってくるが、それは無視する。
		if stroke.CreatedAt.After(startTime) && now.Sub(stroke.CreatedAt) > thresholdResponseTime {
			fmt.Println("response too late")
			w.Leave()
		}
		w.Logs = append(w.Logs, Log{ // TODO: mutex
			Time:       now,
			RoomID:     roomID,
			StrokeID:   stroke.ID,
			StrokeTime: stroke.CreatedAt,
		})
	})
	w.es.On("bad_request", func(data string) {
		fmt.Println("bad_request")
		fmt.Println(data)
		w.Leave()
	})
	w.es.On("watcher_count", func(data string) {
		fmt.Println("watcher_count")
		fmt.Println(data)
	})
	w.es.OnError(func(err error) {
		fmt.Println("error")

		if e, ok := err.(*sse.BadContentType); ok {
			fmt.Println("bad content type " + e.ContentType)
		}
		if e, ok := err.(*sse.BadStatusCode); ok {
			fmt.Printf("bad status code %d\n", e.StatusCode)
			if 400 <= e.StatusCode && e.StatusCode < 500 {
				w.Leave()
			}
		}
		fmt.Println(err)
	})
	w.es.OnEnd(func() {
		w.EndCh <- struct{}{}
	})

	w.es.Start()
}

func (w *RoomWatcher) Leave() {
	if w.es != nil {
		w.es.Close()
	}
}
