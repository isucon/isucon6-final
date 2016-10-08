package main

import (
	"html/template"
	"net/http"
)

type Message struct {
	Message     string
	MessageHTML template.HTML
	Kind        string
}

func serveMessages(w http.ResponseWriter, req *http.Request) error {
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

	return templates["messages.tmpl"].Execute(w, viewParamsDebugMessages{viewParamsLayout{nil, day}, msgs})
}

func getMessages() ([]Message, error) {
	msgs := make([]Message, 0)

	rows, err := db.Query(`
      SELECT message, kind FROM messages ORDER BY id ASC`)
	if err != nil {
		return msgs, err
	}

	defer rows.Close()

	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Message, &msg.Kind)
		if err != nil {
			return msgs, err
		}
		msg.MessageHTML = template.HTML(msg.Message)
		msgs = append(msgs, msg)
	}
	if err := rows.Err(); err != nil {
		return msgs, err
	}

	return msgs, nil
}

func updateMessages(msgs []Message) error {
	tx, err := db.Begin()

	_, err = db.Exec("DELETE FROM messages")
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, msg := range msgs {
		_, err = db.Exec("INSERT INTO messages (message, kind) VALUES (?,?)", msg.Message, msg.Kind)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
