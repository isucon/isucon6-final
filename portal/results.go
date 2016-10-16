package main

import (
	"database/sql"
	"sort"
	"time"
)

type PlotLine struct {
	TeamName string           `json:"name"`
	Data     map[string]int64 `json:"data"`
}

type LatestScore struct {
	TeamID   int
	TeamName string
	Score    int64
	At       time.Time
}

type LatestScores []LatestScore

func (ls LatestScores) Len() int           { return len(ls) }
func (ls LatestScores) Less(i, j int) bool { return ls[i].Score > ls[j].Score }
func (ls LatestScores) Swap(i, j int)      { ls[i], ls[j] = ls[j], ls[i] }

// 自分のチームであれば問答無用で、そうでなければオフィシャルユーザーでなくランキング固定の時間より前のデータを取得
// プロットには成功したスコアしか載せない
func getResults(db *sql.DB, teamID int, topNum int, rankingFixAt time.Time) ([]PlotLine, []LatestScore, error) {
	rows, err := db.Query(`
SELECT teams.id, teams.name, results.score, results.created_at
FROM results JOIN teams ON results.team_id = teams.id
WHERE results.pass = 1
AND (teams.id = ? OR (teams.category <> 'official' AND results.created_at <= ?))
ORDER BY results.team_id ASC, results.id ASC
	`, teamID, rankingFixAt)

	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	plotLines := []PlotLine{}
	latestScores := []LatestScore{}
	lastTeamID := 0
	var plotLine PlotLine
	var latestScore LatestScore

	for rows.Next() {
		var (
			TeamID   int
			TeamName string
			Score    int64
			At       time.Time
		)

		err := rows.Scan(&TeamID, &TeamName, &Score, &At)
		if err != nil {
			return nil, nil, err
		}

		if lastTeamID != TeamID {
			if lastTeamID != 0 {
				plotLines = append(plotLines, plotLine)
				latestScores = append(latestScores, latestScore)
			}
			lastTeamID = TeamID
			plotLine = PlotLine{TeamName: TeamName, Data: make(map[string]int64)}
		}
		plotLine.Data[At.Format("2006-01-02T15:04:05")] = Score

		latestScore = LatestScore{TeamID: TeamID, TeamName: TeamName, Score: Score, At: At}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if lastTeamID != 0 {
		plotLines = append(plotLines, plotLine)
		latestScores = append(latestScores, latestScore)
	}

	sort.Sort(LatestScores(latestScores))
	topLatestScores := []LatestScore{}
	for i, ls := range latestScores {
		if i < topNum {
			topLatestScores = append(topLatestScores, ls)
		}
	}

	return plotLines, latestScores, nil
}

type TeamResult struct {
	ID    int
	Score int64
	Pass  int
	At    time.Time
	Msg   string
}

// 特定のチームのスコアとメッセージ一覧を全部取得
func getTeamResults(db *sql.DB, teamID int) ([]TeamResult, error) {
	rows, err := db.Query(`
SELECT id, score, pass, created_at, messages FROM results
WHERE team_id = ?
ORDER BY id DESC
	`, teamID)

	if err != nil {
		return nil, err
	}

	teamResults := []TeamResult{}

	defer rows.Close()

	for rows.Next() {
		var r TeamResult

		err := rows.Scan(&r.ID, &r.Score, &r.Pass, &r.At, &r.Msg)
		if err != nil {
			return nil, err
		}

		teamResults = append(teamResults, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teamResults, nil
}
