package scenario

import (
	"encoding/json"
	"io"
	"math/rand"
	"strconv"

	"github.com/catatsuy/isucon6-final/bench/action"
	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/seed"
	"github.com/catatsuy/isucon6-final/bench/session"
)

// トップページと画像に負荷をかける
func LoadIndexPage(origins []string) bool {
	s := session.New(randomOrigin(origins))
	defer s.Bye()

	images := []string{}
	for i := 0; i < 100; i++ {
		images = append(images, "/img/"+strconv.Itoa(1000-i))
	}

	ok := action.Get(s, "/", action.OK(func(body io.Reader, l *fails.Logger) bool {
		return true
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}
		images = extractImages(doc)
		return true
	}))
	if !ok {
		return false
	}

	// assetで失敗しても画像はリクエストかける
	ok = loadStaticFiles(s, false /*checkHash*/)
	if !ok {
		return false
	}

	ok = loadImages(s, images)
	if !ok {
		return false
	}
	return true
}

// /api/rooms にリクエストして、その中の一つの部屋を開いてstrokeをPOST
func DrawOnRandomRoom(origins []string) {
	s := session.New(randomOrigin(origins))
	defer s.Bye()

	var rooms []Room

	ok := action.Get(s, "/api/rooms", action.OK(func(body io.Reader, l *fails.Logger) bool {
		return true

		var res Response
		err := json.NewDecoder(body).Decode(&res)
		if err != nil {
			l.Add("レスポンスのJSONが読みとれませんでした", err)
			return false
		}
		if len(res.Rooms) != 100 {
			l.Add("部屋の数が100件になっていません: "+strconv.Itoa(len(res.Rooms)), nil)
			return false
		}
		rooms = res.Rooms
		return true
	}))
	if !ok {
		return
	}

	//room := rooms[rand.Intn(80)+20] // 上の方はスキップしてちょっと後ろの方を見ることにする

	room := Room{ID: 1000}

	roomURL := "/rooms/" + strconv.FormatInt(room.ID, 10)

	token, ok := fetchCSRFToken(s, roomURL)
	if !ok {
		return
	}

	seedStrokes := seed.GetStrokes("www")
	seedStroke := seed.FluctuateStroke(seedStrokes[rand.Intn(len(seedStrokes))])
	_, _ = drawStroke(s, token, room.ID, seedStroke)
}
