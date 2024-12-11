package config

import "os"

var IsLeader = false
var LeaderPort = 7000

var Port = os.Getenv("PORT")
var ProxyPort = "7005"
var GeminiAPIKey = os.Getenv("GEMINI_API_KEY")

func Validate() {
	if GeminiAPIKey == "" {
		panic("GEMINI_API_KEY is not set")
	}
	if Port == "7000" {
		IsLeader = true
	}
}
