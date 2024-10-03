package database

import "sync"

var DB map[int]int
var Mutex sync.RWMutex

func Initialize() {
	DB = make(map[int]int)
}

func Insert(key int, value int) {
	Mutex.Lock()
	defer Mutex.Unlock()
	DB[key] = value
}

func Lookup(key int) (int, bool) {
	Mutex.RLock()
	defer Mutex.RUnlock()
	value, ok := DB[key]
	return value, ok
}
