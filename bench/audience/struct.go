package audience

import "time"

type StrokeLog struct {
	Time       time.Time `json:"time"`
	RoomID     int       `json:"room_id"`
	StrokeID   int64     `json:"stroke_id"`
	StrokeTime time.Time `json:"stroke_time"`
}

type Response struct {
	Errors     []string    `json:"errors"`
	StrokeLogs []StrokeLog `json:"stroke_logs"`
}
