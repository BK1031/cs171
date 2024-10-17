package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
)

var serverMap = make(map[string]net.Conn)
var ClientID = ""
var PortsList []string

func main() {
	ports := flag.String("ports", "", "Comma-separated list of ports")
	flag.Parse()

	if *ports == "" {
		log.Fatal("Please provide ports using the -ports flag")
	}

	Initialize(*ports)
	// println("Client ID: ", ClientID)
	initalizeConnections()

	go handleCLIInput()

	for {
	}
}

func handleCLIInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		switch command {
		case "exit":
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			sendMessage(PortsList[0], command)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading standard input:", err)
	}
}

func initalizeConnections() {
	for _, port := range PortsList {
		conn, err := net.Dial("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		serverMap[port] = conn

		sendMessage(port, "ping")
		go listenForReponse(port)
	}
}

func sendMessage(port string, message string) {
	var conn net.Conn
	if _, ok := serverMap[port]; !ok {
		conn, err := net.Dial("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		serverMap[port] = conn
	}

	conn = serverMap[port]

	writer := bufio.NewWriter(conn)
	writer.WriteString(fmt.Sprintf("%s %s\n", ClientID, message))
	writer.Flush()
}

func listenForReponse(port string) {
	var conn net.Conn
	if _, ok := serverMap[port]; !ok {
		conn, err := net.Dial("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		serverMap[port] = conn
	}

	conn = serverMap[port]

	reader := bufio.NewReader(conn)
	for {
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from connection to port %s: %v", port, err)
			return
		}

		trimmedResponse := strings.TrimSpace(response)
		fmt.Printf("Output: %s\n", trimmedResponse)
	}
}

func Initialize(ports string) {
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
