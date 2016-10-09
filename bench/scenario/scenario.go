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

var (
	StrokeReceiveScore int64 = 1
)

// TODO: ステータスコード以外にもチェックしたい
func loadImages(s *session.Session, images []string) bool {
	status := true
	for _, image := range images {
		ok := action.Get(s, image, func(body io.Reader, l *fails.Logger) bool {
			return false
		})
		status = status && ok
	}
	return status

	// TODO: 画像を並列リクエストするようにしてみたが、 connection reset by peer というエラーが出るので直列に戻した
	// もしかすると s.Transport.MaxIdleConnsPerHost ずつ処理するといけるのかも
	//errs := make(chan error, len(images))
	//for _, image := range images {
	//	go func(image string) {
	//		err := s.Get(image, func(status int, body io.Reader) error {
	//			if status != 200 {
	//				return errors.New("ステータスが200ではありません: " + strconv.Itoa(status))
	//			}
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

// ページ内のCSRFトークンが毎回変わっていることをチェック
func CheckCSRFTokenRefreshed(origins []string) {
	s := newSession(origins)

	token1, ok := fetchCSRFToken(s, "/")
	if !ok {
		return
	}

	token2, ok := fetchCSRFToken(s, "/")
	if !ok {
		return
	}

	if token1 == token2 {
		fails.Critical("csrf_tokenが使いまわされています", nil)
	}
}
