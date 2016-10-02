package scenario

import (
	"encoding/json"
	"errors"
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
	"github.com/catatsuy/isucon6-final/bench/stderr"
)

// 一人がroomを作る→大勢がそのroomをwatchする
func Matsuri(s *session.Session, aud string, timeoutCh chan struct{}) {
	var token string

	err := s.Get("/", func(status int, body io.Reader) error {
		if status != 200 {
			fails.Critical()
			return errors.New("ステータスが200ではありません: " + strconv.Itoa(status))
		}
		doc, err := makeDocument(body)
		if err != nil {
			return err
		}

		token = extractCsrfToken(doc)

		if token == "" {
			fails.Critical()
			return errors.New("csrf_tokenが取得できませんでした")
		}

		score.Increment(IndexGetScore)

		return nil
	})
	if err != nil {
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

	err = s.Post("/api/rooms", postBody, headers, func(status int, body io.Reader) error {
		if status != 200 {
			fails.Critical()
			return errors.New("ステータスが200ではありません: " + strconv.Itoa(status))
		}

		var res Response
		err := json.NewDecoder(body).Decode(&res)
		if err != nil {
			fails.Critical()
			stderr.Log.Println(err.Error())
			return errors.New("レスポンス内容が正しくありません")
		}
		if res.Room == nil || res.Room.ID <= 0 {
			fails.Critical()
			return errors.New("レスポンス内容が正しくありません")
		}
		RoomID = res.Room.ID

		score.Increment(CreateRoomScore)

		return nil
	})

	if err != nil {
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

				err := s.Post("/api/strokes/rooms/"+strconv.FormatInt(RoomID, 10), postBody, headers, func(status int, body io.Reader) error {
					responseTime := time.Now()

					if status != 200 {
						fails.Critical()
						return errors.New("ステータスが200ではありません: " + strconv.Itoa(status))
					}

					var res Response
					err = json.NewDecoder(body).Decode(&res)
					if err != nil {
						fails.Critical()
						stderr.Log.Println(err.Error())
						return errors.New("レスポンス内容が正しくありません")
					}
					if res.Stroke == nil || res.Stroke.ID <= 0 {
						fails.Critical()
						return errors.New("レスポンス内容が正しくありません")
					}

					timeTaken := responseTime.Sub(postTime).Seconds()
					if timeTaken < 1 { // TODO: この時間は要調整
						score.Increment(CreateStrokeScore * 2)
					} else if timeTaken < 3 {
						score.Increment(CreateStrokeScore)
					}

					postTimes[res.Stroke.ID] = postTime

					return nil
				})
				if err != nil || len(timeoutCh) > 0 {
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
		stderr.Log.Println("failed to call audience " + aud + " :" + err.Error())
		fails.Add("予期せぬエラー (主催者に連絡してください)")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			stderr.Log.Println("failed to call audience " + aud)
		} else {
			stderr.Log.Println("failed to call audience " + aud + " :" + string(body))
		}
		fails.Add("予期せぬエラー (主催者に連絡してください)")
		return
	}

	var audRes AudienceResponse
	err = json.NewDecoder(resp.Body).Decode(&audRes)
	if err != nil {
		stderr.Log.Println("failed to decode json from audience: " + err.Error())
		fails.Add("予期せぬエラー (主催者に連絡してください)")
		return
	}
	for _, msg := range audRes.Errors {
		fails.Add(msg)
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
