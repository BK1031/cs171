package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"server/config"
	"server/consensus"
	"server/database"
	"server/llm"
	"strconv"
	"strings"
)

func main() {
	config.Validate()
	database.Initialize()
	llm.Initialize("gemini-1.5-flash")

	StartServer()
}

func StartServer() {
	http.HandleFunc("/", handleMessage)

	fmt.Printf("Server %s is listening on port %s\n", config.Port, config.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", config.Port), nil); err != nil {
		fmt.Println(err)
	}
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	message := string(body)

	if strings.HasPrefix(message, "failNode") {
		FailNode()
	}
	if strings.HasPrefix(message, "create") {
		id := strings.Split(message, " ")[1]
		if config.IsLeader {
			consensus.PrepareProposal(fmt.Sprintf("create-%s", id), id)
			database.CreateContext(id)
			fmt.Printf("NEW CONTEXT %s\n", id)
			for _, peer := range []string{"7001", "7002"} {
				SendMessage(peer, fmt.Sprintf("accept-create %s", id))
			}
		}
	}
	if strings.HasPrefix(message, "query") {
		id := strings.Split(message, " ")[1]
		query := strings.TrimPrefix(message, "query "+id+" ")
		eq := database.Get(id)
		database.Set(id, fmt.Sprintf("%s\nQuery: %s", eq, query))
		q := database.Get(id)
		fmt.Printf("NEW QUERY on %s with %s\n", id, q)
		response, err := llm.Query(q)
		if err != nil {
			println("error querying", err)
			return
		}
		if config.IsLeader {
			database.Responses[config.Port] = response
			fmt.Printf("(%s) Response %s: %s\n", id, config.Port, response)
		} else {
			SendMessage(fmt.Sprintf("%d", config.LeaderPort), fmt.Sprintf("response %s %s %s", id, config.Port, response))
		}
	}
	if strings.HasPrefix(message, "response") {
		if config.IsLeader {
			id := strings.Split(message, " ")[1]
			port := strings.Split(message, " ")[2]
			response := strings.TrimPrefix(message, "response "+id+" "+port+" ")
			database.Responses[port] = response
			fmt.Printf("(%s) Response %s: %s\n", id, port, response)
		}
	}
	if strings.HasPrefix(message, "choose") {
		id := strings.Split(message, " ")[1]
		port := strings.Split(message, " ")[2]
		response := database.Responses[port]
		if config.IsLeader {
			consensus.PrepareProposal(fmt.Sprintf("choose-%s", id), response)
			database.ResetResponses()
			eq := database.Get(id)
			database.Set(id, fmt.Sprintf("%s\nAnswer: %s", eq, response))
			fmt.Printf("CHOSEN ANSWER on %s with %s\n", id, response)
			for _, peer := range []string{"7001", "7002"} {
				SendMessage(peer, fmt.Sprintf("accept-choose %s %s", id, response))
			}
			database.PrintContext(id)
		}
	}
	if strings.HasPrefix(message, "viewall") {
		database.PrintContexts()
		return
	}
	if strings.HasPrefix(message, "view") {
		id := strings.Split(message, " ")[1]
		database.PrintContext(id)
		return
	}
	if strings.HasPrefix(message, "accept-create") {
		id := strings.Split(message, " ")[1]
		database.CreateContext(id)
		fmt.Printf("NEW CONTEXT %s\n", id)
		fmt.Println("ACK")
	}
	if strings.HasPrefix(message, "accept-choose") {
		parts := strings.Split(message, " ")
		id := parts[1]
		response := strings.Join(parts[2:], " ")
		database.ResetResponses()
		eq := database.Get(id)
		database.Set(id, fmt.Sprintf("%s\nAnswer: %s", eq, response))
		fmt.Printf("CHOSEN ANSWER on %s with %s\n", id, response)
		fmt.Println("ACK")
	}
	if strings.HasPrefix(message, "prepare") {
		parts := strings.Split(message, " ")
		instanceID := parts[1]
		proposalNumber, _ := strconv.Atoi(parts[2])
		leaderID := parts[3]

		proposal := consensus.ProposalID{
			Number:   proposalNumber,
			LeaderID: leaderID,
		}
		accepted := consensus.HandlePrepare(instanceID, proposal)
		if accepted {
			SendMessage(leaderID, fmt.Sprintf("promise %s %d %s",
				instanceID,
				proposalNumber,
				leaderID,
			))
		}
	}
}

func SendMessage(dest string, message string) error {
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%s", config.ProxyPort),
		"text/plain",
		strings.NewReader(fmt.Sprintf("%s %s %s", config.Port, dest, message)),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func FailNode() {
	println("Received fail node message")
	os.Exit(1)
}
