package net

import (
	"log"
	"time"
)

// BufferSize is the size of each message queue's buffer
const BufferSize = 32

var pollTime = 200 * time.Millisecond

// Queue is a wrapper for the matchmaking queue
type Queue struct {
	index      []uint64
	nextIndex  uint64
	clients    map[uint64]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	commands   chan QueueCommand
}

// Start the queue's internal update loop
func (queue *Queue) Start() {
	queue.index = make([]uint64, 0)
	queue.clients = make(map[uint64]*Client)
	queue.register = make(chan *Client, BufferSize)
	queue.unregister = make(chan *Client, BufferSize)
	queue.broadcast = make(chan Message, BufferSize)
	queue.commands = make(chan QueueCommand, BufferSize)

	for {
		start := time.Now()

		// Perform all pending tasks
		toDelete := make([]uint64, 0)
		for len(queue.register)+len(queue.unregister)+len(queue.broadcast) > 0 {
			select {
			case client := <-queue.register:
				queue.clients[queue.nextIndex] = client
				client.ID = queue.nextIndex
				queue.index = append(queue.index, queue.nextIndex)
				queue.nextIndex++
				log.Println("Size of Queue: ", len(queue.clients))
				break
			case client := <-queue.unregister:
				toDelete = append(toDelete, client.ID)
				log.Println("Size of Queue: ", len(queue.clients))
				break
			case message := <-queue.broadcast:
				log.Println("Sending message to all clients in Pool")
				for _, client := range queue.clients {
					client.SendMessage(message)
				}
				break
			case command := <-queue.commands:
				queue.processCommand(command)
			}
		}

		// Delete any clients that are pending deletion
		for _, clientID := range toDelete {
			// Should be safe to delete as deletion can only occurr here, in a single thread
			delete(queue.clients, clientID)
		}

		// Tick all clients
		for _, client := range queue.clients {
			client.Tick()
		}

		// Wait til next iteration if the time taken is less than the designated poll time
		elapsed := time.Now().Sub(start)
		remainingPollTime := pollTime - elapsed
		if remainingPollTime > 0 {
			time.Sleep(remainingPollTime)
		}
	}
}

// Add a client to the register queue, to be added next cycle
func (queue *Queue) Add(client *Client) {
	queue.register <- client
}

// Remove adds a client to the unregister queue, to be removed next cycle
func (queue *Queue) Remove(client *Client) {
	queue.unregister <- client
}

// Broadcast sends a message to all connected clients
func (queue *Queue) Broadcast(message Message) {
	queue.broadcast <- message
}

// MatchMake matches people in the queue based on various factors, such as mmr, latency, and waiting time
func (queue *Queue) MatchMake() {
	// Matchmaking algo goes here. For now just return all the pairs we can

	// var pairs = []ClientPair{}

}

func (queue *Queue) processCommand(command QueueCommand) {
	log.Printf("Processing command of type [ %v ] with data [ %v ]", command.Type, command.Data)
}
