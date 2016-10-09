package scenario

import (
	"io"
	"math/rand"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/catatsuy/isucon6-final/bench/action"
	"github.com/catatsuy/isucon6-final/bench/fails"
)

var (
	StrokeReceiveScore int64 = 1
)

// トップページと画像に負荷をかける
func LoadIndexPage(origins []string) {
	s := newSession(origins)

	var token string
	var images []string

	ok := action.Get(s, "/", func(body io.Reader, l *fails.Logger) bool {
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}

		token, ok = extractCsrfToken(doc, l)
		if !ok {
			return false
		}

		images = extractImages(doc)
		if len(images) < 100 {
			l.Critical("画像の枚数が少なすぎます", nil)
			return false
		}

		return true
	})
	if !ok {
		return
	}

	loadImages(s, images)
}

// トップページを開いて適当な部屋を開く（Ajaxじゃないのは「別タブで」開いたということにでもしておく）
func LoadRoomPage(origins []string) {
	s := newSession(origins)

	var images []string
	var rooms []string

	ok := action.Get(s, "/", func(body io.Reader, l *fails.Logger) bool {
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}

		images = extractImages(doc)

		doc.Find("a").Each(func(_ int, selection *goquery.Selection) {
			if url, ok := selection.Attr("href"); ok {
				if strings.HasPrefix(url, "/rooms/") {
					rooms = append(rooms, url)
				}
			}
		})

		return true
	})
	if !ok {
		return
	}

	ok = loadImages(s, images)
	if !ok {
		return
	}

	roomURL := rooms[rand.Intn(len(rooms))]

	_ = action.Get(s, roomURL, func(body io.Reader, l *fails.Logger) bool {

		// TODO: polylineのidを上で開いたSVGと比較するか？

		return true
	})
}
