package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"server/config"
	"server/database"
	"strconv"
	"strings"
	"time"
)

var connectionMap = make(map[string]net.Conn)

func main() {
	ports := flag.String("ports", "", "Comma-separated list of ports")
	leaderPort := flag.String("leader", "", "Leader port")
	flag.Parse()

	if *ports == "" || *leaderPort == "" {
		log.Fatal("Please provide ports and leader port using the -ports and -leader flags")
	}

	config.Initialize(*ports, *leaderPort)
	database.Initialize()
	go handleCLIInput()

	server, err := net.Listen("tcp", ":"+config.PortsList[0])
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	if config.IsLeader {
		println("State: Leader")
	} else {
		println("State: Follower")
	}

	time.Sleep(3 * time.Second)
	go initalizeFollowerConnections()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			return
		}
		go handleConnection(conn)
	}
}

func handleCLIInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		switch command {
		case "dictionary":
			handleDump()
		case "exit":
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Invalid command. Available commands: dictionary, exit")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading standard input:", err)
	}
}

func initalizeFollowerConnections() {
	for _, port := range config.PortsList {
		if port == config.PortsList[0] {
			continue
		}
		conn, err := net.Dial("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		connectionMap[port] = conn

		writer := bufio.NewWriter(conn)
		writer.WriteString(port + " ping\n")
		writer.Flush()
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		// time.Sleep(time.Duration(config.NetworkDelay) * time.Second)

		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			continue
		}

		message = strings.TrimSpace(message)
		clientID := strings.Split(message, " ")[0]

		if strings.Split(message, " ")[1] == "ping" {
			connectionMap[clientID] = conn
		} else if strings.Split(message, " ")[0] == "secondary" {
			fmt.Printf("%s\n", message)
			continue
		} else {
			trimmedMessage := strings.TrimPrefix(message, clientID+" ")
			fmt.Printf("Cmd: %s\n", trimmedMessage)
			response := handleMessage(message)
			if response != "" {
				sendResponse(clientID, response)
			}
		}
	}
}

func handleMessage(message string) string {
	args := strings.Split(message, " ")
	clientID := args[0]
	command := args[1]

	if command == "insert" {
		if len(args) != 4 {
			return "missing parameters"
		}
		key, err := strconv.Atoi(args[2])
		if err != nil {
			return "invalid key"
		}
		value, err := strconv.Atoi(args[3])
		if err != nil {
			return "invalid value"
		}
		return handleInsert(clientID, key, value)
	} else if command == "lookup" {
		if len(args) != 3 {
			return "error"
		}
		key, err := strconv.Atoi(args[2])
		if err != nil {
			return "invalid key"
		}
		return handleLookup(clientID, key)
	} else if command == "dictionary" {
		return handleDump()
	} else if command == "ping" {
		return ""
	}
	return "invalid command"
}

func handleInsert(clientID string, key int, value int) string {
	if key%2 == 0 && config.IsLeader {
		// Forward to secondary server
		forwardMessage(fmt.Sprintf("%s insert %d %d", clientID, key, value))
		println("Output: Forwarding to secondary server")
		return ""
	}
	// If key is odd, insert locally
	database.Insert(key, value)
	fmt.Printf("Output: Successfully inserted key %d\n", key)
	return "Success"
}

func handleLookup(clientID string, key int) string {
	if key%2 == 0 && config.IsLeader {
		// Forward to secondary server
		forwardMessage(fmt.Sprintf("%s lookup %d", clientID, key))
		println("Output: Forwarding to secondary server")
		return ""
	}
	// If key is odd, lookup locally
	value, ok := database.Lookup(key)
	if !ok {
		fmt.Printf("Output: NOT FOUND\n")
		return "NOT FOUND"
	}
	fmt.Printf("Output: %d\n", value)
	return strconv.Itoa(value)
}

func handleDump() string {
	dumpString := database.Dump()
	if config.IsLeader {
		fmt.Printf("Output: %s, ", dumpString)
		forwardMessage(fmt.Sprintf("%s dictionary", config.PortsList[0]))
	}
	return dumpString
}

func forwardMessage(message string) {
	var conn net.Conn
	if _, ok := connectionMap[config.PortsList[1]]; !ok {
		conn, err := net.Dial("tcp", ":"+config.PortsList[1])
		if err != nil {
			log.Fatal(err)
		}
		connectionMap[config.PortsList[1]] = conn
	}

	conn = connectionMap[config.PortsList[1]]

	writer := bufio.NewWriter(conn)
	writer.WriteString(message + "\n")
	writer.Flush()
}

func sendResponse(clientID string, response string) {
	conn, ok := connectionMap[clientID]
	if !ok {
		return
	}
	writer := bufio.NewWriter(conn)
	writer.WriteString(response + "\n")
	writer.Flush()
}

func dumpConnectionMap() {
	for clientID, conn := range connectionMap {
		println("Client " + clientID + " connected from " + conn.RemoteAddr().String())
	}
}
