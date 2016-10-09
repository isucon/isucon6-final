package scenario

import (
	"io"
	"math/rand"

	"fmt"
	"io/ioutil"
	"math"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/catatsuy/isucon6-final/bench/action"
	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/seed"
	"github.com/catatsuy/isucon6-final/bench/session"
	"github.com/catatsuy/isucon6-final/bench/svg"
)

func newSession(origins []string) *session.Session {
	return session.New(origins[rand.Intn(len(origins))])
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
					l.Critical("投稿が反映されていません（pointが足りません）", err)
					return false
				}
				for j, p := range polyLine.Points {
					if math.Abs(float64(stroke.Points[j].X)-float64(p.X)) > 0.1 || math.Abs(float64(stroke.Points[j].Y)-float64(p.Y)) > 0.1 {
						fmt.Println(stroke.Points[j].X, p.X, stroke.Points[j].Y, p.Y)
						l.Critical("投稿が反映されていません（x,yの値が改変されています）", err)
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
