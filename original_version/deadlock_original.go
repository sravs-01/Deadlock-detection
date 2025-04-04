package main

import (
	"fmt"
)

// Process represents a process in the system.
type Process struct {
	id            int
	waitingFor    []int
	receivedProbes map[int]bool
}

// sendProbe sends a probe message to a dependent process.
func sendProbe(probeSender int, origin int, target int, processes map[int]*Process) bool {
	fmt.Printf("Probe sent from Process %d to Process %d for origin %d\n", probeSender, target, origin)

	// If the target process is waiting for other processes, forward the probe.
	if len(processes[target].waitingFor) > 0 {
		for _, dependent := range processes[target].waitingFor {
			// Check if the probe has returned to the origin.
			if dependent == origin {
				fmt.Println("Deadlock detected involving process", origin)
				return true
			}
			if !processes[target].receivedProbes[dependent] {
				processes[target].receivedProbes[dependent] = true
				if sendProbe(target, origin, dependent, processes) {
					return true
				}
			}
		}
	}
	return false
}

// detectDeadlock initiates deadlock detection from each process.
func detectDeadlock(processes map[int]*Process) {
	for id, process := range processes {
		fmt.Println("\nStarting deadlock detection from Process", id)
		process.receivedProbes = make(map[int]bool) // Reset received probes.
		for _, dependent := range process.waitingFor {
			if sendProbe(id, id, dependent, processes) {
				fmt.Println("Deadlock confirmed!")
				return
			}
		}
	}
	fmt.Println("No deadlock detected.")
}

func main() {
	// Initialize processes with a cyclic dependency.
	processes := map[int]*Process{
		1: {id: 1, waitingFor: []int{2}, receivedProbes: make(map[int]bool)},
		2: {id: 2, waitingFor: []int{3}, receivedProbes: make(map[int]bool)},
		3: {id: 3, waitingFor: []int{1}, receivedProbes: make(map[int]bool)}, // Cycle: 1 → 2 → 3 → 1.
	}

	// Detect deadlock using the Chandy-Misra-Haas algorithm.
	detectDeadlock(processes)
}
