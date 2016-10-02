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

func Add(msg string) string {
	fmt.Fprintln(os.Stderr, msg) // TODO: デバッグモードのみにする

	mu.Lock()
	msgs = append(msgs, msg)
	mu.Unlock()

	return msg
}

func Critical() {
	isCritical = true
}

func GetIsCritical() bool {
	return isCritical
}
