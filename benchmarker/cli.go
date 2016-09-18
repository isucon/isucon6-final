package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/catatsuy/isucon6-final/benchmarker/checker"
	"github.com/catatsuy/isucon6-final/benchmarker/score"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota

	FailThreshold     = 5
	InitializeTimeout = time.Duration(10) * time.Second
	BenchmarkTimeout  = 10 * time.Second
	WaitAfterTimeout  = 10

	PostsPerPage = 20
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
}

type user struct {
	AccountName string
	Password    string
}

type Output struct {
	Pass     bool     `json:"pass"`
	Score    int64    `json:"score"`
	Suceess  int64    `json:"success"`
	Fail     int64    `json:"fail"`
	Messages []string `json:"messages"`
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		target   string
		seedData string

		version bool
		debug   bool
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	flags.StringVar(&target, "target", "", "")
	flags.StringVar(&target, "t", "", "(Short)")

	flags.StringVar(&seedData, "seeddata", "", "seedData directory")
	flags.StringVar(&seedData, "s", "", "seedData directory")

	flags.BoolVar(&version, "version", false, "Print version information and quit.")

	flags.BoolVar(&debug, "debug", false, "Debug mode")
	flags.BoolVar(&debug, "d", false, "Debug mode")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeError
	}

	// Show version
	if version {
		fmt.Fprintf(cli.errStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	targetHost, terr := checker.SetTargetHost(target)
	if terr != nil {
		outputNeedToContactUs(terr.Error())
		return ExitCodeError
	}

	initialize := make(chan bool)

	setupInitialize(targetHost, initialize)

	strokes, err := prepareSeedData(seedData)
	if err != nil {
		fmt.Println("%v", err)
	}

	initReq := true

	if !initReq {
		fmt.Println(outputResultJson(false, []string{"初期化リクエストに失敗しました"}))

		return ExitCodeError
	}

	// 最初にDOMチェックなどをやってしまい、通らなければさっさと失敗させる

	if score.GetInstance().GetFails() > 0 {
		fmt.Println(outputResultJson(false, score.GetFailErrorsStringSlice()))
		return ExitCodeError
	}

	makeNewRoomScenarioCh := makeChan(2)

	timeoutCh := time.After(BenchmarkTimeout)

L:
	for {
		select {
		case <-makeNewRoomScenarioCh:
			go func() {
				makeNewRoomScenario(checker.NewSession())
				makeNewRoomScenarioCh <- struct{}{}
			}()
		case <-timeoutCh:
			break L
		}
	}

	time.Sleep(WaitAfterTimeout)

	msgs := []string{}

	if !debug {
		msgs = score.GetFailErrorsStringSlice()
	} else {
		msgs = score.GetFailRawErrorsStringSlice()
	}

	fmt.Println(outputResultJson(true, msgs))

	return ExitCodeOK
}

func outputResultJson(pass bool, messages []string) string {
	output := Output{
		Pass:     pass,
		Score:    score.GetInstance().GetScore(),
		Suceess:  score.GetInstance().GetSucesses(),
		Fail:     score.GetInstance().GetFails(),
		Messages: messages,
	}

	b, _ := json.Marshal(output)

	return string(b)
}

// 主催者に連絡して欲しいエラー
func outputNeedToContactUs(message string) {
	outputResultJson(false, []string{"！！！主催者に連絡してください！！！", message})
}

func makeChan(len int) chan struct{} {
	ch := make(chan struct{}, len)
	for i := 0; i < len; i++ {
		ch <- struct{}{}
	}
	return ch
}

func setupInitialize(targetHost string, initialize chan bool) {
	go func(targetHost string) {
		client := &http.Client{
			Timeout: InitializeTimeout,
		}

		parsedURL, _ := url.Parse("/initialize")
		parsedURL.Scheme = "http"
		parsedURL.Host = targetHost

		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			return
		}

		req.Header.Set("User-Agent", checker.UserAgent)

		res, err := client.Do(req)

		if err != nil {
			initialize <- false
			return
		}
		defer res.Body.Close()
		initialize <- true
	}(targetHost)
}
