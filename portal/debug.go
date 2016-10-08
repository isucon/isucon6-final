package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

func expvarHandler(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")

	fmt.Fprintf(w, "%q: ", "db")
	json.NewEncoder(w).Encode(db.Stats())

	fmt.Fprintf(w, ",\n%q: ", "runtime")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"NumGoroutine": runtime.NumGoroutine(),
	})

	rows, err := db.Query("SELECT status,COUNT(*) FROM queues GROUP BY status")
	if err != nil {
		return err
	}
	defer rows.Close()
	queueStats := map[string]int{}
	for rows.Next() {
		var (
			st string
			c  int
		)
		rows.Scan(&st, &c)
		queueStats[st] = c
	}

	fmt.Fprintf(w, ",\n%q: ", "queue")
	json.NewEncoder(w).Encode(queueStats)

	fmt.Fprintf(w, ",\n%q: ", "app")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":   appVersion,
		"startedAt": appStartedAt,
	})

	expvar.Do(func(kv expvar.KeyValue) {
		fmt.Fprintf(w, ",\n")
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")

	return nil
}

func serveDebugQueue(w http.ResponseWriter, req *http.Request) error {
	rows, err := db.Query(`
      SELECT
        queues.id,team_id,name,status,queues.ip_address,IFNULL(bench_node, ''),IFNULL(result_json, ''),created_at
      FROM queues
        LEFT JOIN teams ON queues.team_id = teams.id
      ORDER BY queues.created_at DESC
      LIMIT 50
	`)
	if err != nil {
		return err
	}

	type queueItem struct {
		ID        int
		TeamID    int
		TeamName  string
		Status    string
		IPAddr    string
		BenchNode string
		Result    string
		Time      time.Time
	}

	type viewParamsDebugQueue struct {
		viewParamsLayout
		Items []*queueItem
	}

	items := []*queueItem{}

	defer rows.Close()
	for rows.Next() {
		var item queueItem
		err := rows.Scan(&item.ID, &item.TeamID, &item.TeamName, &item.Status, &item.IPAddr, &item.BenchNode, &item.Result, &item.Time)
		if err != nil {
			return err
		}

		items = append(items, &item)
	}

	return templates["debug-queue.tmpl"].Execute(w, viewParamsDebugQueue{viewParamsLayout{nil, day}, items})
}

func serveDebugProxies(w http.ResponseWriter, req *http.Request) error {
	addrs, err := getProxyAddrs()
	if err != nil {
		return err
	}

	type viewParamsDebugProxies struct {
		viewParamsLayout
		Addrs []string
	}

	return templates["debug-proxies.tmpl"].Execute(w, viewParamsDebugProxies{viewParamsLayout{nil, day}, addrs})
}

func serveDebugMessages(w http.ResponseWriter, req *http.Request) error {
	if req.Method == http.MethodPost {
		var msgs []Message

		err := req.ParseForm()
		if err != nil {
			return err
		}
		l := len(req.PostForm["kind"])
		if l != len(req.PostForm["kind"]) {
			return errHTTP(http.StatusBadRequest)
		}
		for i := 0; i < l; i++ {
			kind := req.PostForm["kind"][i]
			message := req.PostForm["message"][i]
			if message != "" {
				msgs = append(msgs, Message{Kind: kind, Message: message})
			}
		}
		err = updateMessages(msgs)
		if err != nil {
			return err
		}
	}

	msgs, err := getMessages()
	msgs = append(msgs, Message{})
	if err != nil {
		return err
	}

	type viewParamsDebugMessages struct {
		viewParamsLayout
		Messages []Message
	}

	return templates["debug-messages.tmpl"].Execute(w, viewParamsDebugMessages{viewParamsLayout{nil, day}, msgs})
}
