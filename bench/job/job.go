package job

import "github.com/catatsuy/isucon6-final/bench/score"

type Job struct {
	ID     int    `json:"id"`
	TeamID int    `json:"teamID"`
	URLs   string `json:"urls"`
}

type Result struct {
	Job    *Job
	Output *score.Output
	Stderr string
}
