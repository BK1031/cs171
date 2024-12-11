package database

import "fmt"

var DB map[string]string

func Initialize() {
	DB = make(map[string]string)
	ResetResponses()
}

func Get(key string) string {
	return DB[key]
}

func Set(key string, value string) {
	DB[key] = value
}

func CreateContext(key string) {
	if _, ok := DB[key]; !ok {
		Set(key, "")
	}
}

func PrintContext(key string) {
	println(fmt.Sprintf("-------- CONTEXT %s --------", key))
	println(Get(key))
	println("-------------------------------")
}

func PrintContexts() {
	println("======= ALL CONTEXTS =======")
	for k := range DB {
		PrintContext(k)
	}
	println("================================")
}

var Responses map[string]string

func ResetResponses() {
	Responses = make(map[string]string)
}

func DebugDumpResponses() {
	println("-------- DEBUG DUMP RESPONSES --------")
	for k, v := range Responses {
		fmt.Printf("(%s) Response %s: %s\n", k, k, v)
	}
}
