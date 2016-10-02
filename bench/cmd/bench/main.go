package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/catatsuy/isucon6-final/bench/fails"
	"github.com/catatsuy/isucon6-final/bench/scenario"
	"github.com/catatsuy/isucon6-final/bench/score"
	"github.com/catatsuy/isucon6-final/bench/session"
)

var BenchmarkTimeout int
var Audience1 string

func main() {

	host := ""

	flag.StringVar(&host, "host", "", "ベンチマーク対象のIPアドレス")
	flag.StringVar(&Audience1, "audience1", "", "オーディエンスAPIのURLその1 (http://xxx.xxx.xxx.xxx/)")
	flag.IntVar(&BenchmarkTimeout, "timeout", 60, "ソフトタイムアウト")

	flag.Parse()

	if !regexp.MustCompile(`\A[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\z`).MatchString(host) {
		log.Fatal("hostの指定が間違っています（例: 127.0.0.1）")
	}
	baseURL := "https://" + host

	// 初期チェックで失敗したらそこで終了
	initialCheck(baseURL)
	if len(fails.Get()) > 0 {
		output()
		return
	}

	benchmark(baseURL)
	output()
}

func initialCheck(baseURL string) {
	scenario.CheckCSRFTokenRefreshed(session.New(baseURL))
}

func benchmark(baseURL string) {
	loadIndexPageCh := makeChan(2)
	checkCSRFTokenRefreshedCh := makeChan(1)
	matsuriRoomCh := makeChan(1)

	timeoutCh := time.After(time.Duration(BenchmarkTimeout) * time.Second)

L:
	for {
		select {
		case <-loadIndexPageCh:
			go func() {
				//scenario.LoadIndexPage(session.New(baseURL))
				loadIndexPageCh <- struct{}{}
			}()
		case <-checkCSRFTokenRefreshedCh:
			go func() {
				//scenario.CheckCSRFTokenRefreshed(session.New(baseURL))
				checkCSRFTokenRefreshedCh <- struct{}{}
			}()
		case <-matsuriRoomCh:
			go func() {
				scenario.MatsuriRoom(session.New(baseURL), Audience1)
				//matsuriRoomCh <- struct{}{} // Never again.
			}()
		case <-timeoutCh:
			break L
		}
	}
}

func output() {
	b, _ := json.Marshal(struct {
		Score    int64    `json:"score"`
		Messages []string `json:"messages"`
	}{Score: score.Get(), Messages: fails.GetUnique()})

	fmt.Println(string(b))
}

func makeChan(len int) chan struct{} {
	ch := make(chan struct{}, len)
	for i := 0; i < len; i++ {
		ch <- struct{}{}
	}
	return ch
}
