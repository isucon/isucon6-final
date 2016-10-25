package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type AgentMember struct {
	Name   string
	Addr   string
	Status int
}

// consulの/v1/agent/membersをそのままPOSTする用
// curl -s '127.0.0.1:8500/v1/agent/members' | curl -XPOST -d @- http://127.0.0.1/mBGWHqBVEjUSKpBF/proxy/update
// https://github.com/isucon/isucon6-final/pull/121#issuecomment-252422888
func serveProxyUpdate(w http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodPost {
		return errHTTP(http.StatusMethodNotAllowed)
	}
	var members []AgentMember
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &members)
	if err != nil {
		return err
	}

	proxyAddrs := make([]string, 0)
	for _, m := range members {
		if m.Status == 1 && strings.Contains(m.Name, "proxy") { // FIXME: 決め打ちで良いか？
			proxyAddrs = append(proxyAddrs, "('"+m.Addr+"')")
		}
	}

	if len(proxyAddrs) == 0 {
		w.Write([]byte(``))
		return nil
	}

	tx, err := db.Begin()

	_, err = db.Exec("DELETE FROM proxies")
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = db.Exec("INSERT INTO proxies (ip_address) VALUES " + strings.Join(proxyAddrs, ","))
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	w.Write([]byte(strings.Join(proxyAddrs, "\n")))
	return nil
}

func serveProxyNginxConf(w http.ResponseWriter, req *http.Request) error {
	conf := ""
	rows, err := db.Query("SELECT id, IFNULL(ip_address,'') FROM teams")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ID int
		var IPAddr string
		err := rows.Scan(&ID, &IPAddr)
		if err != nil {
			return err
		}
		if IPAddr != "" {
			conf += fmt.Sprintf(`
# team%d
server {
  listen %d;
  proxy_pass %s:443;
}`,
				ID, teamIDToPortNum(ID), IPAddr)
		}
	}
	w.Write([]byte(conf))
	return nil
}

func teamIDToPortNum(teamID int) int {
	return teamID + 10000
}

func getProxyAddrs() ([]string, error) {
	addrs := make([]string, 0)

	rows, err := db.Query(`
      SELECT ip_address FROM proxies ORDER BY ip_address ASC`)
	if err != nil {
		return addrs, err
	}

	defer rows.Close()

	for rows.Next() {
		var Addr string
		err := rows.Scan(&Addr)
		if err != nil {
			return addrs, err
		}
		addrs = append(addrs, Addr)
	}
	if err := rows.Err(); err != nil {
		return addrs, err
	}

	return addrs, nil
}

func getProxyURLs(teamID int) (string, error) {
	addrs, err := getProxyAddrs()
	if err != nil {
		return "", err
	}
	urls := ""

	for i, addr := range addrs {
		if i != 0 {
			urls += ","
		}
		urls += "https://" + addr + ":" + strconv.Itoa(teamIDToPortNum(teamID))
	}
	return urls, nil
}
