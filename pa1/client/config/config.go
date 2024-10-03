package config

import "os"

var LeaderPort = os.Getenv("LEADER_PORT")
