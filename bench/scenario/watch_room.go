package scenario

import (
	"github.com/catatsuy/isucon6-final/bench/session"
	"io"
	"strconv"
	"errors"
)

func GetCSRFTokenFromRoom(s *session.Session, roomID int) (string,error){
	var token string

	err := s.Get("/rooms/" + strconv.Itoa(roomID), func(status int, body io.Reader) error {
		if status != 200 {
			return errors.New("ステータスコードが200ではありません")
		}
		doc, err := makeDocument(body)
		if err != nil {
			return err
		}

		token = extractCsrfToken(doc)

		if token == "" {
			return errors.New("トークンが取得できませんでした")
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	return token, nil
}
