package scenario

import (
	"errors"
	"io"
	"strconv"

	"encoding/json"

	"fmt"
	"os"

	"net/url"

	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/http"
	"github.com/catatsuy/isucon6-final/bench/score"
	"github.com/catatsuy/isucon6-final/bench/seed"
	"github.com/catatsuy/isucon6-final/bench/session"
)

var (
	IndexGetScore      int64 = 2
	SVGGetScore        int64 = 1
	CreateRoomScore    int64 = 20
	CreateStrokeScore  int64 = 20
	StrokeReceiveScore int64 = 1
)

func makeDocument(r io.Reader) (*goquery.Document, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.New(fails.Add("ページのHTMLがパースできませんでした"))
	}
	return doc, nil
}

func extractImages(doc *goquery.Document) []string {
	imageUrls := []string{}

	doc.Find("img").Each(func(_ int, selection *goquery.Selection) {
		if url, ok := selection.Attr("src"); ok {
			imageUrls = append(imageUrls, url)
		}
	})

	return imageUrls
}

func extractCsrfToken(doc *goquery.Document) string {
	var token string

	doc.Find("html").Each(func(_ int, selection *goquery.Selection) {
		if t, ok := selection.Attr("data-csrf-token"); ok {
			token = t
		}
	})

	return token
}

func loadImages(s *session.Session, images []string) error {
	var lastErr error
	for _, image := range images {
		err := s.Get(image, func(status int, body io.Reader) error {
			if status != 200 {
				return errors.New(fails.Add("GET /, ステータスが200ではありません: " + strconv.Itoa(status)))
			}
			score.Increment(SVGGetScore)
			return nil
		})
		if err != nil {
			lastErr = err
		}
	}
	return lastErr

	// TODO: 画像を並列リクエストするようにしてみたが、 connection reset by peer というエラーが出るので直列に戻した
	// もしかすると s.Transport.MaxIdleConnsPerHost ずつ処理するといけるのかも
	//errs := make(chan error, len(images))
	//for _, image := range images {
	//	go func(image string) {
	//		err := s.Get(image, func(status int, body io.Reader) error {
	//			if status != 200 {
	//				return errors.New(fails.Add("GET /, ステータスが200ではありません: " + strconv.Itoa(status)))
	//			}
	//			score.Increment(SVGGetScore)
	//			return nil
	//		})
	//		errs <- err
	//	}(image)
	//}
	//var lastErr error
	//for i := 0; i < len(images); i++ {
	//	err := <-errs
	//	if err != nil {
	//		lastErr = err
	//	}
	//}
	//return lastErr
}

// トップページと画像に負荷をかける
func LoadIndexPage(s *session.Session) {
	var token string
	var images []string

	err := s.Get("/", func(status int, body io.Reader) error {
		if status != 200 {
			return errors.New(fails.Add("GET /, ステータスが200ではありません: " + strconv.Itoa(status)))
		}
		doc, err := makeDocument(body)
		if err != nil {
			return err
		}

		token = extractCsrfToken(doc)

		if token == "" {
			return errors.New(fails.Add("GET /, csrf_tokenが取得できませんでした"))
		}

		images = extractImages(doc)
		if len(images) < 100 {
			return errors.New(fails.Add("GET /, 画像の枚数が少なすぎます"))
		}

		score.Increment(IndexGetScore)

		return nil
	})
	if err != nil {
		return
	}

	err = loadImages(s, images)
	if err != nil {
		return
	}
}

// ページ内のCSRFトークンが毎回変わっていることをチェック
func CheckCSRFTokenRefreshed(s *session.Session) {
	var token string

	err := s.Get("/", func(status int, body io.Reader) error {
		if status != 200 {
			return errors.New(fails.Add("GET /, ステータスが200ではありません: " + strconv.Itoa(status)))
		}
		doc, err := makeDocument(body)
		if err != nil {
			return err
		}

		token = extractCsrfToken(doc)

		if token == "" {
			return errors.New(fails.Add("GET /, csrf_tokenが取得できませんでした"))
		}

		score.Increment(IndexGetScore)

		return nil
	})
	if err != nil {
		return
	}

	_ = s.Get("/", func(status int, body io.Reader) error {
		if status != 200 {
			return errors.New(fails.Add("GET /, ステータスが200ではありません: " + strconv.Itoa(status)))
		}
		doc, err := makeDocument(body)
		if err != nil {
			return err
		}

		t := extractCsrfToken(doc)

		if t == token {
			return errors.New(fails.Add("GET /, csrf_tokenが使いまわされています"))
		}

		score.Increment(IndexGetScore)

		return nil
	})
}

