package score

import "sync"

var mu sync.RWMutex
var score int64

func Get() int64 {
	mu.RLock()
	s := score
	mu.RUnlock()
	return s
}

func Increment(point int64) {
	mu.Lock()
	score += point
	mu.Unlock()
}
