package consensus

import (
	"fmt"
	"net/http"
	"server/config"
	"strings"
	"sync"
)

type ProposalID struct {
	Number   int
	LeaderID string
}

type Instance struct {
	ID            string
	AcceptedValue string
	AcceptedID    ProposalID
	PromisedID    ProposalID
}

var (
	instances             = make(map[string]*Instance)
	instancesMutex        sync.RWMutex
	currentProposalNumber = 0
)

func (p ProposalID) GreaterThan(other ProposalID) bool {
	if p.Number != other.Number {
		return p.Number > other.Number
	}
	return p.LeaderID > other.LeaderID
}

func PrepareProposal(instanceID string, value string) {
	currentProposalNumber++
	proposal := ProposalID{
		Number:   currentProposalNumber,
		LeaderID: config.Port,
	}

	fmt.Printf("PREPARE %d from %s for instance %s\n",
		proposal.Number,
		config.Port,
		instanceID)

	// Send prepare messages to all nodes
	for _, port := range []string{"7001", "7002"} {
		SendPrepare(port, instanceID, proposal)
	}
}

func HandlePrepare(instanceID string, proposal ProposalID) bool {
	instancesMutex.Lock()
	defer instancesMutex.Unlock()

	instance, exists := instances[instanceID]
	if !exists {
		instances[instanceID] = &Instance{ID: instanceID}
		instance = instances[instanceID]
	}

	if proposal.GreaterThan(instance.PromisedID) {
		instance.PromisedID = proposal
		fmt.Printf("PROMISE %d to %s for instance %s\n",
			proposal.Number,
			proposal.LeaderID,
			instanceID)
		return true
	}
	return false
}

func Accept(instanceID string, proposal ProposalID, value string) {
	instancesMutex.Lock()
	defer instancesMutex.Unlock()

	instance, exists := instances[instanceID]
	if !exists {
		instances[instanceID] = &Instance{ID: instanceID}
		instance = instances[instanceID]
	}

	if !instance.PromisedID.GreaterThan(proposal) {
		instance.AcceptedID = proposal
		instance.AcceptedValue = value
		fmt.Printf("ACCEPTED %d from %s for instance %s with value: %s\n",
			proposal.Number,
			proposal.LeaderID,
			instanceID,
			value)
	}
}

func SendPrepare(port string, instanceID string, proposal ProposalID) {
	// Format: "prepare {instanceID} {proposalNumber} {leaderID}"
	message := fmt.Sprintf("prepare %s %d %s",
		instanceID,
		proposal.Number,
		proposal.LeaderID,
	)

	// Send prepare message to the specified port
	err := SendMessage(port, message)
	if err != nil {
		fmt.Printf("Error sending prepare to %s: %v\n", port, err)
		return
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
