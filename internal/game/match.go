package game

import (
	"bytes"
	"strconv"

	"github.com/6a/blade-ii-game-server/internal/protocol"
)

const clientDataDelimiter string = "."
const payloadDelimiter string = ":"

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
	for len(match.Client1.connection.ReceiveQueue) > 0 {
		message := match.Client1.connection.GetNextReceiveMessage()
		if message.Type == protocol.Type(protocol.WSMTText) {
			if message.Payload.Code == protocol.WSCMatchInstruction {
				if isValidMove(message.Payload.Message) {
					match.Client2.SendMessage(protocol.NewMessageFromPayload(protocol.WSMTText, message.Payload))
				} else {
					match.Server.Remove(match.Client1, protocol.WSCMatchIllegalMove, "")

					// Record the result in the database
				}
			} else if message.Type == protocol.Type(protocol.WSCMatchForfeit) {
				match.Server.Remove(match.Client1, protocol.WSCMatchForfeit, "")

				// Record the result in the database
			} else if message.Type == protocol.Type(protocol.WSCMatchRelayMessage) {
				// Relay text message
			}
		} else {
			// Handle non-text messages?
		}
	}
}

// TickClient performs the tick actions for the specified client
func (match *Match) TickClient(client *GClient) {

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
