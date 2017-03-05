package session

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/http2"
)

const (
	DefaultTimeout      = time.Duration(5) * time.Second
	MaxIdleConnsPerHost = 6
)

type Session struct {
	Scheme    string
	Host      string
	UserAgent string
	Client    *http.Client
	Transport *http.Transport
}

func New(baseURL string) *Session {
	s := &Session{}

	s.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
	}

	// TLSClientConfigを渡すとhttp2が使えないので、強制的にhttp2も使えるようにする
	// Transportを外から渡すことでコネクションを共有させない
	if err := http2.ConfigureTransport(s.Transport); err != nil {
		panic(fmt.Errorf("Failed to configure h2 transport: %s", err))
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

func (s *Session) Bye() {
	s.Transport.CloseIdleConnections()
}
