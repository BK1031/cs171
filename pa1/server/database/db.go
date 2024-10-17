package database

import (
	"fmt"
	"server/config"
	"sort"
	"sync"
)

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

func Dump() string {
	Mutex.RLock()
	defer Mutex.RUnlock()
	result := ""

	if config.IsLeader {
		result = "primary {"
	} else {
		result = "secondary {"
	}

	// Create a sorted slice of keys
	keys := make([]int, 0, len(DB))
	for key := range DB {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	for i, key := range keys {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("(%d, %d)", key, DB[key])
	}
	result += "}"
	return result
}
