package queue

import (
	"log"
	"sort"
	"time"

	"github.com/6a/blade-ii-game-server/internal/connection"
)

// BufferSize is the size of each message queue's buffer
const BufferSize = 32

var pollTime = 500 * time.Millisecond

// Queue is a wrapper for the matchmaking queue
type Queue struct {
	index      []uint64
	nextIndex  uint64
	clients    map[uint64]*Client
	register   chan *Client
	unregister chan uint64
	broadcast  chan connection.Message
	commands   chan Command
}

// Start the queue's internal update loop
func (queue *Queue) Start() {
	queue.index = make([]uint64, 0)
	queue.clients = make(map[uint64]*Client)
	queue.register = make(chan *Client, BufferSize)
	queue.unregister = make(chan uint64, BufferSize)
	queue.broadcast = make(chan connection.Message, BufferSize)
	queue.commands = make(chan Command, BufferSize)

	for {
		start := time.Now()

		// Perform all pending tasks
		toRemove := make([]uint64, 0)
		for len(queue.register)+len(queue.unregister)+len(queue.broadcast) > 0 {
			select {
			case client := <-queue.register:
				queue.clients[queue.nextIndex] = client
				client.ID = queue.nextIndex
				queue.index = append(queue.index, queue.nextIndex)
				queue.nextIndex++
				log.Printf("Client [%s] joined the matchmaking queue. Total clients: %v", client.UID, len(queue.clients))
				client.SendMessage(connection.NewMessage(connection.WSMTText, connection.WSCAuthSuccess, "Added to matchmaking queue"))
				break
			case clientid := <-queue.unregister:
				toRemove = append(toRemove, clientid)
				break
			case message := <-queue.broadcast:
				log.Println("Broadcast:")
				log.Println(message)
				for _, client := range queue.clients {
					client.SendMessage(message)
				}
				break
			case command := <-queue.commands:
				queue.processCommand(command)
			}
		}

		// Sort the pending removal slice in ascending order so that we can iterate over the queue index once
		sort.Slice(toRemove, func(i, j int) bool { return toRemove[i] < toRemove[j] })

		// Store a value to use so that we can retain our position when iterating down the queue index
		indexIterator := len(queue.index) - 1

		// Remove any clients that are pending removal
		for index := len(toRemove) - 1; index >= 0; index-- {
			// Remove from the queue (map)
			removalIndex := toRemove[index]
			deletedClientUID := queue.clients[removalIndex].UID
			delete(queue.clients, removalIndex)

			// Remove from the queue index (slice)
			for indexIterator >= 0 {
				if queue.index[index] == removalIndex {
					queue.index = append(queue.index[:indexIterator], queue.index[indexIterator+1:]...)
					break
				}

				indexIterator--
			}

			log.Printf("Client [%s] left the matchmaking queue. Total clients: %v", deletedClientUID, len(queue.clients))
		}

		// Tick all clients
		for _, client := range queue.clients {
			client.Tick()
		}

		// Perform matchmaking
		// matchedClients := queue.MatchMake()

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
	queue.unregister <- client.ID
}

// Broadcast sends a message to all connected clients
func (queue *Queue) Broadcast(message connection.Message) {
	queue.broadcast <- message
}

// MatchMake matches people in the queue based on various factors, such as mmr, latency, and waiting time
func (queue *Queue) MatchMake() (pairs []ClientPair) {
	// Matchmaking algo goes here. For now just return all the pairs we can put into pairs of two

	pairs = make([]ClientPair, 0)

	for len(queue.clients) > 1 {
		pairs = append(pairs[:2])
	}

	return pairs
}

func (queue *Queue) processCommand(command Command) {
	log.Printf("Processing command of type [ %v ] with data [ %v ]", command.Type, command.Data)
}
