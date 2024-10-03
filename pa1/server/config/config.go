package config

import "os"

var Port = os.Getenv("PORT")
var LeaderPort = os.Getenv("LEADER_PORT")

var IsLeader = Port == LeaderPort

var NetworkDelay = 3
