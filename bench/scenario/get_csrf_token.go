package scenario

import (
	"errors"
	"io"

	"github.com/catatsuy/isucon6-final/bench/session"
	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/PuerkitoBio/goquery"
)

func GetCSRFToken(s *session.Session, path string) (string, error) {
	var token string

	err := s.Get(path, func(body io.Reader, l *fails.Logger) error {
		doc, err := goquery.NewDocumentFromReader(body)
		if err != nil {
			l.Add("ページのHTMLがパースできませんでした", err)
			return err
		}

		token = extractCsrfToken(doc)

		if token == "" {
			l.Add("トークンが取得できませんでした", nil)
			return errors.New("bad csrf_token")
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	return token, nil
}
