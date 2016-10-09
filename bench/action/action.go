package action

import (
	"bytes"
	"io"
	"net"
	"net/url"
	"strconv"

	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/http"
	"github.com/catatsuy/isucon6-final/bench/score"
	"github.com/catatsuy/isucon6-final/bench/session"
)

const (
	GetScore  = 1
	PostScore = 20
)

type CheckFunc func(body io.Reader, l *fails.Logger) bool

func request(s *session.Session, method, path string, body io.Reader, headers map[string]string, checkFunc CheckFunc) bool {
	l := &fails.Logger{Prefix: "[" + method + " " + path + "] "}

	u, err := url.Parse(path)
	if err != nil {
		return false
	}
	u.Scheme = s.Scheme
	u.Host = s.Host

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		l.Critical("予期せぬ失敗です (主催者に連絡してください)", err)
		return false
	}

	req.Header.Set("User-Agent", s.UserAgent)
	if headers != nil {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}

	res, err := s.Client.Do(req)

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			l.Add("リクエストがタイムアウトしました", err)
			return false
		}
		l.Add("リクエストが失敗しました", err)
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		l.Add("ステータスが200ではありません: "+strconv.Itoa(res.StatusCode), nil)
		return false
	}

	return checkFunc(res.Body, l)
}

func Get(s *session.Session, path string, checkFunc CheckFunc) bool {
	ok := request(s, "GET", path, nil, nil, checkFunc)
	if ok {
		score.Increment(GetScore)
	}
	return ok
}

func Post(s *session.Session, path string, body []byte, headers map[string]string, checkFunc CheckFunc) bool {
	ok := request(s, "POST", path, bytes.NewBuffer(body), headers, checkFunc)
	if ok {
		score.Increment(PostScore)
	}
	return ok
}
