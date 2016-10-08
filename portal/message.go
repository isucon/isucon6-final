package main

import "html/template"

type Message struct {
	Message     string
	MessageHTML template.HTML
	Kind        string
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
