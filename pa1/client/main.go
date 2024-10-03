package main

import (
	"bufio"
	"client/config"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":"+config.LeaderPort)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	writer.WriteString("insert 1")
	writer.Flush()

	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	log.Println(response)
}
