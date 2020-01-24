package queue

import (
	"log"
	"sort"
	"time"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

// BufferSize is the size of each message queue's buffer
const BufferSize = 32

// readyCheckTime is how long to wait when a ready check is sent
const readyCheckTime = time.Second * 15

// How frequently to update the matchmaking queue (minimum wait between iterations)
var pollTime = 500 * time.Millisecond

// How frequently to update the match found goroutine (minimum wait between iterations)
var pollTimeMatchFound = 500 * time.Millisecond

// Queue is a wrapper for the matchmaking queue
type Queue struct {
	index      []uint64
	nextIndex  uint64
	clients    map[uint64]*Client
	register   chan *Client
	unregister chan uint64
	broadcast  chan protocol.Message
	commands   chan Command
}

// Start the queue's internal update loop
func (queue *Queue) Start() {
	queue.index = make([]uint64, 0)
	queue.clients = make(map[uint64]*Client)
	queue.register = make(chan *Client, BufferSize)
	queue.unregister = make(chan uint64, BufferSize)
	queue.broadcast = make(chan protocol.Message, BufferSize)
	queue.commands = make(chan Command, BufferSize)

	for {
		start := time.Now()

		// Perform all pending tasks
		toRemove := make([]uint64, 0)
		for len(queue.register)+len(queue.unregister)+len(queue.broadcast) > 0 {
			select {
			case client := <-queue.register:
				queue.clients[queue.nextIndex] = client
				client.QueueID = queue.nextIndex
				queue.index = append(queue.index, queue.nextIndex)
				queue.nextIndex++
				log.Printf("Client [%s] joined the matchmaking queue. Total clients: %v", client.PublicID, len(queue.clients))
				client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthSuccess, "Added to matchmaking queue"))
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
			deletedClientUID := queue.clients[removalIndex].PublicID
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

		// Perform matchmaking - these pairs will be dealt with in the other goroutine loop
		matchedClients := queue.matchMake()
		for _, clientPair := range matchedClients {
			clientPair.SendMatchStartMessage()
			go queue.pollReadyCheck(clientPair)
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
	queue.unregister <- client.QueueID
}

// Broadcast sends a message to all connected clients
func (queue *Queue) Broadcast(message protocol.Message) {
	queue.broadcast <- message
}

func (queue *Queue) pollReadyCheck(clientPair ClientPair) {
	start := time.Now()

	for {
		timedOut := time.Now().Sub(clientPair.ReadyStart) > readyCheckTime

		if (clientPair.C1.Ready && clientPair.C2.Ready) || timedOut {
			client1ReadyValid := clientPair.C1.ReadyTime.Sub(clientPair.ReadyStart) <= readyCheckTime
			client2ReadyValid := clientPair.C2.ReadyTime.Sub(clientPair.ReadyStart) <= readyCheckTime

			clientPair.C1.IsReadyChecking = false
			clientPair.C1.Ready = false
			clientPair.C2.IsReadyChecking = false
			clientPair.C2.Ready = false

			if client1ReadyValid && client2ReadyValid {
				matchid, err := database.CreateMatch(clientPair.C1.DBID, clientPair.C2.DBID)
				if err != nil {
					log.Printf("Failed to create a match: %s", err.Error())
				}

				clientPair.SendMatchConfirmedMessage(matchid)

				queue.Remove(clientPair.C1)
				queue.Remove(clientPair.C2)
				return
			}

			if !client1ReadyValid {
				queue.Remove(clientPair.C1)
			}

			if !client2ReadyValid {
				queue.Remove(clientPair.C2)
			}

			return
		}

		// Wait til next iteration if the time taken is less than the designated poll time
		elapsed := time.Now().Sub(start)
		remainingPollTime := pollTimeMatchFound - elapsed
		if remainingPollTime > 0 {
			time.Sleep(remainingPollTime)
		}
	}
}

func (queue *Queue) matchMake() (pairs []ClientPair) {
	// Matchmaking algo goes here. For now just return all the pairs we can put into pairs of two

	pairs = make([]ClientPair, 0)

	currentPair := ClientPair{}
	for _, index := range queue.index {
		client := queue.clients[index]
		if !client.IsReadyChecking {
			if currentPair.C1 == nil {
				currentPair.C1 = client
			} else {
				currentPair.C2 = client
				pairs = append(pairs, currentPair)
				currentPair = ClientPair{}
			}
		}
	}

	return pairs
}

func (queue *Queue) processCommand(command Command) {
	log.Printf("Processing command of type [ %v ] with data [ %v ]", command.Type, command.Data)
}
