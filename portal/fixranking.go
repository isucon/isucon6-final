package main

import (
	"database/sql"
	"net/http"
	"time"
)

func getRankingFixedAt(db *sql.DB) (time.Time, error) {
	var t time.Time

	row := db.QueryRow("SELECT fixed_at FROM fixranking WHERE id = 1")
	err := row.Scan(&t)
	if err != nil {
		if err == sql.ErrNoRows {
			return t, nil
		}
		return t, err
	}

	return t, nil
}

func fixRanking(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO fixranking (id, fixed_at) VALUES(1, NOW())")
	if err != nil {
		return err
	}

	return nil
}

func serveFixRanking(w http.ResponseWriter, req *http.Request) error {

	if req.Method == http.MethodPost {
		err := fixRanking(db)
		if err != nil {
			return err
		}
	}

	fixedAt, err := getRankingFixedAt(db)
	if err != nil {
		return err
	}

	// メッセージも挿入
	msgs, err := getMessages()
	if err != nil {
		return err
	}
	msgs = append(msgs, Message{Kind: "danger", Message: "競技残り1時間を切りました。自チームを除き、スコアの表示は固定されています"})

	err = updateMessages(msgs)
	if err != nil {
		return err
	}

	return templates["fixranking.tmpl"].Execute(w, struct {
		viewParamsLayout
		IsFixed bool
		FixedAt time.Time
	}{
		viewParamsLayout{nil},
		!fixedAt.IsZero(),
		fixedAt,
	})
}
