package scenario

import (
	"errors"
	"io"

	"github.com/catatsuy/isucon6-final/bench/session"
)

func GetCSRFToken(s *session.Session, path string) (string, error) {
	var token string

	err := s.Get(path, func(status int, body io.Reader) error {
		if status != 200 {
			return errors.New("GET " + path + ", ステータスコードが200ではありません")
		}
		doc, err := makeDocument(body)
		if err != nil {
			return errors.New("GET " + path + ", HTMLが正しくありません")
		}

		token = extractCsrfToken(doc)

		if token == "" {
			return errors.New("GET " + path + ", トークンが取得できませんでした")
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	return token, nil
}
