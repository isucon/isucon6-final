package scenario

import (
	"io"

	"github.com/PuerkitoBio/goquery"
	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/session"
)

func GetCSRFToken(s *session.Session, path string) (string, bool) {
	var token string

	ok := s.Get(path, func(body io.Reader, l *fails.Logger) bool {
		doc, err := goquery.NewDocumentFromReader(body)
		if err != nil {
			l.Add("ページのHTMLがパースできませんでした", err)
			return false
		}

		token = extractCsrfToken(doc)

		if token == "" {
			l.Add("トークンが取得できませんでした", nil)
			return false
		}

		return true
	})
	if !ok {
		return "", false
	}
	return token, true
}
