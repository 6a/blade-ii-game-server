package game

import (
	"bytes"
	"log"
	"strconv"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

const clientDataDelimiter string = "."
const payloadDelimiter string = ":"
const debugGameID uint64 = 20

// Match is a wrapper for a matche's data and client connections etc
type Match struct {
	ID      uint64
	Client1 *GClient
	Client2 *GClient
	State   MatchState
	Server  *Server
}

// Tick reads any incoming messages and passes outgoing messages to the queue
func (match *Match) Tick() {
	// Client 1
	match.tickClient(match.Client1, match.Client2)

	// Client 2
	match.tickClient(match.Client2, match.Client1)
}

// IsFull returns true when the match is occupied by two players
func (match *Match) IsFull() bool {
	return match.Client1 != nil && match.Client2 != nil
}

// BroadCast sends the specified message to both clients
func (match *Match) BroadCast(message protocol.Message) {
	match.Client1.SendMessage(message)
	match.Client2.SendMessage(message)
}

// SendMatchData sends match data (cards) to each client
func (match *Match) SendMatchData(cards string) {
	var client1Buffer bytes.Buffer
	var client2Buffer bytes.Buffer

	client1Buffer.WriteString("0")
	client1Buffer.WriteString(SerializedCardsDelimiter)
	client1Buffer.WriteString(cards)

	client2Buffer.WriteString("1")
	client2Buffer.WriteString(SerializedCardsDelimiter)
	client2Buffer.WriteString(cards)

	client1MessageString := makeMessageString(InstructionCards, client1Buffer.String())
	match.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchInstruction, client1MessageString))

	client2MessageString := makeMessageString(InstructionCards, client2Buffer.String())
	match.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchInstruction, client2MessageString))
}

// SendPlayerData sends each player's (their own) name to the respective client
func (match *Match) SendPlayerData() {
	match.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchInstruction, makeMessageString(InstructionPlayerData, match.Client1.DisplayName)))
	match.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchInstruction, makeMessageString(InstructionPlayerData, match.Client2.DisplayName)))
}

// SendOpponentData sends each player the opponents data
func (match *Match) SendOpponentData() {
	var client1Buffer bytes.Buffer
	client1Buffer.WriteString(match.Client2.DisplayName)
	client1Buffer.WriteString(clientDataDelimiter)
	client1Buffer.WriteString(match.Client2.PublicID)

	match.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchInstruction, makeMessageString(InstructionOpponentData, client1Buffer.String())))

	var client2Buffer bytes.Buffer
	client2Buffer.WriteString(match.Client1.DisplayName)
	client2Buffer.WriteString(clientDataDelimiter)
	client2Buffer.WriteString(match.Client1.PublicID)

	match.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchInstruction, makeMessageString(InstructionOpponentData, client2Buffer.String())))
}

// SetMatchStart sets the phase + start time for the current match
// Fails silently but logs errors
//
// Database update is performed in a goroutine
func (match *Match) SetMatchStart() {
	// Update the local match state
	match.State.Phase = Play

	// Early exit if we are currently in the debug match (dont write to the db)
	if match.ID == debugGameID {
		return
	}

	go func() {
		// Update the match phase in the database
		err := database.SetMatchStart(match.ID)
		if err != nil {
			// Output to log but otherwise continue
			log.Printf("Failed to update match phase: %s", err.Error())
		}
	}()
}

// SetMatchResult sends the match result to the database
// Fails silently but logs errors
//
// Performed in a goroutine
func (match *Match) SetMatchResult() {
	// Early exit if the winner is not yet decided
	if match.State.Winner == 0 {
		return
	}

	// Early exit if we are currently in the debug match (dont write to the db)
	if match.ID == debugGameID {
		return
	}

	go func() {
		// Update the match phase in the database
		err := database.SetMatchResult(match.ID, match.State.Winner)
		if err != nil {
			// Output to log but otherwise continue
			log.Printf("Failed to update match result: %s", err.Error())
		}
	}()
}

func isValidMove(payloadMessage string) bool {

	return true
}

func makeMessageString(instruction B2MatchInstruction, data string) string {
	var buffer bytes.Buffer

	buffer.WriteString(strconv.Itoa(int(instruction)))
	buffer.WriteString(payloadDelimiter)
	buffer.WriteString(data)

	return buffer.String()
}

// tickClient performs the tick actions for the specified client
func (match *Match) tickClient(client *GClient, other *GClient) {
	for len(client.connection.ReceiveQueue) > 0 {
		message := client.connection.GetNextReceiveMessage()
		if message.Type == protocol.Type(protocol.WSMTText) {
			if message.Payload.Code == protocol.WSCMatchInstruction {
				if isValidMove(message.Payload.Message) {
					// Update game state

					// Forward to other client
					other.SendMessage(message)
				} else {
					// Remove the offending client (this will also end the game)
					match.Server.Remove(client, protocol.WSCMatchIllegalMove, "")

					// Record the result in the database

				}
			} else if message.Type == protocol.Type(protocol.WSCMatchForfeit) {
				match.Server.Remove(client, protocol.WSCMatchForfeit, "")

				// Record the result in the database
			} else if message.Type == protocol.Type(protocol.WSCMatchRelayMessage) {
				other.SendMessage(message)
			}
		} else {
			// Handle non-text messages?
		}
	}
}
