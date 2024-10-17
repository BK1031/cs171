package config

import (
	"log"
	"strings"
)

var PortsList []string
var LeaderPort string
var IsLeader bool
var NetworkDelay = 3

func Initialize(ports string, leaderPort string) {
	if ports == "" || leaderPort == "" {
		log.Fatal("Ports and LeaderPort must be provided")
	}
	PortsList = strings.Split(ports, ",")
	for _, port := range PortsList {
		if port == "" {
			log.Fatal("invalid port")
		}
	}
	LeaderPort = leaderPort
	IsLeader = LeaderPort == PortsList[0]
}
