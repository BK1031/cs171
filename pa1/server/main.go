package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"server/config"
	"server/database"
	"strconv"
	"strings"
)

func main() {
	database.Initialize()
	server, err := net.Listen("tcp", ":"+config.Port)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	if config.IsLeader {
		log.Println("State: Leader")
	} else {
		log.Println("State: Follower")
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		// time.Sleep(time.Duration(config.NetworkDelay) * time.Second)

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from connection: %v", err)
			}
			return
		}

		message := string(buffer[:n])
		log.Printf("Cmd: %s", message)
		response := handleMessage(message)

		log.Printf("Output: %s", response)
		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Printf("Error writing to connection: %v", err)
			return
		}
	}
}

func handleMessage(message string) string {
	args := strings.Split(message, " ")
	command := args[0]

	if command == "insert" {
		if len(args) != 3 {
			return "missing parameters"
		}
		key, err := strconv.Atoi(args[2])
		if err != nil {
			return "invalid key"
		}
		value, err := strconv.Atoi(args[2])
		if err != nil {
			return "invalid value"
		}
		return handleInsert(key, value)
	} else if command == "lookup" {
		if len(args) != 2 {
			return "error"
		}
		key, err := strconv.Atoi(args[2])
		if err != nil {
			return "invalid key"
		}
		return handleLookup(key)
	} else if command == "dictionary" {
		return "todo"
	}
	return "invalid command"
}

func handleInsert(key int, value int) string {
	database.Insert(key, value)
	return fmt.Sprintf("Successfully inserted key %d", key)
}

func handleLookup(key int) string {
	value, ok := database.Lookup(key)
	if !ok {
		return "key not found"
	}
	return fmt.Sprintf("Successfully found key %d with value %d", key, value)
}
