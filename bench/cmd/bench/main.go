package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"net/url"
	"strings"

	"errors"
	"os"

	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/scenario"
	"github.com/catatsuy/isucon6-final/bench/score"
	"github.com/catatsuy/isucon6-final/bench/session"
)

var BenchmarkTimeout int
var Audience1 string

func main() {

	var urls string

	flag.StringVar(&urls, "urls", "", "ベンチマーク対象のURL（scheme, host, portまで。カンマ区切りで複数可。例： https://xxx.xxx.xxx.xxx,https://xxx.xxx.xxx.xxx:1443）")
	flag.StringVar(&Audience1, "audience1", "", "オーディエンスAPIのURLその1 (http://xxx.xxx.xxx.xxx/)")
	flag.IntVar(&BenchmarkTimeout, "timeout", 60, "ソフトタイムアウト")

	flag.Parse()

	origins, err := makeOrigins(urls)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}

	// 初期チェックで失敗したらそこで終了
	initialCheck(origins)
	if len(fails.Get()) > 0 {
		output()
		return
	}

	benchmark(origins)
	output()
}

func makeOrigins(urls string) ([]string, error) {
	if urls == "" {
		return nil, errors.New("urlsが指定されていません")
	}
	origins := strings.Split(urls, ",")
	for _, origin := range origins {
		u, err := url.Parse(origin)
		if err != nil {
			return nil, err
		}
		if u.Scheme != "https" || u.Host == "" {
			return nil, errors.New("urlsの指定が間違っています")
		}
	}
	return origins, nil
}

func initialCheck(origins []string) {
	scenario.CheckCSRFTokenRefreshed(session.New(origins[0]))
}

func benchmark(origins []string) {
	loadIndexPageCh := makeChan(2)
	loadRoomPageCh := makeChan(2)
	checkCSRFTokenRefreshedCh := makeChan(1)
	matsuriCh := makeChan(1)
	matsuriEndCh := make(chan struct{})
	matsuriTimeoutCh := make(chan struct{}, 2) // http://mattn.kaoriya.net/software/lang/go/20160706165757.htm

	timeoutCh := time.After(time.Duration(BenchmarkTimeout) * time.Second)

L:
	for {
		select {
		case <-loadIndexPageCh:
			go func() {
				scenario.LoadIndexPage(session.New(origins[0]))
				loadIndexPageCh <- struct{}{}
			}()
		case <-loadRoomPageCh:
			go func() {
				scenario.LoadRoomPage(session.New(origins[0]))
				loadRoomPageCh <- struct{}{}
			}()
		case <-checkCSRFTokenRefreshedCh:
			go func() {
				scenario.CheckCSRFTokenRefreshed(session.New(origins[0]))
				checkCSRFTokenRefreshedCh <- struct{}{}
			}()
		case <-matsuriCh:
			go func() {
				scenario.Matsuri(session.New(origins[0]), Audience1, matsuriTimeoutCh)
				//matsuriRoomCh <- struct{}{} // Never again.
				matsuriEndCh <- struct{}{}
			}()
		case <-timeoutCh:
			break L
		}
	}
	matsuriTimeoutCh <- struct{}{}
	<-matsuriEndCh
}

func output() {
	s := score.Get()
	pass := true
	if fails.GetIsCritical() {
		s = 0
		pass = false
	}
	b, _ := json.Marshal(score.Output{
		Pass:     pass,
		Score:    s,
		Messages: fails.GetUnique(),
	})

	fmt.Println(string(b))
}

func makeChan(len int) chan struct{} {
	ch := make(chan struct{}, len)
	for i := 0; i < len; i++ {
		ch <- struct{}{}
	}
	return ch
}
