package game

import (
	"log"
	"time"

	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

// BufferSize is the size of each message queue's buffer
const BufferSize = 256

// readyCheckTime is how long to wait when a ready check is sent
const readyCheckTime = time.Second * 5

// How frequently to update the matchmaking queue (minimum wait between iterations)
const pollTime = 250 * time.Millisecond

// Server is the game server itself
type Server struct {
	matches    map[uint64]*Match
	register   chan *GClient
	unregister chan UnregisterRequest
	broadcast  chan protocol.Message
	commands   chan protocol.Command
}

// AddClient takes a new client and their various data, wraps them up and adds them to the game server
func (gs *Server) AddClient(wsconn *websocket.Conn, dbid uint64, pid string, displayname string, avatar uint8, matchID uint64) {
	client := NewClient(wsconn, dbid, pid, displayname, matchID, avatar, gs)
	client.StartEventLoop()

	gs.register <- client
}

// Remove adds a client to the unregister queue, to be removed next cycle
func (gs *Server) Remove(client *GClient, reason protocol.B2Code, message string) {
	gs.unregister <- UnregisterRequest{
		Client:  client,
		Reason:  reason,
		Message: message,
	}
}

// MainLoop is the main logic loop for the game server
func (gs *Server) MainLoop() {
	for {
		start := time.Now()

		// Perform all pending tasks
		for len(gs.register)+len(gs.broadcast)+len(gs.commands) > 0 {
			select {
			case client := <-gs.register:
				if match, ok := gs.matches[client.MatchID]; ok && match.Client1 != nil {
					if match.GetPhase() >= Play {
						gs.Remove(client, protocol.WSCMatchFull, "Attempted to join a match which already has both clients registered")
					} else {
						if match.Client1.DBID == client.DBID {
							gs.Remove(gs.matches[client.MatchID].Client1, protocol.WSCMatchMultipleConnections, "Removing old connection from same client")

							gs.matches[client.MatchID].Client1 = client

							client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))
							log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
						} else {
							match.Client2 = client

							client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))

							cardsToSend, drawsUntilValid := GenerateCards()
							initialisedCards := InitialiseCards(cardsToSend, drawsUntilValid)

							match.State.Cards = initialisedCards

							// Update the match phase
							match.SetMatchStart()

							if match.State.Cards.Player1Hand[0].Value() < match.State.Cards.Player1Hand[0].Value() {
								match.State.Turn = Player1
							} else {
								match.State.Turn = Player2
							}

							match.SendMatchData(cardsToSend.Serialized())
							match.SendPlayerData()
							match.SendOpponentData()
							log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
						}
					}
				} else {
					gs.matches[client.MatchID] = &Match{
						ID:      client.MatchID,
						Client1: client,
					}

					client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))
					log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
				}

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

		gs.handleUnregisterQueue()

		// Tick all matches
		for _, match := range gs.matches {
			if match.GetPhase() > WaitingForPlayers {
				match.Tick()
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

func (gs *Server) processCommand(command protocol.Command) {
	log.Printf("Processing command of type [ %v ] with data [ %v ]", command.Type, command.Data)
}

// handleUnregisterQueue handles unregister requests as they come down the unregister queue
func (gs *Server) handleUnregisterQueue() {
	for len(gs.unregister) > 0 {
		select {
		case req := <-gs.unregister:
			if match, ok := gs.matches[req.Client.MatchID]; ok {
				initiator := req.Client
				var initiatorReason protocol.B2Code
				var initiatorMessage string

				var other *GClient
				var otherReason protocol.B2Code
				var otherMessage string

				if match.Client1.DBID == req.Client.DBID {
					other = match.Client2
				} else {
					other = match.Client1
				}

				if req.Reason == protocol.WSCMatchForfeit || req.Reason == protocol.WSCUnknownConnectionError {
					initiatorReason = protocol.WSCMatchForfeit
					initiatorMessage = "Post-forfeit quit"

					otherReason = protocol.WSCMatchForfeit
					otherMessage = "Opponent forfeited the match"

					if match.GetPhase() > WaitingForPlayers {
						match.State.Winner = other.DBID
						match.SetMatchResult()
					}
				} else if req.Reason == protocol.WSCMatchIllegalMove {
					initiatorReason = protocol.WSCMatchIllegalMove
					initiatorMessage = "Post-illegal move forfeit quit"

					otherReason = protocol.WSCMatchForfeit
					otherMessage = "Opponent forfeited the match"

					if match.GetPhase() > WaitingForPlayers {
						match.State.Winner = other.DBID
						match.SetMatchResult()
					}
				} else {
					initiatorReason = req.Reason
					initiatorMessage = req.Message

					otherReason = req.Reason
					otherMessage = req.Message
				}

				initiator.Close(protocol.NewMessage(protocol.WSMTText, initiatorReason, initiatorMessage))

				if match.GetPhase() > WaitingForPlayers {
					match.SetPhase(Finished)

					if (req.Client.IsSameConnection(match.Client1)) || req.Client.IsSameConnection(match.Client2) {
						other.Close(protocol.NewMessage(protocol.WSMTText, otherReason, otherMessage))

						delete(gs.matches, match.ID)
						log.Printf("Client's [%s][%s] left the game server - match [%d] ended", match.Client1.PublicID, match.Client2.PublicID, match.ID)
					} else {
						log.Printf("Client [%s] left the game server - stale connection - match [%d] still active", initiator.PublicID, match.ID)
					}
				} else {
					if req.Client.IsSameConnection(match.Client1) {
						match.Client1 = nil
					} else if req.Client.IsSameConnection(match.Client2) {
						match.Client2 = nil
					}

					log.Printf("Client [%s] left the game server - match [%d] still waiting for clients", initiator.PublicID, match.ID)
				}
			} else {
				req.Client.Close(protocol.NewMessage(protocol.WSMTText, req.Reason, req.Message))
				log.Printf("Client [%s] left the game server (was not in game)", req.Client.PublicID)
			}
		}
	}
}

// Init initializes the game server including starting the internal loop
func (gs *Server) Init() {
	gs.matches = make(map[uint64]*Match)
	gs.register = make(chan *GClient, BufferSize)
	gs.unregister = make(chan UnregisterRequest, BufferSize)
	gs.broadcast = make(chan protocol.Message, BufferSize)
	gs.commands = make(chan protocol.Command, BufferSize)

	go gs.MainLoop()
}

// NewServer creates a new game server
func NewServer() *Server {
	gs := Server{}
	return &gs
}
