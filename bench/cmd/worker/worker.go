package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/Songmu/timeout"
	"github.com/sesta/isucon6-final/portal/job"
	"github.com/mitchellh/go-homedir"
)

// CLI is the command line object
type CLI struct {
	outStream, errStream io.Writer
}

func main() {
	cli := &CLI{outStream: os.Stdout, errStream: os.Stderr}
	os.Exit(cli.Run(os.Args[1:]))
}

const (
	exitCodeOK = iota
	exitCodeErr

	maxStderrLen = 10000
)

func (cli *CLI) Run(args []string) int {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(cli.errStream)
	var (
		portalHost string
		benchPath  string
	)
	h, _ := homedir.Dir()
	defaultBenchPath := filepath.Join(h, "isucon6f", "bench")
	fs := flag.NewFlagSet("isu6q bench worker", flag.ContinueOnError)
	fs.SetOutput(cli.errStream)
	fs.StringVar(&portalHost, "portal", "localhost", "portal Host")
	fs.StringVar(&benchPath, "bench", defaultBenchPath, "benchmark path")

	if err := fs.Parse(args); err != nil {
		return exitCodeErr
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		for range c {
			sigReceived = true
		}
	}()
	return cli.start(portalHost, benchPath)
}

type portal struct {
	host string
}

func (ptl *portal) newJobURL() string {
	return fmt.Sprintf("http://%s/mBGWHqBVEjUSKpBF/job/new", ptl.host)
}

func (ptl *portal) resultURL() string {
	return fmt.Sprintf("http://%s/mBGWHqBVEjUSKpBF/job/result", ptl.host)
}

var sigReceived bool

func (ptl *portal) waitJob() *job.Job {
	for !sigReceived {
		time.Sleep(3 * time.Second)
		j, err := ptl.fetchJob()
		if err != nil {
			log.Println(err)
			continue
		}
		if j != nil {
			return j
		}
	}
	return nil
}

func (ptl *portal) fetchJob() (*job.Job, error) {
	u := ptl.newJobURL()
	vals := url.Values{}
	h, _ := os.Hostname()
	vals.Set("bench_node", h)
	req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var j job.Job
		err = json.NewDecoder(resp.Body).Decode(&j)
		return &j, err
	case http.StatusNoContent:
		return nil, nil
	default:
		dump, _ := httputil.DumpResponse(resp, true)
		return nil, fmt.Errorf("response invalid: %s", string(dump))
	}
	return nil, nil
}

func (ptl *portal) postResult(res *job.Result) error {
	u := ptl.resultURL()
	json, _ := json.Marshal(res)

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (cli *CLI) start(portalHost, benchPath string) int {
	ptl := &portal{host: portalHost}
	for !sigReceived {
		j := ptl.waitJob()
		if sigReceived {
			break
		}
		if j == nil {
			continue
		}
		log.Printf("received job: %#v\n", j)
		tio := buildBenchCmd(benchPath, j.URLs)
		_, out, stderr, err := tio.Run()
		if err != nil {
			log.Println(err)
		}
		if stderr != "" {
			log.Println(stderr)
		}
		log.Println("bench result: " + out)
		var o job.Output
		err = json.Unmarshal([]byte(out), &o)
		if err != nil {
			log.Printf("bench failed: %#v, err: %s\n", j, err.Error())
		}
		if utf8.RuneCountInString(stderr) > maxStderrLen {
			// stderrが大きすぎたら削る
			stderr = string([]rune(stderr)[:maxStderrLen])
		}
		res := &job.Result{
			Job:    j,
			Output: &o,
			Stderr: stderr,
		}
		err = ptl.postResult(res)
		if err != nil {
			log.Println(err)
		}

	}
	return exitCodeOK
}

func buildBenchCmd(benchPath, urls string) *timeout.Timeout {
	log.Println(benchPath, "-urls", urls, "-timeout", "60")
	cmd := exec.Command(benchPath, "-urls", urls, "-timeout", "60")
	cmd.Dir = filepath.Dir(benchPath)
	return &timeout.Timeout{
		Cmd:       cmd,
		Duration:  120 * time.Second,
		KillAfter: 5 * time.Second,
	}
}
