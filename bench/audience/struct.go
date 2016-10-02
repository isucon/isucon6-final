package audience

import "time"

type StrokeLog struct {
	ReceivedTime time.Time `json:"received_time"`
	RoomID       int64     `json:"room_id"`
	StrokeID     int64     `json:"stroke_id"`
}

type AudienceResponse struct {
	Errors     []string    `json:"errors"`
	StrokeLogs []StrokeLog `json:"stroke_logs"`
}
