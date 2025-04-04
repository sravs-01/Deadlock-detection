package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	ProbeInterval   time.Duration
	ChanBufferSize  int
	SimulationTime  time.Duration
	TimeoutDuration time.Duration
}

type Probe struct {
	SessionID int
	Initiator int
	Sender    int
	Visited   []int
	AckChan   chan bool
}

type Process struct {
	ID            int
	successors    []*Process
	probeChan     chan Probe
	visited       map[int]bool
	config        Config
	mu            sync.Mutex
	wg            sync.WaitGroup
	stopChan      chan struct{}
	messagesSent  int
	deadlockCache map[int]bool
}

func NewProcess(id int, config Config) *Process {
	return &Process{
		ID:            id,
		probeChan:     make(chan Probe, config.ChanBufferSize),
		visited:       make(map[int]bool),
		deadlockCache: make(map[int]bool),
		config:        config,
		stopChan:      make(chan struct{}),
	}
}

func (p *Process) Run(ctx context.Context) {
	// Log the start of deadlock detection for this process.
	log.Printf("Starting deadlock detection from Process %d, Visited: [%d]", p.ID, p.ID)
	p.wg.Add(1)
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.ProbeInterval)
	defer ticker.Stop()

	p.wg.Add(1)
	go p.handleProbes(ctx)

	for {
		select {
		case <-ctx.Done():
			close(p.stopChan)
			return
		case <-ticker.C:
			p.mu.Lock()
			if len(p.successors) > 0 {
				// Create a new session ID based on timestamp.
				sessionID := int(time.Now().UnixNano())
				p.visited[sessionID] = true
				probe := Probe{
					SessionID: sessionID,
					Initiator: p.ID,
					Sender:    p.ID,
					Visited:   []int{p.ID},
					AckChan:   make(chan bool, len(p.successors)),
				}
				// Log initiation of probe including visited list.
				log.Printf("Probe sent from Process %d to Process %d for origin %d, Visited: %v", p.ID, p.successors[0].ID, p.ID, probe.Visited)
				p.mu.Unlock()
				p.sendProbe(probe, ctx)
			} else {
				p.mu.Unlock()
			}
		}
	}
}

func (p *Process) handleProbes(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case probe := <-p.probeChan:
			p.mu.Lock()
			//cknowledge receipt of the probe.
			select {
			case probe.AckChan <- true:
			default:
			}

			// Check for deadlock: if the probe returns to its initiator.
			if probe.Initiator == p.ID && !p.deadlockCache[probe.SessionID] {
				log.Printf("Deadlock detected involving process %d, Visited: %v", p.ID, probe.Visited)
				p.deadlockCache[probe.SessionID] = true
				p.mu.Unlock()
				log.Println("Deadlock confirmed!")
				continue
			}
			// Forward the probe if not already processed.
			if !p.visited[probe.SessionID] && len(p.successors) > 0 {
				p.visited[probe.SessionID] = true
				newProbe := Probe{
					SessionID: probe.SessionID,
					Initiator: probe.Initiator,
					Sender:    p.ID,
					Visited:   append(probe.Visited, p.ID),
					AckChan:   make(chan bool, len(p.successors)),
				}
				// Log forwarding of probe with the updated visited list.
				log.Printf("Probe sent from Process %d to Process %d for origin %d, Visited: %v", p.ID, p.successors[0].ID, probe.Initiator, newProbe.Visited)
				p.mu.Unlock()
				p.sendProbe(newProbe, ctx)
			} else {
				p.mu.Unlock()
			}
		}
	}
}

func (p *Process) sendProbe(probe Probe, ctx context.Context) {
	select {
	case <-p.stopChan:
		return
	default:
	}

	p.mu.Lock()
	// Copy the list of successors to avoid race conditions.
	successors := make([]*Process, len(p.successors))
	copy(successors, p.successors)
	p.mu.Unlock()

	deadline := time.After(p.config.TimeoutDuration)
	for _, succ := range successors {
		select {
		case succ.probeChan <- probe:
			p.mu.Lock()
			p.messagesSent++
			p.mu.Unlock()
		case <-deadline:
			return
		case <-ctx.Done():
			return
		}
	}

	// Wait for acknowledgments in a separate goroutine.
	go func(expected int) {
		acks := 0
		for acks < expected {
			select {
			case <-probe.AckChan:
				acks++
			case <-deadline:
				return
			case <-ctx.Done():
				return
			}
		}
	}(len(successors))
}

func main() {
	config := Config{
		ProbeInterval:   10 * time.Second,
		ChanBufferSize:  20,
		SimulationTime:  30 * time.Second,
		TimeoutDuration: 8 * time.Second,
	}

	// Create processes with cyclic dependencies: p0 → p1 → p2 → p3 → p0.
	p0 := NewProcess(0, config)
	p1 := NewProcess(1, config)
	p2 := NewProcess(2, config)
	p3 := NewProcess(3, config)

	p0.successors = []*Process{p1}
	p1.successors = []*Process{p2}
	p2.successors = []*Process{p3}
	p3.successors = []*Process{p0}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for shutdown signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go p0.Run(ctx)
	go p1.Run(ctx)
	go p2.Run(ctx)
	go p3.Run(ctx)

	select {
	case <-time.After(config.SimulationTime):
		log.Println("Simulation completed")
	case <-sigChan:
		log.Println("Shutdown signal received")
	}
	cancel()

	p0.wg.Wait()
	p1.wg.Wait()
	p2.wg.Wait()
	p3.wg.Wait()

	fmt.Printf("Final Metrics: P0: %d messages sent, P1: %d messages sent, P2: %d messages sent, P3: %d messages sent\n",
		p0.messagesSent, p1.messagesSent, p2.messagesSent, p3.messagesSent)
}
