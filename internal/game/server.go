package game

import (
	"log"
	"time"

	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

// BufferSize is the size of each message queue's buffer
const BufferSize = 32

// readyCheckTime is how long to wait when a ready check is sent
const readyCheckTime = time.Second * 5

// How frequently to update the matchmaking queue (minimum wait between iterations)
const pollTime = 1000 * time.Millisecond

// Server is the game server itself
type Server struct {
	matches    map[uint64]*Match
	register   chan *GClient
	unregister chan UnregisterRequest
	broadcast  chan protocol.Message
	commands   chan protocol.Command
}

// AddClient takes a new client and their various data, wraps them up and adds them to the game server
func (gs *Server) AddClient(wsconn *websocket.Conn, dbid uint64, pid string, matchID uint64) {
	client := NewClient(wsconn, dbid, pid, matchID, gs)
	client.StartEventLoop()

	gs.register <- &client
}

// Remove adds a client to the unregister queue, to be removed next cycle
func (gs *Server) Remove(client *GClient, reason protocol.B2Code, message string) {
	gs.unregister <- UnregisterRequest{
		matchID:  client.MatchID,
		clientID: client.DBID,
		Reason:   reason,
		Message:  message,
	}
}

// MainLoop is the main logic loop for the game server
func (gs *Server) MainLoop() {
	for {
		start := time.Now()

		// Perform all pending tasks
		toRemove := make([]UnregisterRequest, 0)
		for len(gs.register)+len(gs.unregister)+len(gs.broadcast) > 0 {
			select {
			case client := <-gs.register:
				if match, ok := gs.matches[client.MatchID]; ok {
					match.Client2 = client

					client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))

					cardsToSend, drawsUntilValid := Generate()
					initialisedCards := Initialise(cardsToSend, drawsUntilValid)

					match.State.Cards = initialisedCards

					if match.State.Cards.Player1Hand[0].Value() < match.State.Cards.Player1Hand[0].Value() {
						match.State.Turn = Player1
					} else {
						match.State.Turn = Player2
					}

					match.BroadCast(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchData, cardsToSend.Serialized()))

				} else {
					gs.matches[client.MatchID] = &Match{
						MatchID: client.MatchID,
						Client1: client,
					}

					client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))
				}

				log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
				break
			case clientid := <-gs.unregister:
				toRemove = append(toRemove, clientid)
				break
			case message := <-gs.broadcast:
				for _, match := range gs.matches {
					match.BroadCast(message)
				}
				break
			case command := <-gs.commands:
				gs.processCommand(command)
			}
		}

		// // Remove any clients that are pending removal
		// for index := len(toRemove) - 1; index >= 0; index-- {
		// 	// Remove from the queue (map)
		// 	removalIndex := toRemove[index].clientID

		// 	if client, ok := queue.clients[removalIndex]; ok {
		// 		deletedClientPID := client.PublicID
		// 		go client.Close(protocol.NewMessage(protocol.WSMTText, toRemove[index].Reason, toRemove[index].Message))
		// 		delete(queue.clients, removalIndex)

		// 		// Remove from the queue index (slice)
		// 		for indexIterator >= 0 {
		// 			if queue.index[index] == removalIndex {
		// 				if len(queue.index) == 1 {
		// 					queue.index = make([]uint64, 0)
		// 				} else {
		// 					queue.index = append(queue.index[:indexIterator], queue.index[indexIterator+1:]...)
		// 				}
		// 				break
		// 			}

		// 			indexIterator--
		// 		}

		// 		log.Printf("Client [%s] left the matchmaking queue. Total clients: %v", deletedClientPID, len(queue.clients))
		// 	}
		// }

		// Tick all matches
		for _, match := range gs.matches {
			match.Tick()
		}

		// Perform matchmaking
		// queue.matchedPairs = append(queue.matchedPairs, queue.matchMake()...)
		// for index := len(queue.matchedPairs) - 1; index >= 0; index-- {
		// 	if !queue.matchedPairs[index].IsReadyChecking {
		// 		queue.matchedPairs[index].SendMatchStartMessage()
		// 	}

		// 	if queue.pollReadyCheck(queue.matchedPairs[index]) {
		// 		if len(queue.matchedPairs) == 1 {
		// 			queue.matchedPairs = make([]ClientPair, 0)
		// 		} else {
		// 			queue.matchedPairs = queue.matchedPairs[:len(queue.matchedPairs)-1]
		// 		}
		// 	}
		// }

		// Wait til next iteration if the time taken is less than the designated poll time
		elapsed := time.Now().Sub(start)
		remainingPollTime := pollTime - elapsed
		if remainingPollTime > 0 {
			time.Sleep(remainingPollTime)
		}
	}
}

func (gs *Server) processCommand(command protocol.Command) {
	log.Printf("Processing command of type [ %v ] with data [ %v ]", command.Type, command.Data)
}

// Init initializes the game server including starting the internal loop
func (gs *Server) Init() {
	gs.matches = make(map[uint64]*Match)
	gs.register = make(chan *GClient, BufferSize)
	gs.unregister = make(chan UnregisterRequest, BufferSize)
	// gs.broadcast = make(chan protocol.Message, BufferSize)
	// gs.commands = make(chan Command, BufferSize)

	go gs.MainLoop()
}

// NewServer creates a new game server
func NewServer() Server {
	gs := Server{}
	return gs
}
