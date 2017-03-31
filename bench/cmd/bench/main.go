package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sesta/isucon6-final/bench/fails"
	"github.com/sesta/isucon6-final/bench/scenario"
	"github.com/sesta/isucon6-final/bench/score"
	"github.com/sesta/isucon6-final/portal/job"
)

var BenchmarkTimeout int
var InitialCheckOnly bool
var MatsuriNum = 10
var LoadIndexPageNum = 10
var DrawOnRandomRoomNum = 2

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	var urls string

	flag.StringVar(&urls, "urls", "", "ベンチマーク対象のURL（scheme, host, portまで。カンマ区切りで複数可。例： https://xxx.xxx.xxx.xxx,https://xxx.xxx.xxx.xxx:1443）")
	flag.IntVar(&BenchmarkTimeout, "timeout", 60, "ソフトタイムアウト")
	flag.BoolVar(&InitialCheckOnly, "initialcheck", false, "初期チェックだけ行う")

	flag.Parse()

	origins, err := makeOrigins(urls)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}

	initialCheck(origins)

	// 初期チェックのみモードではない、かつ、この時点でcriticalが出ていなければ負荷をかけにいく
	if !InitialCheckOnly && !fails.GetIsCritical() {
		benchmark(origins)
	}

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
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		scenario.CSRFTokenRefreshed(origins)
		scenario.StrokeReflectedToTop(origins)
		scenario.RoomWithoutStrokeNotShownAtTop(origins)
		scenario.CantDrawFirstStrokeOnSomeoneElsesRoom(origins)
		scenario.TopPageContent(origins)
		scenario.APIAndHTMLMustBeConsistent(origins)
		scenario.CheckStaticFiles(origins)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		scenario.WatcherCountIncreases(origins)
	}()
	wg.Wait()
}

func benchmark(origins []string) {
	var wg sync.WaitGroup
	for i := 0; i < MatsuriNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scenario.Matsuri(origins, BenchmarkTimeout-5)
		}()
	}

	loadIndexPageCh := makeChan(LoadIndexPageNum)
	drawOnRandomRoomCh := makeChan(DrawOnRandomRoomNum)
	timeoutCh := time.After(time.Duration(BenchmarkTimeout) * time.Second)

L:
	for {
		select {
		case <-loadIndexPageCh:
			go func() {
				ok := scenario.LoadIndexPage(origins)
				if ok {
					time.Sleep(500 * time.Millisecond)
				} else {
					fmt.Fprintln(os.Stderr, "LoadIndexPage failed. waiting for 1s.")
					time.Sleep(1000 * time.Millisecond)
				}
				loadIndexPageCh <- struct{}{}
			}()
		case <-drawOnRandomRoomCh:
			go func() {
				scenario.DrawOnRandomRoom(origins)
				time.Sleep(500 * time.Millisecond)
				drawOnRandomRoomCh <- struct{}{}
			}()
		case <-timeoutCh:
			break L
		}
	}

	wg.Wait()
}

func output() {
	b, _ := json.Marshal(job.Output{
		Pass:     !fails.GetIsCritical(),
		Score:    score.Get(),
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
