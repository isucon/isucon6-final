package scenario

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/score"
	"github.com/catatsuy/isucon6-final/bench/seed"
	"github.com/catatsuy/isucon6-final/bench/session"
)

// 一人がroomを作る→大勢がそのroomをwatchする
func Matsuri(s *session.Session, aud string, timeoutCh chan struct{}) {
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
	}
	for _, strokeLog := range audRes.StrokeLogs {
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
