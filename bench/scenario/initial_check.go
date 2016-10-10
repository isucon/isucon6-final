package scenario

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/catatsuy/isucon6-final/bench/action"
	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/seed"
	"github.com/catatsuy/isucon6-final/bench/session"
)

// 部屋を作って線を描くとトップページに出てくる
func StrokeReflectedToTop(origins []string) {
	s1 := session.New(randomOrigin(origins))
	s2 := session.New(randomOrigin(origins))
	defer s1.Bye()
	defer s2.Bye()

	token, ok := fetchCSRFToken(s1, "/")
	if !ok {
		return
	}

	roomID, ok := makeRoom(s1, token)
	if !ok {
		fails.Critical("部屋の作成に失敗しました", nil)
		return
	}

	strokes := seed.GetStrokes("star")
	stroke := seed.FluctuateStroke(strokes[0])
	_, ok = drawStroke(s1, token, roomID, stroke)
	if !ok {
		fails.Critical("線の投稿に失敗しました", nil)
		return
	}

	// 描いた直後にトップページに表示される
	_ = action.Get(s2, "/", action.OK(func(body io.Reader, l *fails.Logger) bool {
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}
		imageUrls := extractImages(doc)

		found := false
		for _, img := range imageUrls {
			if img == "/img/"+strconv.FormatInt(roomID, 10) {
				found = true
			}
		}
		if !found {
			l.Critical("投稿が反映されていません", nil)
			return false
		}
		return true
	}))
}

// 線の描かれてない部屋はトップページに並ばない
func RoomWithoutStrokeNotShownAtTop(origins []string) {
	s1 := session.New(randomOrigin(origins))
	s2 := session.New(randomOrigin(origins))
	defer s1.Bye()
	defer s2.Bye()

	token, ok := fetchCSRFToken(s1, "/")
	if !ok {
		return
	}

	roomID, ok := makeRoom(s1, token)
	if !ok {
		fails.Critical("部屋の作成に失敗しました", nil)
		return
	}

	_ = action.Get(s2, "/", action.OK(func(body io.Reader, l *fails.Logger) bool {
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}
		imageUrls := extractImages(doc)

		for _, img := range imageUrls {
			if img == "/img/"+strconv.FormatInt(roomID, 10) {
				l.Critical("まだ線の無い部屋が表示されています", nil)
				return false
			}
		}
		return true
	}))
}

// 線がSVGに反映される
func StrokeReflectedToSVG(origins []string) {
	s1 := session.New(randomOrigin(origins))
	defer s1.Bye()

	token, ok := fetchCSRFToken(s1, "/")
	if !ok {
		return
	}

	roomID, ok := makeRoom(s1, token)
	if !ok {
		fails.Critical("部屋の作成に失敗しました", nil)
		return
	}

	strokes := seed.GetStrokes("wwws")
	for _, stroke := range strokes {
		stroke2 := seed.FluctuateStroke(stroke)
		strokeID, ok := drawStroke(s1, token, roomID, stroke2)
		if !ok {
			fails.Critical("線の投稿に失敗しました", nil)
			return
		}

		s2 := session.New(randomOrigin(origins))
		_ = checkStrokeReflectedToSVG(s2, roomID, strokeID, stroke2)
		s2.Bye()
	}
}

// ページ内のCSRFトークンが毎回変わっている
func CSRFTokenRefreshed(origins []string) {
	s1 := session.New(randomOrigin(origins))
	s2 := session.New(randomOrigin(origins))
	defer s1.Bye()
	defer s2.Bye()

	token1, ok := fetchCSRFToken(s1, "/")
	if !ok {
		return
	}

	token2, ok := fetchCSRFToken(s2, "/")
	if !ok {
		return
	}

	if token1 == token2 {
		fails.Critical("csrf_tokenが使いまわされています", nil)
	}
}

// 他人の作った部屋に最初の線を描けない
func CantDrawFirstStrokeOnSomeoneElsesRoom(origins []string) {
	s1 := session.New(randomOrigin(origins))
	s2 := session.New(randomOrigin(origins))
	defer s1.Bye()
	defer s2.Bye()

	token1, ok := fetchCSRFToken(s1, "/")
	if !ok {
		return
	}

	roomID, ok := makeRoom(s1, token1)
	if !ok {
		fails.Critical("部屋の作成に失敗しました", nil)
		return
	}

	token2, ok := fetchCSRFToken(s2, "/")
	if !ok {
		return
	}

	strokes := seed.GetStrokes("star")
	stroke := seed.FluctuateStroke(strokes[0])

	postBody, _ := json.Marshal(struct {
		RoomID int64 `json:"room_id"`
		seed.Stroke
	}{
		RoomID: roomID,
		Stroke: stroke,
	})

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-csrf-token": token2,
	}

	u := "/api/strokes/rooms/" + strconv.FormatInt(roomID, 10)
	ok = action.Post(s2, u, postBody, headers, action.BadRequest(func(body io.Reader, l *fails.Logger) bool {
		// JSONも検証する？
		return true
	}))
}

// トップページの内容が正しいかをチェック
func TopPageContent(origins []string) {
	s := session.New(randomOrigin(origins))
	defer s.Bye()

	_ = action.Get(s, "/", action.OK(func(body io.Reader, l *fails.Logger) bool {
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}
		images := extractImages(doc)
		if len(images) < 100 {
			l.Critical("画像の枚数が少なすぎます", nil)
			return false
		}

		reactidNum := doc.Find("[data-reactid]").Length()
		expected := 1325
		if reactidNum != expected {
			l.Critical("トップページの内容が正しくありません",
				fmt.Errorf("data-reactidの数が一致しません (expected %d, actual %d)", expected, reactidNum))
			return false
		}
		return true
	}))
}

// 静的ファイルが正しいかをチェック
func CheckAssets(origins []string) {
	s := session.New(randomOrigin(origins))
	defer s.Bye()

	ok := loadAssets(s, true /*checkHash*/)
	if !ok {
		fails.Critical("静的ファイルが正しくありません", nil)
	}
}
