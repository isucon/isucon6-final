package sse

import (
	"net/http"
	"time"
	"strconv"
	"bufio"
	"strings"
)

type Listener func(data string) error

type Stream struct {
	Client *http.Client
	Request *http.Request
	Listeners map[string]Listener
	RetryTime time.Duration
	WillRetry bool
	LastEventID string
	URL string
}

func NewStream(c *http.Client, urlStr string) (*Stream) {
	return &Stream{
		Client: c,
		Listeners: map[string]Listener{},
		RetryTime: 1 * time.Second,
		WillRetry: true,
		URL: urlStr,
	}
}

func (s *Stream) On(event string, listener Listener) {
	s.Listeners[event] = listener
}

func (s *Stream) emit(event string, data string) {
	if listener, ok := s.Listeners[event]; ok {
		listener(data)
	}
}

func (s *Stream) Close() {
	if s.Request != nil {
		s.WillRetry = false
		//close(s.Request.Cancel) // TODO: use context.Context for cancel
	}
}

func (s *Stream) retry() {
	if s.WillRetry {
		time.Sleep(s.RetryTime)
		s.Start()
	}
}

var defaultEvent = "message"

func (s *Stream) Start() {
	s.Request = nil // TODO: mutex
	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		return // TODO: errならabortでいいか
	}
	s.Request = req

	// TODO: should support means to set other headers?
	req.Header.Set("Accept", "text/event-stream")

	// TODO: User-Agent

	if s.LastEventID != "" {
		req.Header.Set("Last-Event-ID", s.LastEventID)
	}

	// TODO: mutex?
	t := s.Client.Timeout
	s.Client.Timeout = 0
	resp, err := s.Client.Do(req)
	s.Client.Timeout = t
	if err != nil {
		s.emit("error", "connection error: " + err.Error())
	}
	defer resp.Body.Close()

	if 500 <= resp.StatusCode {
		s.emit("error", "bad response status: " + strconv.Itoa(resp.StatusCode))
		s.retry()
		return
	} else if 400 <= resp.StatusCode && resp.StatusCode < 500 {
		// 4XX系ならretryしない
		s.emit("error", "bad response status: " + strconv.Itoa(resp.StatusCode))
		return
	}
	// TODO: 3XXのときはどうするか

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/event-stream") {
		s.emit("error", "bad content type")
		s.retry()
		return
	}

	data := ""
	event := defaultEvent

	scanner := bufio.NewScanner(resp.Body) // TODO: もしBOMがあったら無視する仕様

	for scanner.Scan() { // TODO: scanner.Scanは\r?\nをdelimiterとするが、SSEの仕様上は\r単独もあり得る

		line := scanner.Text()

		// https://www.w3.org/TR/eventsource/#event-stream-interpretation
		if line == "" {
			if data != "" {
				s.emit(event, data)
				event = defaultEvent
				data = ""
			}
			continue
		}
		split := strings.SplitN(line, ":", 2)
		field := split[0]
		value := ""
		if len(split) == 2 {
			value = strings.TrimPrefix(split[1], " ")
		}
		switch field {
		case "event":
			event = value
		case "retry":
			if n, err := strconv.Atoi(value); err != nil {
				s.RetryTime = time.Duration(n) * time.Millisecond
			}
		case "id":
			s.LastEventID = value
		case "data":
			if data != "" {
				data += "\n"
			}
			data += value
		default:
			// ignore
		}
	}

	if err := scanner.Err(); err != nil {
		s.emit("error", err)
	}
	s.retry()
}

