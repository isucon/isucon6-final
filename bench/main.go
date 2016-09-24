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

var BenchmarkTimeout = 30 * time.Second

func main() {

	host := ""

	flag.StringVar(&host, "host", "", "ベンチマーク対象のIPアドレス")

	flag.Parse()

	if !regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`).MatchString(host) {
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

	timeoutCh := time.After(BenchmarkTimeout)

L:
	for {
		select {
		case <-loadIndexPageCh:
			go func() {
				scenario.LoadIndexPage(session.New(baseURL))
				loadIndexPageCh <- struct{}{}
			}()
		case <-checkCSRFTokenRefreshedCh:
			go func() {
				scenario.CheckCSRFTokenRefreshed(session.New(baseURL))
				checkCSRFTokenRefreshedCh <- struct{}{}
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
