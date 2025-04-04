package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sort"
	"sync"
)

// DeadlockRequest represents the RPC request for deadlock detection
type DeadlockRequest struct {
	ProcessID int
}

// DeadlockResponse contains the result of deadlock detection
type DeadlockResponse struct {
	Message        string
	DeadlockStatus string
	MessageCounts  map[int]int
	TotalMessages  int // New field to store total message count
}

// AddProcessRequest represents a request to add a new process dynamically
type AddProcessRequest struct {
	ID        int
	Neighbors []int
}

// AddProcessResponse contains the response for adding a process
type AddProcessResponse struct {
	Message string
}

// Process represents a node in the system
type Process struct {
	ID        int
	Neighbors []int
}

// DeadlockService manages the processes and detection logic
type DeadlockService struct {
	processes           []*Process
	mutex               sync.Mutex
	globalMessageCounter int
}

// DetectDeadlock handles RPC requests for deadlock detection
func (ds *DeadlockService) DetectDeadlock(req DeadlockRequest, res *DeadlockResponse) error {
	if req.ProcessID < 0 || req.ProcessID >= len(ds.processes) {
		res.Message = "Invalid Process ID"
		return nil
	}

	start := req.ProcessID
	log.Printf("Starting deadlock detection from Process %d, Visited: []", start)

	visited := make(map[int]bool)
	messageCounts := make(map[int]int)
	localMessageCounter := 0 // Track messages for this request only

	var deadlockDetected bool
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ds.mutex.Lock()
		defer ds.mutex.Unlock()
		deadlockDetected = ds.runDDFS(start, start, visited, messageCounts, &localMessageCounter)
	}()

	wg.Wait() // Ensure detection is complete before responding

	res.Message = fmt.Sprintf("Deadlock detection completed for Process %d", req.ProcessID)
	res.MessageCounts = messageCounts
	res.TotalMessages = localMessageCounter // Use request-specific counter

	if deadlockDetected {
		res.DeadlockStatus = "Deadlock confirmed!"
		log.Printf("Deadlock detected involving Process %d, Visited: %v", start, getSortedVisitedList(visited))
	} else {
		res.DeadlockStatus = "No deadlock detected."
	}
	return nil
}

// runDDFS performs Distributed Depth-First Search (DDFS) to detect deadlocks
func (ds *DeadlockService) runDDFS(origin, current int, visited map[int]bool, messageCounts map[int]int, localCounter *int) bool {
	visited[current] = true

	for _, neighbor := range ds.processes[current].Neighbors {
		messageCounts[origin]++
		*localCounter++ // Increment request-specific counter

		log.Printf("Probe sent from Process %d to Process %d for origin %d, Visited: %v", current, neighbor, origin, getSortedVisitedList(visited))

		if neighbor == origin {
			return true // Deadlock detected
		}

		if !visited[neighbor] {
			if ds.runDDFS(origin, neighbor, visited, messageCounts, localCounter) {
				return true
			}
		}
	}
	return false
}

// getSortedVisitedList formats the visited map into a sorted slice for logging
func getSortedVisitedList(visited map[int]bool) []int {
	keys := make([]int, 0, len(visited))
	for key := range visited {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}

// AddProcess allows dynamic addition of processes via RPC
func (ds *DeadlockService) AddProcess(req AddProcessRequest, res *AddProcessResponse) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	newProcess := &Process{ID: req.ID, Neighbors: req.Neighbors}
	ds.processes = append(ds.processes, newProcess)

	res.Message = fmt.Sprintf("Process %d added successfully!", req.ID)
	log.Printf("Process %d added with neighbors %v", req.ID, req.Neighbors)
	return nil
}

func main() {
	// Define initial processes (circular dependency for testing deadlock)
	processes := []*Process{
		{ID: 0, Neighbors: []int{1}},
		{ID: 1, Neighbors: []int{2}},
		{ID: 2, Neighbors: []int{3}},
		{ID: 3, Neighbors: []int{0}}, // Circular dependency
	}

	service := &DeadlockService{processes: processes}

	// Start RPC server
	rpc.Register(service)
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()
	log.Println("Deadlock detection server started on :1234")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
