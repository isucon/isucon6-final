package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/catatsuy/isucon6-final/portal/job"
)

var s *httptest.Server

func TestMain(m *testing.M) {
	*startsAtHour = -1
	*endsAtHour = -1

	flag.Parse()
	err := initWeb()
	if err != nil {
		log.Fatal(err)
	}

	s = httptest.NewServer(buildMux())
	n := m.Run()
	s.Close()
	os.Exit(n)
}

type testHTTPClient struct {
	*http.Client
	*testing.T
}

func (c *testHTTPClient) Must(resp *http.Response, err error) *http.Response {
	require.NoError(c.T, err)
	return resp
}

func newTestClient(t *testing.T) *testHTTPClient {
	jar, _ := cookiejar.New(nil)
	return &testHTTPClient{
		Client: &http.Client{Jar: jar},
		T:      t,
	}
}

func TestLogin(t *testing.T) {
	resp, err := http.Get(s.URL)
	require.NoError(t, err)
	require.Equal(t, "/login", resp.Request.URL.Path)

	jar, _ := cookiejar.New(nil)
	cli := &http.Client{Jar: jar}

	resp, err = cli.PostForm(s.URL+"/login", url.Values{"team_id": {"26"}, "password": {"p6aYuUempoticryg"}})
	require.NoError(t, err)
	require.Equal(t, "/", resp.Request.URL.Path)
}

func readAll(r io.Reader) string {
	b, _ := ioutil.ReadAll(r)
	return string(b)
}

func benchGetJob(bench *testHTTPClient) *job.Job {
	resp := bench.Must(bench.Post(s.URL+"/mBGWHqBVEjUSKpBF/job/new", "", nil))
	if !assert.Equal(bench.T, http.StatusOK, resp.StatusCode) {
		return nil
	}

	var j job.Job
	err := json.NewDecoder(resp.Body).Decode(&j)
	require.NoError(bench.T, err)

	return &j
}

func benchPostResult(bench *testHTTPClient, j *job.Job, output *job.Output) {
	time.Sleep(1 * time.Second)

	result := job.Result{
		Job:    j,
		Output: output,
		Stderr: "",
	}
	resultJSON, err := json.Marshal(result)
	require.NoError(bench.T, err)

	resp := bench.Must(bench.Post(s.URL+"/mBGWHqBVEjUSKpBF/job/result", "application/json", bytes.NewBuffer(resultJSON)))
	require.Equal(bench.T, http.StatusOK, resp.StatusCode)
}

func cliLogin(cli *testHTTPClient, teamID int, password string) {
	resp := cli.Must(
		cli.PostForm(
			s.URL+"/login",
			url.Values{
				"team_id":  {fmt.Sprint(teamID)},
				"password": {password},
			},
		),
	)
	require.Equal(cli.T, "/", resp.Request.URL.Path)
}

func TestPostJob(t *testing.T) {
	var (
		cli   = newTestClient(t)
		cli2  = newTestClient(t)
		bench = newTestClient(t)
	)

	var resp *http.Response

	// cli: ログイン
	resp = cli.Must(cli.PostForm(s.URL+"/login", url.Values{"team_id": {"26"}, "password": {"p6aYuUempoticryg"}}))
	require.Equal(t, "/", resp.Request.URL.Path)

	// bench: ジョブ取る
	resp = bench.Must(bench.Post(s.URL+"/mBGWHqBVEjUSKpBF/job/new", "", nil))
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// cli: ジョブいれる→まだIP入れてないのでエラー
	resp = cli.Must(cli.PostForm(s.URL+"/queue", nil))
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// cli: IP入れる
	resp = cli.Must(cli.PostForm(s.URL+"/team", url.Values{"ip_address": {"127.0.0.1"}, "instance_name": {""}}))
	assert.Contains(t, readAll(resp.Body), `<input class="form-control" type="text" name="ip_address" value="127.0.0.1" autocomplete="off">`, "IP入った表示")

	// cli: ジョブ入れる
	resp = cli.Must(cli.PostForm(s.URL+"/queue", nil))
	assert.Contains(t, readAll(resp.Body), `<span class="label label-default">26*</span>`, "ジョブ入った表示")

	// cli2: ログイン
	resp = cli2.Must(cli2.PostForm(s.URL+"/login", url.Values{"team_id": {"5"}, "password": {"Y7i06XOllyJI5ogn"}}))
	require.Equal(t, "/", resp.Request.URL.Path)
	assert.Contains(t, readAll(resp.Body), `<span class="label label-default">26</span>`, "他人のジョブ入った表示")

	// bench: ジョブ取る
	j := benchGetJob(bench)

	// cli: ジョブいれる (2) → 入らない
	resp = cli.Must(cli.PostForm(s.URL+"/queue", nil))
	assert.Contains(t, readAll(resp.Body), `Job already queued`)

	// cli2: IP入れる
	resp = cli2.Must(cli2.PostForm(s.URL+"/team", url.Values{"ip_address": {"127.0.0.2"}, "instance_name": {""}}))
	assert.Contains(t, readAll(resp.Body), `<input class="form-control" type="text" name="ip_address" value="127.0.0.2" autocomplete="off">`, "IP入った表示")

	// cli2: ジョブ入れる → 入る
	resp = cli2.Must(cli2.PostForm(s.URL+"/queue", nil))
	assert.Contains(t, readAll(resp.Body), `<span class="label label-default">5*</span>`, "ジョブ入った表示")

	// bench: ジョブ取る → 放置
	j2 := benchGetJob(bench)
	_ = j2

	// cli: トップリロード
	resp = cli.Must(cli.Get(s.URL + "/"))
	assert.Contains(t, readAll(resp.Body), `<span class="label label-success">26*</span>`, "ジョブ実行中の表示")

	// bench: 結果入れる
	benchPostResult(bench, j, &job.Output{Pass: false, Score: 5000})

	// cli: トップリロード
	resp = cli.Must(cli.Get(s.URL + "/"))
	body := readAll(resp.Body)
	require.Contains(t, body, `<th>Status</th><td>FAIL</td>`)
	require.Contains(t, body, `<th>Score</th><td>5000</td>`)
	require.Contains(t, body, `<th>Best</th><td>-</td>`)

	// cli: ジョブいれる (3)
	resp = cli.Must(cli.PostForm(s.URL+"/queue", url.Values{"ip_addr": {"127.0.0.1"}}))
	assert.NotContains(t, readAll(resp.Body), `Job already queued`)

	// bench: ジョブ取る
	j = benchGetJob(bench)

	// bench: 結果入れる
	benchPostResult(bench, j, &job.Output{Pass: true, Score: 3000})

	// cli: トップリロード
	resp = cli.Must(cli.Get(s.URL + "/"))
	body = readAll(resp.Body)
	require.Contains(t, body, `<th>Status</th><td>PASS</td>`)
	require.Contains(t, body, `<th>Score</th><td>3000</td>`)
	require.Contains(t, body, `<th>Best</th><td>3000</td>`)
	require.Regexp(t, `<td>RUDT</td>\s*<td>3000</td>`, body)
	require.NotContains(t, body, "流れ弾")

	// bench: 結果入れる
	benchPostResult(bench, j2, &job.Output{Pass: true, Score: 4500})

	resp = cli2.Must(cli2.Get(s.URL + "/"))
	body = readAll(resp.Body)
	require.Contains(t, body, `<th>Status</th><td>PASS</td>`)
	require.Contains(t, body, `<th>Score</th><td>4500</td>`)
	require.Contains(t, body, `<th>Best</th><td>4500</td>`)
	require.Regexp(t, `<td>流れ弾</td>\s*<td>4500</td>(?s:.*)<td>RUDT</td>\s*<td>3000</td>`, body)
}

