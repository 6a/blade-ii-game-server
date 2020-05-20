// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game implements the Blade II Online game server.
package game

import (
	"log"
	"time"

	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

const (

	// BufferSize is the size of each message queue's buffer.
	BufferSize = 2048

	// How frequently to update the matchmaking queue (minimum wait between iterations).
	pollTime = 250 * time.Millisecond
)

// Server is the game server itself
type Server struct {

	// A map containing all the matches, keyed by match ID.
	matches map[uint64]*Match

	// Channel for new client's that have been successfully authenticated, and have
	// been confirmed to eligible for a match.
	connect chan *GClient

	// Channel for client's that are to be disconnected, but not necessarily on the same tick.
	disconnect chan DisconnectRequest

	// Channel for client's that are to be disconnected, but specifically on this tick.
	immediateDisconnect chan DisconnectRequest

	// Channel for messages that should be broadcasted to all clients.
	broadcast chan protocol.Message

	// Channel for server commands.
	commands chan protocol.Command
}

// Init initializes the game server including starting the internal loop.
func (gs *Server) Init() {

	// Initialize the matches map.
	gs.matches = make(map[uint64]*Match)

	// Initialize the various channels.
	gs.connect = make(chan *GClient, BufferSize)
	gs.disconnect = make(chan DisconnectRequest, BufferSize)
	gs.immediateDisconnect = make(chan DisconnectRequest, BufferSize)
	gs.broadcast = make(chan protocol.Message, BufferSize)
	gs.commands = make(chan protocol.Command, BufferSize)

	go gs.MainLoop()
}

// NewServer creates and returns a pointer to a new game server.
func NewServer() *Server {

	// Create a new game server.
	gs := Server{}

	// Initialize the game server.
	gs.Init()

	// Return a pointer to the newly created game server.
	return &gs
}

// AddClient takes a websocket connection various data, wraps them up and adds them to the game server as a client, to be processed later.
func (gs *Server) AddClient(wsconn *websocket.Conn, dbid uint64, pid string, displayname string, avatar uint8, matchID uint64) {

	// Create a new client
	client := NewClient(wsconn, dbid, pid, displayname, matchID, avatar, gs)

	// Add it to the connect queue.
	gs.connect <- client
}

// Remove adds a client to the disconnect queue, to be disconnected later, along with a reason code and a message.
func (gs *Server) Remove(client *GClient, reason protocol.B2Code, message string) {

	// Create a new disconnect request
	disconnectRequest := DisconnectRequest{
		Client:  client,
		Reason:  reason,
		Message: message,
	}

	// Add it to the disconnect queue
	gs.disconnect <- disconnectRequest
}

// MainLoop is the main logic loop for the game server.
func (gs *Server) MainLoop() {

	// Loop forever.
	for {

		// Log the start time for this server tick - so that we can introduce a wait if the tick takes less time than
		// the minimum wait, to reduce server load.
		start := time.Now()

		// If any of the queues have something in them, process their data until all the queues are empty.
		for len(gs.connect)+len(gs.disconnect)+len(gs.broadcast)+len(gs.commands) > 0 {

			// Using a select, read from either the connect, broadcast, or command queue - whichever comes first.
			select {
			case client := <-gs.connect:

				// If the match ID specified by the incoming client already exists, it should be ok to join in some fashion.
				// Otherwise, the match needs to be created.
				if match, ok := gs.matches[client.MatchID]; ok {

					// If the game is already in play, the player cannot be added, and are booted out. Otherwise, add them to the game.
					if match.GetPhase() >= Play {
						gs.Remove(client, protocol.WSCMatchFull, "Attempted to join a match which already has both clients registered")
					} else {

						// Depending on the state of the match, add the client to it as either player 1 or player 2.

						if match.Client1 == nil {

							// Dependingo on client 1 and client 2...
							if match.Client2 == nil {

								// If client 1 and client 2 are both nil, add the client in as player 1.
								match.Client1 = client

								// Send a message to the client informing them that they joined a match.
								client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))

								log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
							} else if client.DBID == match.Client2.DBID {

								// If client 2's database ID is the same as the incoming client's database ID, they are the same client, and
								// the old one needs to be replaced.

								// Remove the old connection.
								gs.Remove(match.Client2, protocol.WSCMatchMultipleConnections, "Removing old connection from same client")

								// Set the incoming client as client 2.
								match.Client2 = client

								// Send a message to the client informing them that they joined a match.
								client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))

								log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
							} else {

								// If we reach here, client 2 is either nil or has a different database ID to the incoming client, so
								// the incoming client becomes player 1.
								match.Client1 = client

								// Send a message to the client informing them that they joined a match.
								client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))

								log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
							}
						} else if match.Client1.DBID == client.DBID {

							// If client 1's database ID matches the incoming client's database ID, they are ther same user, and
							// therefore the old connection must be replaced.

							// Remove the old connection.
							gs.Remove(gs.matches[client.MatchID].Client1, protocol.WSCMatchMultipleConnections, "Removing old connection from same client")

							// The incoming client becomes player 1.
							match.Client1 = client

							// Send a message to the client informing them that they joined a match.
							client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))

							log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
						} else {

							// Finally, if we reach here, it means player 1 is valid (and is another user), and therefore we assign the
							// incoming client as player 2.
							match.Client2 = client

							// Send a message to the client informing them that they joined a match.
							client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))

							log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
						}

						// At this stage, if both clients are now present, the match is ready to start.
						if match.Client1 != nil && match.Client2 != nil {

							// Generate the cards for this game.
							cardsToSend := GenerateCards()

							// Generate the initialized cards, to be set as the initial card state for the match.
							initializedCards := InitializeCards(cardsToSend)

							// Set the initial card state for the match.
							match.State.Cards = initializedCards

							// Set the match phase to start.
							match.SetMatchStart()

							// Send all the match data to each player.
							match.SendCardData(cardsToSend.Serialized())
							match.SendPlayerData()
							match.SendOpponentData()

							log.Printf("Match [%v] started. Total matches: %v", client.MatchID, len(gs.matches))
						}
					}
				} else {

					// Create a new match with the client that just joined, and add it to the match map.
					gs.matches[client.MatchID] = NewMatch(client.MatchID, client, gs)

					// Send a message to the client informing them that they joined a match.
					client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchJoined, "Joined match"))

					log.Printf("Client [%s] joined match [%v]. Total matches: %v", client.PublicID, client.MatchID, len(gs.matches))
				}

				break
			case message := <-gs.broadcast:

				// Broadcasted messages are simply broadcasted to all matches in the match map.
				for _, match := range gs.matches {
					match.BroadCast(message)
				}

				break
			case command := <-gs.commands:

				// Process the command.
				gs.processCommand(command)

				break
			case disconnectRequest := <-gs.disconnect:

				// Add the disconnect request to the immediate disconnect request queue. This acts as a sort of sync barrier
				// to prevent disconnect requests from being added between the tick and the disconnect handler.
				gs.immediateDisconnect <- disconnectRequest
				break
			}
		}

		// Tick all matches
		for _, match := range gs.matches {

			// only tick a match if it is current in a play state.
			if match.GetPhase() == Play {
				match.Tick()
			}
		}

		// Handle any pending disconnect requests.
		gs.handleDisconnectRequests()

		// Add a delay before the next iteration if the time taken is less than the designated poll time.
		elapsed := time.Now().Sub(start)
		remainingPollTime := pollTime - elapsed
		if remainingPollTime > 0 {
			time.Sleep(remainingPollTime)
		}
	}
}

