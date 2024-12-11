package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var failedLinks = make([]string, 0)

func main() {
	go StartServer()
	time.Sleep(2 * time.Second)
	for {
		var input string
		fmt.Print("Enter command: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		HandleCommand(input)
	}
}

func HandleCommand(command string) {
	if strings.HasPrefix(command, "failLink") {
		parts := strings.Split(command, " ")
		src := parts[1]
		dest := parts[2]
		failedLinks = append(failedLinks, fmt.Sprintf("%s-%s", src, dest))
	} else if strings.HasPrefix(command, "fixLink") {
		parts := strings.Split(command, " ")
		src := parts[1]
		dest := parts[2]
		failedLinks = remove(failedLinks, fmt.Sprintf("%s-%s", src, dest))
	} else if strings.HasPrefix(command, "failNode") {
		parts := strings.Split(command, " ")
		src := parts[1]
		println("failing node", src)
		SendMessage(src, "failNode")
	}
	if strings.HasPrefix(command, "create") {
		parts := strings.Split(command, " ")
		id := parts[1]
		SendAll(fmt.Sprintf("create %s", id))
	}
	if strings.HasPrefix(command, "query") {
		parts := strings.Split(command, " ")
		id := parts[1]
		query := strings.TrimPrefix(command, "query "+id+" ")
		SendAll(fmt.Sprintf("query %s %s", id, query))
	}
	if strings.HasPrefix(command, "choose") {
		parts := strings.Split(command, " ")
		id := parts[1]
		responseNumber := parts[2]
		SendAll(fmt.Sprintf("choose %s %s", id, responseNumber))
	}
	if strings.HasPrefix(command, "viewall") {
		SendAll("viewall")
		return
	}
	if strings.HasPrefix(command, "view") {
		parts := strings.Split(command, " ")
		id := parts[1]
		SendAll(fmt.Sprintf("view %s", id))
		return
	}
}

func StartServer() {
	http.HandleFunc("/", handleMessage)
	fmt.Printf("Network Server is listening on port %s\n", "7005")
	if err := http.ListenAndServe(":7005", nil); err != nil {
		fmt.Println(err)
		return
	}
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	message := string(body)
	println("Received:", message)
	parts := strings.SplitN(message, " ", 3)
	src := parts[0]
	dest := parts[1]
	messageContent := parts[2]
	ForwardMessage(src, dest, messageContent)
}

func SendMessage(dest string, message string) error {
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%s", dest),
		"text/plain",
		strings.NewReader(message),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func SendAll(message string) {
	for _, node := range []string{"7000", "7001", "7002"} {
		SendMessage(node, message)
	}
}

func ForwardAll(src string, message string) {
	for _, node := range []string{"7000", "7001", "7002"} {
		ForwardMessage(src, node, message)
	}
}

func ForwardMessage(src string, dest string, message string) {
	if CanForwardMessage(src, dest) {
		SendMessage(dest, message)
	}
}

func CanForwardMessage(src string, dest string) bool {
	for _, link := range failedLinks {
		if link == fmt.Sprintf("%s-%s", src, dest) {
			return false
		}
	}
	return true
}

func remove(slice []string, s string) []string {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func debugPrintFailedLinks() {
	fmt.Println("Failed Links:")
	for _, link := range failedLinks {
		fmt.Println(link)
	}
}