// 一人がroomを作る→大勢がそのroomをwatchする
func MatsuriRoom(s *session.Session, aud string) {
	var token string

	err := s.Get("/", func(status int, body io.Reader) error {
		if status != 200 {
			return errors.New(fails.Add("GET /, ステータスが200ではありません: " + strconv.Itoa(status)))
		}
		doc, err := makeDocument(body)
		if err != nil {
			return err
		}

		token = extractCsrfToken(doc)

		if token == "" {
			return errors.New(fails.Add("GET /, csrf_tokenが取得できませんでした"))
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
			return errors.New(fails.Add("GET /api/rooms, ステータスが200ではありません: " + strconv.Itoa(status)))
		}

		var res Response
		err := json.NewDecoder(body).Decode(&res)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errors.New(fails.Add("GET /api/rooms, レスポンス内容が正しくありません"))
		}
		if res.Room == nil || res.Room.ID <= 0 {
			return errors.New(fails.Add("GET /api/rooms, レスポンス内容が正しくありません"))
		}
		RoomID = res.Room.ID

		score.Increment(CreateRoomScore)

		return nil
	})

	if err != nil {
		return
	}

	// TODO: strokeを順次postしていく
	seedStroke := seed.GetStroke("main001")

	postTimes := make(map[int64]time.Time)

	end := make(chan struct{})

	go func() {
		for _, str := range seedStroke {
			// TODO: 指定時間以上たったら終わる

			postBody, _ := json.Marshal(struct {
				RoomID int64 `json:"room_id"`
				seed.Stroke
			}{
				RoomID: RoomID,
				Stroke: str,
			})

			postTime := time.Now()

			err := s.Post("/api/strokes/rooms/"+strconv.FormatInt(RoomID, 10), postBody, headers, func(status int, body io.Reader) error {
				responseTime := time.Now()

				if status != 200 {
					return errors.New(fails.Add("POST /api/strokes/rooms/" + strconv.FormatInt(RoomID, 10) + ", ステータスが200ではありません: " + strconv.Itoa(status)))
				}

				var res Response
				err = json.NewDecoder(body).Decode(&res)
				if err != nil || res.Stroke == nil || res.Stroke.ID <= 0 {
					return errors.New(fails.Add("POST /api/strokes/rooms/" + strconv.FormatInt(RoomID, 10) + ", レスポンス内容が正しくありません"))
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
			if err != nil {
				break
			}
		}
		end <- struct{}{}
	}()

	resp, err := http.Get(aud + "?scheme=" + url.QueryEscape(s.Scheme) + "&host=" + url.QueryEscape(s.Host) + "&room=" + strconv.FormatInt(RoomID, 10))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to call audience "+aud+" :"+err.Error())
	}
	defer resp.Body.Close()

	var audRes AudienceResponse
	err = json.NewDecoder(resp.Body).Decode(&audRes)
	if err != nil {
		fmt.Println(err.Error())
		// TODO: 主催者に連絡してください的なエラーを出す
		return
	}
	for _, e := range audRes.Errors {
		fmt.Println(e) // TODO: 単純にfails.Addしてしまってよいか？
	}
	for _, l := range audRes.StrokeLogs {
		postTime := postTimes[l.StrokeID]
		timeTaken := l.ReceivedTime.Sub(postTime).Seconds()
		if timeTaken < 1 { // TODO: この時間は要調整
			score.Increment(StrokeReceiveScore * 2)
		} else if timeTaken < 3 {
			score.Increment(StrokeReceiveScore)
		}
	}

	<-end
}
