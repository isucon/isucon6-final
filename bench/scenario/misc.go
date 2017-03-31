package scenario

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/sesta/isucon6-final/bench/action"
	"github.com/sesta/isucon6-final/bench/fails"
	"github.com/sesta/isucon6-final/bench/seed"
	"github.com/sesta/isucon6-final/bench/session"
	"github.com/sesta/isucon6-final/bench/svg"
)

func randomOrigin(origins []string) string {
	return origins[rand.Intn(len(origins))]
}

func fetchCSRFToken(s *session.Session, path string) (string, bool) {
	var token string

	ok := action.Get(s, path, action.OK(func(body io.Reader, l *fails.Logger) bool {
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}

		token, ok = extractCsrfToken(doc, l)

		return ok
	}))
	if !ok {
		return "", false
	}
	return token, true
}

func makeDocument(body io.Reader, l *fails.Logger) (*goquery.Document, bool) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		l.Add("ページのHTMLがパースできませんでした", err)
		return nil, false
	}
	return doc, true
}

func parseResponseJSON(body io.Reader, l *fails.Logger) (*Response, bool) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		l.Add("レスポンス内容が読み込めませんでした", err)
		return nil, false
	}
	var res Response
	err = json.Unmarshal(b, &res)
	if err != nil {
		l.Add("レスポンスのJSONがパースできませんでした: "+string(b[:20]), err)
		return nil, false
	}
	return &res, true
}

func extractCsrfToken(doc *goquery.Document, l *fails.Logger) (string, bool) {
	token := ""

	doc.Find("html").Each(func(_ int, selection *goquery.Selection) {
		if t, ok := selection.Attr("data-csrf-token"); ok {
			token = t
		}
	})

	ok := token != ""
	if !ok {
		l.Add("トークンが取得できませんでした", nil)
	}

	return token, ok
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

// 描いた線がsvgに反映されるか
func checkStrokeReflectedToSVG(s *session.Session, roomID int64, strokeID int64, stroke seed.Stroke) bool {
	imageURL := "/img/" + strconv.FormatInt(roomID, 10)

	return action.Get(s, imageURL, action.OK(func(body io.Reader, l *fails.Logger) bool {
		b, err := ioutil.ReadAll(body)
		if err != nil {
			l.Critical("内容が読み込めませんでした", err)
			return false
		}
		data, err := svg.Parse(b)
		if err != nil {
			l.Critical("SVGがパースできませんでした", err)
			return false
		}
		for i, polyLine := range data.PolyLines {
			if data.PolyLines[i].ID == strconv.FormatInt(strokeID, 10) {
				if len(stroke.Points) != len(polyLine.Points) {
					l.Critical("投稿が反映されていません（pointが足りません）", nil)
					return false
				}
				for j, p := range polyLine.Points {
					if math.Abs(float64(stroke.Points[j].X)-float64(p.X)) > 0.1 || math.Abs(float64(stroke.Points[j].Y)-float64(p.Y)) > 0.1 {
						fmt.Println(stroke.Points[j].X, p.X, stroke.Points[j].Y, p.Y)
						l.Critical("投稿が反映されていません（x,yの値が改変されています）", nil)
						return false
					}
				}
				return true
			}
		}
		// ここに来るのは、IDがstroke.IDと同じpolylineが一つも無かったとき
		l.Critical("投稿が反映されていません", err)
		return false
	}))
}

func loadImages(s *session.Session, images []string) bool {
	ch := make(chan struct{}, session.MaxIdleConnsPerHost)
	OK := true
	for _, image := range images {
		ch <- struct{}{}
		go func(image string) {
			ok := action.Get(s, image, action.OK(func(body io.Reader, l *fails.Logger) bool {
				return true
			}))
			if !ok {
				OK = false // ture -> false になるだけなのでmutexは不要と思われ
			}
			<-ch
		}(image)
	}
	return OK
}

func makeRoom(s *session.Session, token string) (*Room, bool) {
	postBody, _ := json.Marshal(struct {
		Name         string `json:"name"`
		CanvasWidth  int    `json:"canvas_width"`
		CanvasHeight int    `json:"canvas_height"`
	}{
		Name:         "ひたすら椅子を描く部屋【" + strconv.Itoa(rand.Intn(1000)+1000) + "】",
		CanvasWidth:  1024,
		CanvasHeight: 768,
	})

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-csrf-token": token,
	}

	var room *Room

	ok := action.Post(s, "/api/rooms", postBody, headers, action.OK(func(body io.Reader, l *fails.Logger) bool {
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
		room = res.Room

		return true
	}))

	return room, ok
}

func drawStroke(s *session.Session, token string, roomID int64, seedStroke seed.Stroke) (*Stroke, bool) {
	postBody, _ := json.Marshal(struct {
		RoomID int64 `json:"room_id"`
		seed.Stroke
	}{
		RoomID: roomID,
		Stroke: seedStroke,
	})

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-csrf-token": token,
	}

	var stroke *Stroke

	u := "/api/strokes/rooms/" + strconv.FormatInt(roomID, 10)
	ok := action.Post(s, u, postBody, headers, action.OK(func(body io.Reader, l *fails.Logger) bool {

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

		stroke = res.Stroke

		return true
	}))

	return stroke, ok
}

func getRoomsAPI(s *session.Session) ([]Room, bool) {
	var rooms []Room

	ok := action.Get(s, "/api/rooms", action.OK(func(body io.Reader, l *fails.Logger) bool {
		res, ok := parseResponseJSON(body, l)
		if !ok {
			return false
		}
		if len(res.Rooms) != 100 {
			l.Add("部屋の数が100件になっていません: "+strconv.Itoa(len(res.Rooms)), nil)
			return false
		}
		rooms = res.Rooms
		return true
	}))

	return rooms, ok
}

func getRoomAPI(s *session.Session, roomID int64) (*Room, bool) {
	var room *Room

	roomAPIURL := "/api/rooms/" + strconv.FormatInt(roomID, 10)
	ok := action.Get(s, roomAPIURL, action.OK(func(body io.Reader, l *fails.Logger) bool {
		res, ok := parseResponseJSON(body, l)
		if !ok {
			return false
		}
		if res.Room == nil || len(res.Room.Strokes) == 0 {
			l.Add("レスポンス内容が正しくありません", nil)
			return false
		}
		room = res.Room
		return true
	}))

	return room, ok
}
