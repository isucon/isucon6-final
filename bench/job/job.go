package job

type Job struct {
	ID     int    `json:"id"`
	TeamID int    `json:"teamID"`
	URLs   string `json:"urls"`
}

type Result struct {
	Job    *Job
	Output *Output
	Stderr string
}

type Output struct {
	Pass     bool     `json:"pass"`
	Score    int64    `json:"score"`
	Messages []string `json:"messages"`
}
