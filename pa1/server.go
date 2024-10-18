package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Config variables
var PortsList []string
var LeaderPort string
var IsLeader bool
var NetworkDelay = 3

// Database variables
var DB map[int]int
var Mutex sync.RWMutex

var connectionMap = make(map[string]net.Conn)

func main() {
	ports := flag.String("ports", "", "Comma-separated list of ports")
	leaderPort := flag.String("leader", "", "Leader port")
	flag.Parse()

	if *ports == "" || *leaderPort == "" {
		log.Fatal("Please provide ports and leader port using the -ports and -leader flags")
	}

	initializeConfig(*ports, *leaderPort)
	initializeDatabase()
	go handleCLIInput()

	addr, err := net.ResolveTCPAddr("tcp", ":"+PortsList[0])
	if err != nil {
		log.Fatal("Address already in use: " + err.Error())
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("Address already in use: " + PortsList[0] + ": " + err.Error())
	}
	defer listener.Close()

	// Set SO_REUSEADDR
	file, err := listener.File()
	if err != nil {
		log.Fatal(err)
	}
	if err := syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatal(err)
	}

	// if IsLeader {
	// 	println("State: Leader")
	// } else {
	// 	println("State: Follower")
	// }

	go initalizeFollowerConnections()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			return
		}
		go handleConnection(conn)
	}
}

// Config functions
func initializeConfig(ports string, leaderPort string) {
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

// Main server functions
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
	time.Sleep(3 * time.Second)
	for _, port := range PortsList {
		if port == PortsList[0] {
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
	time.Sleep(time.Duration(NetworkDelay) * time.Second)

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
	if key%2 == 0 && IsLeader {
		forwardMessage(fmt.Sprintf("%s insert %d %d", clientID, key, value))
		println("Output: Forwarding to secondary server")
		return ""
	}
	Insert(key, value)
	fmt.Printf("Output: Successfully inserted key %d\n", key)
	return "Success"
}

func handleLookup(clientID string, key int) string {
	if key%2 == 0 && IsLeader {
		forwardMessage(fmt.Sprintf("%s lookup %d", clientID, key))
		println("Output: Forwarding to secondary server")
		return ""
	}
	value, ok := Lookup(key)
	if !ok {
		fmt.Printf("Output: NOT FOUND\n")
		return "NOT FOUND"
	}
	fmt.Printf("Output: %d\n", value)
	return strconv.Itoa(value)
}

func handleDump() string {
	dumpString := Dump()
	if IsLeader {
		fmt.Printf("Output: %s, ", dumpString)
		forwardMessage(fmt.Sprintf("%s dictionary", PortsList[0]))
	}
	return dumpString
}

func forwardMessage(message string) {
	var conn net.Conn
	if _, ok := connectionMap[PortsList[1]]; !ok {
		conn, err := net.Dial("tcp", ":"+PortsList[1])
		if err != nil {
			log.Fatal(err)
		}
		connectionMap[PortsList[1]] = conn
	}

	conn = connectionMap[PortsList[1]]

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

// Database functions
func initializeDatabase() {
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

	if IsLeader {
		result = "primary {"
	} else {
		result = "secondary {"
	}

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
