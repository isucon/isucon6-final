package scenario

import (
	"io"
	"strconv"

	"github.com/catatsuy/isucon6-final/bench/action"
	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/seed"
)

// 部屋を作って線を描くとトップページに出てくる
func StrokeReflectedToTop(origins []string) {
	s1 := newSession(origins)
	s2 := newSession(origins)

	token, ok := fetchCSRFToken(s1, "/")
	if !ok {
		return
	}

	roomID, ok := makeRoom(s1, token)
	if !ok {
		return
	}

	strokes := seed.GetStrokes("star")
	stroke := seed.FluctuateStroke(strokes[0])
	_, ok = drawStroke(s1, token, roomID, stroke)
	if !ok {
		return
	}

	// 描いた直後にトップページに表示される
	_ = action.Get(s2, "/", func(body io.Reader, l *fails.Logger) bool {
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
	})
}

// 線の描かれてない部屋はトップページに並ばない
func RoomWithoutStrokeNotShownAtTop(origins []string) {
	s1 := newSession(origins)
	s2 := newSession(origins)

	token, ok := fetchCSRFToken(s1, "/")
	if !ok {
		return
	}

	roomID, ok := makeRoom(s1, token)
	if !ok {
		return
	}

	_ = action.Get(s2, "/", func(body io.Reader, l *fails.Logger) bool {
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
	})
}

// 線がSVGに反映される
func StrokeReflectedToSVG(origins []string) {
	s1 := newSession(origins)

	token, ok := fetchCSRFToken(s1, "/")
	if !ok {
		return
	}

	roomID, ok := makeRoom(s1, token)
	if !ok {
		return
	}

	strokes := seed.GetStrokes("wwws")
	for _, stroke := range strokes {
		stroke2 := seed.FluctuateStroke(stroke)
		strokeID, ok := drawStroke(s1, token, roomID, stroke2)
		if !ok {
			return
		}

		s2 := newSession(origins)
		ok = checkStrokeReflectedToSVG(s2, roomID, strokeID, stroke2)
	}
}
