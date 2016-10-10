package scenario

import (
	"io"
	"math/rand"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/catatsuy/isucon6-final/bench/action"
	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/session"
)

// トップページと画像に負荷をかける
func LoadIndexPage(origins []string) {
	s := session.New(randomOrigin(origins))
	defer s.Bye()

	var images []string

	ok := action.Get(s, "/", action.OK(func(body io.Reader, l *fails.Logger) bool {
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}
		images = extractImages(doc)
		return true
	}))
	if !ok {
		return
	}

	// assetで失敗しても画像はリクエストかける
	_ = loadAssets(s, false /*checkHash*/)

	_ = loadImages(s, images)
}

// トップページを開いて適当な部屋を開く（Ajaxじゃないのは「別タブで」開いたということにでもしておく）
func LoadRoomPage(origins []string) {
	s := session.New(randomOrigin(origins))
	defer s.Bye()

	var images []string
	var rooms []string

	ok := action.Get(s, "/", action.OK(func(body io.Reader, l *fails.Logger) bool {
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
	}))
	if !ok {
		return
	}

	ok = loadImages(s, images)
	if !ok {
		return
	}

	roomURL := rooms[rand.Intn(len(rooms))]

	_ = action.Get(s, roomURL, action.OK(func(body io.Reader, l *fails.Logger) bool {

		// TODO: polylineのidを上で開いたSVGと比較するか？

		return true
	}))
}
