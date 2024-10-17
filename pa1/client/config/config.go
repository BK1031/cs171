package config

import (
	"log"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

var ClientID = ""
var PortsList []string

func Initialize(ports string) {
	rand.Seed(uint64(time.Now().UnixNano()))
	ClientID = generateRandomString(5)
	if ports == "" {
		log.Fatal("Ports must be provided")
	}
	PortsList = strings.Split(ports, ",")
	for _, port := range PortsList {
		if port == "" {
			log.Fatal("invalid port")
		}
	}
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