// handleDisconnectRequests handles disconnect requests for clients in the server.
func (gs *Server) handleDisconnectRequests() {

	// Loop while there are disconnect requests in the disconnect queue.
	for len(gs.immediateDisconnect) > 0 {
		select {
		case req := <-gs.immediateDisconnect:

			// If the match exists, determine if we need to remove just the client, end the match etc.. Otherwise,
			// just remove the client.
			if match, ok := gs.matches[req.Client.MatchID]; ok {

				// Early exit if the reason was an error but the match has already ended gracefully, as then we dont need to
				// handle the error. Logic is backwards (checks for graceful finish + non win/draw code)
				if match.isMatchGracefullyFinished() && req.Reason != protocol.WSCMatchWin && req.Reason != protocol.WSCMatchDraw {
					break
				}

				// Set up some variables that will allow us to use the same logic regardless of whether the
				// client that requested the disconnect was client 1 or 2.
				initiator := req.Client
				var initiatorReason protocol.B2Code
				var initiatorMessage string

				var other *GClient
				var otherReason protocol.B2Code
				var otherMessage string

				// Determine which of the clients is the other client; the one that did not initiase the disconnect.
				if match.Client1.DBID == req.Client.DBID {
					other = match.Client2
				} else {
					other = match.Client1
				}

				// Act accordingly, depending on the disconnect request reason.
				// Gracefully ended matches are exempt from error checks, as they clients are free
				// to do what they want as no more interactions are required from them, and they can
				// disconnect without issue.

				if req.Reason == protocol.WSCUnknownConnectionError {

					// Unknown errors are websocket errors - such as a broken connection.
					// Set the reason and message payloads accordingly.
					initiatorReason = protocol.WSCMatchForfeit
					initiatorMessage = "Post-forfeit quit"

					otherReason = protocol.WSCMatchForfeit
					otherMessage = "Opponent forfeited the match"

					// For disconnections, we need to determine the winner, as the disconnect was triggered by
					// the websocket, not the match or any other server game server logic. In this instance, the
					// player that disconnected loses, and therefore the winner is the other player.
					if match.GetPhase() > WaitingForPlayers {

						// Set the winner to the other player.
						match.State.Winner = other.DBID

						// Update the match in the database.
						match.SetMatchResult()
					}
				} else if req.Reason == protocol.WSCMatchForfeit {

					// Forfeit means that one of the players forfeited.
					// Set the reason and message payloads accordingly.
					initiatorReason = protocol.WSCMatchForfeit
					initiatorMessage = "Post-forfeit quit"

					otherReason = protocol.WSCMatchForfeit
					otherMessage = "Opponent forfeited the match"

					// Update the match in the database.
					match.SetMatchResult()
				} else if req.Reason == protocol.WSCMatchIllegalMove {

					// Illegal move means that a player's move was invalid, out of order etc..
					// Set the reason and message payloads accordingly.
					initiatorReason = protocol.WSCMatchIllegalMove
					initiatorMessage = "Post-illegal move forfeit quit"

					otherReason = protocol.WSCMatchForfeit
					otherMessage = "Opponent forfeited the match"

					// Update the match in the database.
					match.SetMatchResult()
				} else if req.Reason == protocol.WSCMatchTimeOut {

					// Timeout means that one of the players timed out (did not play a move
					// within the turn time limit).
					// Set the reason and message payloads accordingly.
					initiatorReason = protocol.WSCMatchTimeOut
					initiatorMessage = "Timed out"

					otherReason = protocol.WSCMatchForfeit
					otherMessage = "Opponent timed out"

					// Update the match in the database.
					match.SetMatchResult()
				} else if req.Reason == protocol.WSCMatchWin {

					// A win means that the initiator won the match.
					// Set the reason and message payloads accordingly.
					initiatorReason = protocol.WSCMatchWin
					initiatorMessage = "Victory"

					otherReason = protocol.WSCMatchLoss
					otherMessage = "Defeat"

					// Update the match in the database.
					match.SetMatchResult()
				} else if req.Reason == protocol.WSCMatchLoss {

					// Note that this should never be reached - to declare a loss, simply declare the winner instead.
					log.Panicf("Don't set the reason to loss - rather, set win for the winning client instead")
				} else {

					// Any other reasons fall through to here. Unknown errors, or
					// reasons where the reason and message are the same for both players,
					// are possible reasons why execution reaches this point.
					initiatorReason = req.Reason
					initiatorMessage = req.Message

					otherReason = req.Reason
					otherMessage = req.Message

					// Update the match in the database.
					match.SetMatchResult()
				}

				// Once we reach this point, the match results have been written to the database, and the initiator
				// can be successfully disconnected.
				initiator.Close(protocol.NewMessage(protocol.WSMTText, initiatorReason, initiatorMessage))

				// Now, if the game was started...
				if match.GetPhase() > WaitingForPlayers {

					// Set the game to finished (may already be finished, but should be fine to call again).
					match.SetPhase(Finished)

					// If the client in the incoming disconnect request is one of the clients in the match, that means
					// that the match should be ended. Disconnect the other player (the initiator is already disconnected)
					// and remove the match from the match map. This check is in place, incase the disconnect request was
					// from an old connection for a client in the game - in this case, the connection in the request is
					// considered to be stale, and the other client, and the match, is left is intact.
					if (req.Client.IsSameConnection(match.Client1)) || req.Client.IsSameConnection(match.Client2) {

						// Close the other clients connection.
						other.Close(protocol.NewMessage(protocol.WSMTText, otherReason, otherMessage))

						// Remove the map from the match map.
						delete(gs.matches, match.ID)

						log.Printf("Client's [%s][%s] left the game server - match [%d] ended", match.Client1.PublicID, match.Client2.PublicID, match.ID)
					} else {

						// Noop, as the disconnection request came from a connection that was already replaced.
						log.Printf("Client [%s] left the game server - stale connection - match [%d] still active", initiator.PublicID, match.ID)
					}
				} else {

					// If the game is not yet started, determine which of the clients requested the disconnected, and then just nil the
					// pointer to them in the match - Also checking to see if it's the same connection, and not a stale one from the
					// same client. No need to remove them or anything, as the connection was already closed earlier.
					if req.Client.IsSameConnection(match.Client1) {
						match.Client1 = nil
					} else if req.Client.IsSameConnection(match.Client2) {
						match.Client2 = nil
					}

					log.Printf("Client [%s] left the game server - match [%d] still waiting for clients", initiator.PublicID, match.ID)
				}
			} else {

				// If the match specified by the client does not exist, then for whatever reason the match does not yet
				// exist - in this case, just kill the connection.
				req.Client.Close(protocol.NewMessage(protocol.WSMTText, req.Reason, req.Message))

				log.Printf("Client [%s] left the game server (was not in match)", req.Client.PublicID)
			}
		}
	}
}

// processCommand handles server commands.
//
// Note - not yet implemented, but prints out some diagonstics and returns with a noop.
func (gs *Server) processCommand(command protocol.Command) {
	log.Printf("Processing command of type [ %v ] with data [ %v ]", command.Type, command.Data)
}
