package scenario

import (
	"io"

	"github.com/PuerkitoBio/goquery"
	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/session"
)

func fetchCSRFToken(s *session.Session, path string) (string, bool) {
	var token string

	ok := s.Get(path, func(body io.Reader, l *fails.Logger) bool {
		doc, ok := makeDocument(body, l)
		if !ok {
			return false
		}

		token, ok = extractCsrfToken(doc, l)

		return ok
	})
	if !ok {
		return "", false
	}
	return token, true
}

func makeDocument(body io.Reader, l *fails.Logger) (*goquery.Document, bool) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		l.Add("ページのHTMLがパースできませんでした", err)
		return nil, false
	}
	return doc, true
}

func extractCsrfToken(doc *goquery.Document, l *fails.Logger) (string, bool) {
	token := ""

	doc.Find("html").Each(func(_ int, selection *goquery.Selection) {
		if t, ok := selection.Attr("data-csrf-token"); ok {
			token = t
		}
	})

	ok := token != ""
	if !ok {
		l.Add("トークンが取得できませんでした", nil)
	}

	return token, ok
}
