package main

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

// DeadlockRequest represents the request sent to the server
type DeadlockRequest struct {
	ProcessID int
}

// DeadlockResponse represents the response received from the server
type DeadlockResponse struct {
	Message        string
	DeadlockStatus string
	MessageCounts  map[int]int
	TotalMessages  int // New field
}

// AddProcessRequest represents a request to add a new process dynamically
type AddProcessRequest struct {
	ID        int
	Neighbors []int
}

// AddProcessResponse represents the response for adding a process
type AddProcessResponse struct {
	Message string
}

func main() {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Client: Error connecting to server:", err)
	}
	defer client.Close()

	// Optionally add a new process dynamically
	newProcess := AddProcessRequest{ID: 4, Neighbors: []int{1}}
	var addRes AddProcessResponse
	err = client.Call("DeadlockService.AddProcess", newProcess, &addRes)
	if err == nil {
		log.Println("Client:", addRes.Message)
	}

	// Deadlock detection requests
	testProcesses := []int{0, 2, 3}

	for _, processID := range testProcesses {
		startTime := time.Now()
		log.Printf("Client: Sending deadlock detection request from Process %d...", processID)

		req := DeadlockRequest{ProcessID: processID}
		var res DeadlockResponse

		err = client.Call("DeadlockService.DetectDeadlock", req, &res)
		if err != nil {
			log.Fatal("Client: RPC error:", err)
		}

		timeTaken := time.Since(startTime)
		log.Printf("Client: Response received from server for Process %d", processID)
		
		// Structured output
		fmt.Printf("\n[Result for Process %d]\n", processID)
		fmt.Printf(" ➤ Deadlock Status: %s\n", res.DeadlockStatus)
		fmt.Printf(" ➤ Message Counts: %v\n", res.MessageCounts)
		fmt.Printf(" ➤ Total Messages: %d\n", res.TotalMessages)
		fmt.Printf(" ➤ Time Taken: %v\n\n", timeTaken)
	}
}