func TestPostJobNotWithinContestTime(t *testing.T) {
	cli := newTestClient(t)
	cliLogin(cli, 10, "3svumTo3amDShlFy")

	var resp *http.Response

	*startsAtHour = 24
	resp = cli.Must(cli.PostForm(s.URL+"/queue", nil))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "Final has not started yet\n", readAll(resp.Body))
	*startsAtHour = -1

	*endsAtHour = 0
	resp = cli.Must(cli.PostForm(s.URL+"/queue", nil))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "Final has finished\n", readAll(resp.Body))
	*endsAtHour = -1
}

func TestUpdateTeam(t *testing.T) {
	cli := newTestClient(t)
	admin := newTestClient(t)
	cliLogin(cli, 11, "L6KZ7UJyAEtpVr9G")

	resp := cli.Must(cli.PostForm(s.URL+"/team", url.Values{"instance_name": {"xxxxxx"}, "ip_address": {"0.0.0.0"}}))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := readAll(resp.Body)
	assert.Contains(t, body, `value="xxxxxx"`)
	assert.Contains(t, body, `value="0.0.0.0"`)

	resp = cli.Must(cli.Get(s.URL + "/"))
	body = readAll(resp.Body)
	assert.Contains(t, body, `value="xxxxxx"`)
	assert.Contains(t, body, `value="0.0.0.0"`)

	resp = admin.Must(admin.Get(s.URL + "/mBGWHqBVEjUSKpBF/proxy/nginx.conf"))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readAll(resp.Body)
	assert.Contains(t, body, `# team11`)
	assert.Contains(t, body, `listen 10011;`)
	assert.Contains(t, body, `proxy_pass 0.0.0.0;`)
}

func TestUpdateProxies(t *testing.T) {
	cli := newTestClient(t)
	bench := newTestClient(t)
	admin := newTestClient(t)
	cliLogin(cli, 12, "YJUaDANoex8Y6MB")

	// proxyのIP一覧を入れる
	nodes := `[{"Name":"portal","Addr":"192.168.0.10","Status":1},{"Name":"isu-proxy-1","Addr":"192.168.0.11","Status":1},{"Name":"isu-proxy-2","Addr":"192.168.0.12","Status":1},{"Name":"isu-proxy-3","Addr":"192.168.0.13","Status":0}]`
	resp := admin.Must(admin.Post(s.URL+"/mBGWHqBVEjUSKpBF/proxy/update", "application/json", bytes.NewBuffer([]byte(nodes))))
	body := readAll(resp.Body)
	assert.NotContains(t, body, `192.168.0.10`, "portalのIP")
	assert.Contains(t, body, `192.168.0.11`, "proxy-1のIP")
	assert.Contains(t, body, `192.168.0.12`, "proxy-2のIP")
	assert.NotContains(t, body, `192.168.0.13`, "proxy-3のIP")

	// cli: IP入れる
	resp = cli.Must(cli.PostForm(s.URL+"/team", url.Values{"ip_address": {"127.0.0.1"}, "instance_name": {""}}))

	// cli: ジョブ入れる
	resp = cli.Must(cli.PostForm(s.URL+"/queue", nil))

	// bench: ジョブ取る
	j := benchGetJob(bench)
	require.Equal(t, 12, j.TeamID)
	assert.Contains(t, j.URLs, `https://192.168.0.11:10012`, "proxy-1のIP")
	assert.Contains(t, j.URLs, `https://192.168.0.12:10012`, "proxy-2のIP")
	assert.NotContains(t, j.URLs, `https://192.168.0.13:10012`, "proxy-3のIP")
}
