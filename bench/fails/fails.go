package fails

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

var mu sync.RWMutex
var msgs []string
var isCritical bool

func Get() []string {
	mu.RLock()
	allMsgs := msgs[:]
	mu.RUnlock()
	return allMsgs
}

func GetUnique() []string {
	mu.RLock()
	allMsgs := msgs[:]
	mu.RUnlock()

	sort.Strings(allMsgs)
	var tmp string
	retMsgs := make([]string, 0)

	// 適当にuniqする
	for _, m := range allMsgs {
		if tmp != m {
			tmp = m
			retMsgs = append(retMsgs, m)
		}
	}
	return retMsgs
}

func Add(msg string, err error) {
	mu.Lock()
	msgs = append(msgs, msg)
	mu.Unlock()

	if err != nil {
		msg += " error: " + err.Error()
	}
	fmt.Fprintln(os.Stderr, msg)
}

func Critical(msg string, err error) {
	Add(msg+" (critical)", err)
	isCritical = true
}

func GetIsCritical() bool {
	return isCritical
}

type Logger struct {
	Prefix string
}

func (l *Logger) Add(msg string, err error) {
	Add(l.Prefix+msg, err)
}

func (l *Logger) Critical(msg string, err error) {
	Critical(msg, err)
}
