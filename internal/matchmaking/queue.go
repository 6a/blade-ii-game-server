package matchmaking

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
const readyCheckTime = time.Second * 5

// How frequently to update the matchmaking queue (minimum wait between iterations)
const pollTime = 1000 * time.Millisecond

// Queue is a wrapper for the matchmaking queue
type Queue struct {
	index        []uint64
	nextIndex    uint64
	clients      map[uint64]*MMClient
	register     chan *MMClient
	unregister   chan UnregisterRequest
	broadcast    chan protocol.Message
	commands     chan Command
	matchedPairs []ClientPair
}

// Start the queue's internal update loop in a separate goroutine
func (queue *Queue) Start() {
	queue.index = make([]uint64, 0)
	queue.clients = make(map[uint64]*MMClient)
	queue.register = make(chan *MMClient, BufferSize)
	queue.unregister = make(chan UnregisterRequest, BufferSize)
	queue.broadcast = make(chan protocol.Message, BufferSize)
	queue.commands = make(chan Command, BufferSize)
	queue.matchedPairs = make([]ClientPair, 0)

	go queue.MainLoop()
}

// MainLoop is the main logic loop for the queue
func (queue *Queue) MainLoop() {
	for {
		start := time.Now()

		// Perform all pending tasks
		toRemove := make([]UnregisterRequest, 0)
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
		sort.Slice(toRemove, func(i, j int) bool { return toRemove[i].clientID < toRemove[j].clientID })

		// Store a value to use so that we can retain our position when iterating down the queue index
		indexIterator := len(queue.index) - 1

		// Remove any clients that are pending removal
		for index := len(toRemove) - 1; index >= 0; index-- {
			// Remove from the queue (map)
			removalIndex := toRemove[index].clientID

			if client, ok := queue.clients[removalIndex]; ok {
				deletedClientPID := client.PublicID
				go client.Close(protocol.NewMessage(protocol.WSMTText, toRemove[index].Reason, toRemove[index].Message))
				delete(queue.clients, removalIndex)

				// Remove from the queue index (slice)
				for indexIterator >= 0 {
					if queue.index[index] == removalIndex {
						if len(queue.index) == 1 {
							queue.index = make([]uint64, 0)
						} else {
							queue.index = append(queue.index[:indexIterator], queue.index[indexIterator+1:]...)
						}
						break
					}

					indexIterator--
				}

				log.Printf("Client [%s] left the matchmaking queue. Total clients: %v", deletedClientPID, len(queue.clients))
			}
		}

		// Tick all clients
		for _, client := range queue.clients {
			client.Tick()
		}

		// Perform matchmaking
		queue.matchedPairs = append(queue.matchedPairs, queue.matchMake()...)
		for index := len(queue.matchedPairs) - 1; index >= 0; index-- {
			if !queue.matchedPairs[index].IsReadyChecking {
				queue.matchedPairs[index].SendMatchStartMessage()
			}

			if queue.pollReadyCheck(queue.matchedPairs[index]) {
				if len(queue.matchedPairs) == 1 {
					queue.matchedPairs = make([]ClientPair, 0)
				} else {
					queue.matchedPairs = queue.matchedPairs[:len(queue.matchedPairs)-1]
				}
			}
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
func (queue *Queue) Add(client *MMClient) {
	queue.register <- client
}

// Remove adds a client to the unregister queue, to be removed next cycle
func (queue *Queue) Remove(client *MMClient, reason protocol.B2Code, message string) {
	queue.unregister <- UnregisterRequest{
		clientID: client.QueueID,
		Reason:   reason,
		Message:  message,
	}
}

// Broadcast sends a message to all connected clients
func (queue *Queue) Broadcast(message protocol.Message) {
	queue.broadcast <- message
}

func (queue *Queue) pollReadyCheck(clientPair ClientPair) (finished bool) {
	timedOut := time.Now().Sub(clientPair.ReadyStart) > readyCheckTime
	client1ReadyValid := clientPair.C1.Ready && clientPair.C1.ReadyTime.Sub(clientPair.ReadyStart) <= readyCheckTime
	client2ReadyValid := clientPair.C2.Ready && clientPair.C2.ReadyTime.Sub(clientPair.ReadyStart) <= readyCheckTime

	if (client1ReadyValid && client2ReadyValid) || timedOut {
		if timedOut || !client1ReadyValid || !client2ReadyValid {
			if !client1ReadyValid {
				queue.Remove(clientPair.C1, protocol.WSCReadyCheckFailed, "")
			} else {
				clientPair.C1.IsReadyChecking = false
				clientPair.C1.Ready = false
			}

			if !client2ReadyValid {
				queue.Remove(clientPair.C2, protocol.WSCReadyCheckFailed, "")
			} else {
				clientPair.C2.IsReadyChecking = false
				clientPair.C2.Ready = false
			}

			return true
		}

		matchid, err := database.CreateMatch(clientPair.C1.DBID, clientPair.C2.DBID)
		if err != nil {
			log.Printf("Failed to create a match: %s", err.Error())
		}

		clientPair.SendMatchConfirmedMessage(matchid)

		queue.Remove(clientPair.C1, protocol.WSCInfo, "Match found - closing connection")
		queue.Remove(clientPair.C2, protocol.WSCInfo, "Match found - closing connection")
		return true
	}

	return false
}

func (queue *Queue) matchMake() (pairs []ClientPair) {
	// Matchmaking algo goes here. For now just return all the pairs we can put into pairs of two
	pairs = make([]ClientPair, 0)

	currentPair := ClientPair{}
	for _, index := range queue.index {
		if client, ok := queue.clients[index]; ok {
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
	}

	return pairs
}

func (queue *Queue) processCommand(command Command) {
	log.Printf("Processing command of type [ %v ] with data [ %v ]", command.Type, command.Data)
}
