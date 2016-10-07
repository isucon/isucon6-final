package session

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"time"

	"strconv"

	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/http"
	"github.com/catatsuy/isucon6-final/bench/http/cookiejar"
	"github.com/catatsuy/isucon6-final/bench/score"
)

const (
	DefaultTimeout = time.Duration(10) * time.Second
	GetScore       = 1
	PostScore      = 20
)

type Session struct {
	Scheme    string
	Host      string
	UserAgent string
	Client    *http.Client
	Transport *http.Transport
}

type CheckFunc func(body io.Reader, l *fails.Logger) bool

func New(baseURL string) *Session {
	s := &Session{}

	s.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConnsPerHost: 6,
	}

	jar, _ := cookiejar.New(nil)

	s.Client = &http.Client{
		Transport: s.Transport,
		Jar:       jar,
		Timeout:   DefaultTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("redirect attempted")
		},
	}

	s.UserAgent = "benchmarker"

	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err) // should be cared at initialization
	}
	s.Scheme = u.Scheme
	s.Host = u.Host

	return s
}

func (s *Session) request(method, path string, body io.Reader, headers map[string]string, checkFunc CheckFunc) bool {
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

func (s *Session) Get(path string, checkFunc CheckFunc) bool {
	ok := s.request("GET", path, nil, nil, checkFunc)
	if ok {
		score.Increment(GetScore)
	}
	return ok
}

func (s *Session) Post(path string, body []byte, headers map[string]string, checkFunc CheckFunc) bool {
	ok := s.request("POST", path, bytes.NewBuffer(body), headers, checkFunc)
	if ok {
		score.Increment(PostScore)
	}
	return ok
}
