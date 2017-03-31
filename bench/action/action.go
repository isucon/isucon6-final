package action

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/sesta/isucon6-final/bench/fails"
	"github.com/sesta/isucon6-final/bench/score"
	"github.com/sesta/isucon6-final/bench/session"
	"github.com/sesta/isucon6-final/bench/sse"
)

const (
	GetScore  = 1
	PostScore = 20
)

type Checker interface {
	Check(body io.Reader, l *fails.Logger) bool
	CheckStatus(status int, l *fails.Logger) bool
}

type CheckFunc func(body io.Reader, l *fails.Logger) bool

type StatusChecker struct {
	ExpectedStatus int
	CheckFunc      CheckFunc
}

func (sc StatusChecker) Check(body io.Reader, l *fails.Logger) bool {
	return sc.CheckFunc(body, l)
}

func (sc StatusChecker) CheckStatus(status int, l *fails.Logger) bool {
	if status != sc.ExpectedStatus {
		l.Add(fmt.Sprintf("ステータスが%dではありません: %d", sc.ExpectedStatus, status), nil)
		return false
	}
	return true
}

func OK(f CheckFunc) StatusChecker {
	return StatusChecker{
		ExpectedStatus: 200,
		CheckFunc:      f,
	}
}

func BadRequest(f CheckFunc) StatusChecker {
	return StatusChecker{
		ExpectedStatus: 400,
		CheckFunc:      f,
	}
}

func request(s *session.Session, method, path string, body io.Reader, headers map[string]string, c Checker) bool {
	l := &fails.Logger{Prefix: "[" + method + " " + path + "] "}

	u, err := url.Parse(path)
	if err != nil {
		l.Critical("予期せぬエラー（主催者に連絡してください）",
			errors.New("URLのパースに失敗しました: "+path+", error: "+err.Error()))
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

	ok := c.CheckStatus(res.StatusCode, l)
	if !ok {
		return false
	}

	return c.Check(res.Body, l)
}

func Get(s *session.Session, path string, c Checker) bool {
	ok := request(s, "GET", path, nil, nil, c)
	if ok {
		score.Increment(GetScore)
	}
	return ok
}

func Post(s *session.Session, path string, body []byte, headers map[string]string, c Checker) bool {
	ok := request(s, "POST", path, bytes.NewBuffer(body), headers, c)
	if ok {
		score.Increment(PostScore)
	}
	return ok
}

func SSE(s *session.Session, path string) (*sse.EventSource, bool) {
	u, err := url.Parse(path)
	if err != nil {
		fails.Critical("予期せぬエラー（主催者に連絡してください）",
			errors.New("URLのパースに失敗しました: "+path+", error: "+err.Error()))
		return nil, false
	}
	u.Scheme = s.Scheme
	u.Host = s.Host

	es := sse.NewEventSource(s.Client, u.String())
	es.AddHeader("User-Agent", s.UserAgent)
	return es, true
}
