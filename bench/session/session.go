package session

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/http"
	"github.com/catatsuy/isucon6-final/bench/http/cookiejar"
)

const DefaultTimeout = time.Duration(10) * time.Second

type Session struct {
	Scheme    string
	Host      string
	UserAgent string
	Client    *http.Client
	Transport *http.Transport
}

type CheckFunc func(status int, body io.Reader) error // TODO: Headerも受け取る？

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

func (s *Session) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	u.Scheme = s.Scheme
	u.Host = s.Host
	return http.NewRequest(method, u.String(), body)
}

func (s *Session) Get(path string, checkFunc CheckFunc) error {

	req, err := s.NewRequest("GET", path, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "GET "+path+", error:"+err.Error())
		fails.Add("GET " + path + ", 予期せぬ失敗です (主催者に連絡してください)")
		return err
	}

	req.Header.Set("User-Agent", s.UserAgent)

	res, err := s.Client.Do(req)

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fails.Add("GET " + path + ", リクエストがタイムアウトしました")
			return err
		}
		fmt.Fprintln(os.Stderr, "GET "+path+", error:"+err.Error())
		fails.Add("GET " + path + ", リクエストに失敗しました")
		return err
	}
	defer res.Body.Close()

	err = checkFunc(res.StatusCode, res.Body)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) Post(path string, body []byte, headers map[string]string, checkFunc CheckFunc) error {

	req, err := s.NewRequest("POST", path, bytes.NewBuffer(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "POST "+path+", error:"+err.Error())
		fails.Add("POST " + path + ", 予期せぬ失敗です (主催者に連絡してください)")
		return err
	}

	req.Header.Set("User-Agent", s.UserAgent)
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	res, err := s.Client.Do(req)

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fails.Add("リクエストがタイムアウトしました")
			return err
		}
		fmt.Fprintln(os.Stderr, "POST "+path+", error:"+err.Error())
		fails.Add("POST " + path + ", リクエストに失敗しました")
		return err
	}
	defer res.Body.Close()

	err = checkFunc(res.StatusCode, res.Body)
	if err != nil {
		return err
	}
	return nil
}
