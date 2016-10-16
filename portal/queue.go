package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/catatsuy/isucon6-final/portal/job"
	"github.com/pkg/errors"
)

type errAlreadyQueued int

func (n errAlreadyQueued) Error() string {
	return fmt.Sprintf("job already queued (teamID=%d)", n)
}

func enqueueJob(teamID int) error {
	var id int
	err := db.QueryRow(`
      SELECT id FROM queues
      WHERE team_id = ? AND status IN ('waiting', 'running')`, teamID).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		// 行がない場合はINSERTする
	case err != nil:
		return errors.Wrap(err, "failed to enqueue job when selecting table")
	default:
		return errAlreadyQueued(teamID)
	}
	// XXX: worker nodeが死んだ時のために古くて実行中のジョブがある場合をケアした方が良いかも

	// XXX: ここですり抜けて二重で入る可能性がある
	_, err = db.Exec(`
      INSERT INTO queues (team_id) VALUES (?)`, teamID)
	if err != nil {
		return errors.Wrap(err, "enqueue job failed")
	}
	return nil
}

func dequeueJob(benchNode string) (*job.Job, error) {
	var j job.Job
	err := db.QueryRow(`
    SELECT id, team_id FROM queues
      WHERE status = 'waiting' ORDER BY id LIMIT 1`).Scan(&j.ID, &j.TeamID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "dequeue job failed when scanning job")
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to dequeue job when beginning tx")
	}
	ret, err := tx.Exec(`
    UPDATE queues SET status = 'running', bench_node = ?
      WHERE id = ? AND status = 'waiting'`, benchNode, j.ID)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "failed to dequeue job when locking")
	}
	affected, err := ret.RowsAffected()
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "failed to dequeue job when checking affected rows")
	}
	if affected > 1 {
		tx.Rollback()
		return nil, fmt.Errorf("failed to dequeue job. invalid affected rows: %d", affected)
	}
	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to dequeue job when commiting tx")
	}
	// タッチの差で別のワーカーにジョブを取られたとか
	if affected < 1 {
		return nil, nil
	}
	return &j, nil
}

func doneJob(res *job.Result) error {
	log.Printf("doneJob: job=%#v output=%#v", res.Job, res.Output)

	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "doneJob failed when beginning tx")
	}
	ret, err := tx.Exec(`
UPDATE queues
SET status = 'done', stderr = ?
WHERE id = ?
AND team_id = ?
AND status = 'running'
      `,
		res.Stderr,
		res.Job.ID,
		res.Job.TeamID,
	)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "doneJob failed when locking")
	}
	affected, err := ret.RowsAffected()
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "doneJob failed when checking affected rows")
	}
	if affected != 1 {
		tx.Rollback()
		return fmt.Errorf("doneJob failed. invalid affected rows=%d", affected)
	}

	pass := 0
	if res.Output.Pass {
		pass = 1
	}
	_, err = tx.Exec(`
INSERT INTO results (team_id, queue_id, pass, score, messages)
VALUES (?, ?, ?, ?, ?)
	`,
		res.Job.TeamID, res.Job.ID, pass, res.Output.Score, strings.Join(res.Output.Messages, "\n"),
	)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "INSERT INTO results")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "doneJob failed when commiting tx")
	}
	return nil
}

type QueuedJob struct {
	TeamID int
	Status string
}

// まだ終わってないキューを取得
func getQueuedJobs(db *sql.DB) ([]QueuedJob, error) {
	jobs := []QueuedJob{}
	if getContestStatus() == contestStatusStarted {
		rows, err := db.Query(`
			SELECT team_id, status
			FROM queues
			WHERE status IN ('waiting', 'running')
			  AND team_id <> 9999
			ORDER BY created_at ASC
		`)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var job QueuedJob
			err := rows.Scan(&job.TeamID, &job.Status)
			if err != nil {
				rows.Close()
				return nil, err
			}
			jobs = append(jobs, job)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return jobs, nil
}
