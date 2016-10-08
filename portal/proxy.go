package main

import (
	"net/http"
	"strconv"
)

func serveProxyUpdate(w http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodPost {
		return errHTTP(http.StatusMethodNotAllowed)
	}
	// TODO: consul からノード一覧を取得してproxyのホスト一覧を更新する（とりあえず仮で127.0.0.1を入れてある）
	_, err := db.Exec("INSERT INTO proxies (host) VALUES (?)", "127.0.0.1")
	if err != nil {
		return errHTTP(http.StatusInternalServerError)
	}

	return nil
}

func serveProxyNginxConf(w http.ResponseWriter, req *http.Request) error {
	// TODO: teamテーブルに登録されているIPアドレス一覧から組み立てる（ポート番号は10000+teamIDで良さそう）
	// nginxはstreamディレクティブの中でこのファイルをincludeする。ファイルを更新したらreload
	b := []byte(`
    server {
        listen 10001; # team1
        proxy_pass 127.0.0.1:443;
    }
    server {
        listen 10002; # team2
        proxy_pass 127.0.0.1:443;
    }
    # ...
    server {
        listen 10999; # team999
        proxy_pass 127.0.0.1:443;
    }
	`)
	w.Write(b)
	return nil
}

func getProxyHosts() ([]string, error) {
	hosts := make([]string, 0)

	rows, err := db.Query(`
      SELECT host FROM proxies`)
	if err != nil {
		return hosts, err
	}

	defer rows.Close()

	for rows.Next() {
		var Host string
		err := rows.Scan(&Host)
		if err != nil {
			return hosts, err
		}
		hosts = append(hosts, Host)
	}
	if err := rows.Err(); err != nil {
		return hosts, err
	}

	return hosts, nil
}

func getProxyURLs(teamID int) (string, error) {
	hosts, err := getProxyHosts()
	if err != nil {
		return "", err
	}
	urls := ""
	port := strconv.Itoa(teamID)

	for i, host := range hosts {
		if i != 0 {
			urls += ","
		}
		urls += "https://" + host + ":" + port
	}
	return urls, nil
}
