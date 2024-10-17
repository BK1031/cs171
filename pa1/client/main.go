package main

import (
	"bufio"
	"client/config"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var serverMap = make(map[string]net.Conn)

func main() {
	ports := flag.String("ports", "", "Comma-separated list of ports")
	flag.Parse()

	if *ports == "" {
		log.Fatal("Please provide ports using the -ports flag")
	}

	config.Initialize(*ports)
	println("Client ID: ", config.ClientID)
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
			sendMessage(config.PortsList[0], command)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading standard input:", err)
	}
}

func initalizeConnections() {
	for _, port := range config.PortsList {
		conn, err := net.Dial("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		serverMap[port] = conn

		sendMessage(port, "ping")
		go listenForReponse(port)
	}
}

// sendMessage sends a message to the specified tcp port
// if a connection to that port does not exist, it creates a new connection
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
	writer.WriteString(fmt.Sprintf("%s %s\n", config.ClientID, message))
	writer.Flush()
}

// listenForResponse listens for responses from the specified tcp port
// if a connection to that port does not exist, it creates a new connection
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
