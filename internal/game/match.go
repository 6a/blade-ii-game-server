package game

import (
	"bytes"
	"log"
	"strconv"
	"sync"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

const clientDataDelimiter string = "."
const payloadDelimiter string = ":"
const debugGameID uint64 = 20
const boltedCardOffset = 11

// Match is a wrapper for a matches data and client connections etc
type Match struct {
	ID        uint64
	Client1   *GClient
	Client2   *GClient
	State     MatchState
	Server    *Server
	phaseLock sync.Mutex
}

// Tick reads any incoming messages and passes outgoing messages to the queue
func (match *Match) Tick() {
	// Client 1
	match.tickClient(match.Client1, match.Client2, Player1)

	// Client 2
	match.tickClient(match.Client2, match.Client1, Player2)
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
	match.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchData, client1MessageString))

	client2MessageString := makeMessageString(InstructionCards, client2Buffer.String())
	match.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchData, client2MessageString))
}

// SendPlayerData sends each player's (their own) name to the respective client
func (match *Match) SendPlayerData() {
	var client1Buffer bytes.Buffer
	client1Buffer.WriteString(match.Client1.DisplayName)
	client1Buffer.WriteString(clientDataDelimiter)
	client1Buffer.WriteString(strconv.Itoa(int(match.Client1.Avatar)))

	var client2Buffer bytes.Buffer
	client2Buffer.WriteString(match.Client2.DisplayName)
	client2Buffer.WriteString(clientDataDelimiter)
	client2Buffer.WriteString(strconv.Itoa(int(match.Client2.Avatar)))

	match.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchData, makeMessageString(InstructionPlayerData, client1Buffer.String())))
	match.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchData, makeMessageString(InstructionPlayerData, client2Buffer.String())))
}

// SendOpponentData sends each player the opponents data
func (match *Match) SendOpponentData() {
	var client1Buffer bytes.Buffer
	client1Buffer.WriteString(match.Client2.DisplayName)
	client1Buffer.WriteString(clientDataDelimiter)
	client1Buffer.WriteString(match.Client2.PublicID)
	client1Buffer.WriteString(clientDataDelimiter)
	client1Buffer.WriteString(strconv.Itoa(int(match.Client2.Avatar)))

	match.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchData, makeMessageString(InstructionOpponentData, client1Buffer.String())))

	var client2Buffer bytes.Buffer
	client2Buffer.WriteString(match.Client1.DisplayName)
	client2Buffer.WriteString(clientDataDelimiter)
	client2Buffer.WriteString(match.Client1.PublicID)
	client2Buffer.WriteString(clientDataDelimiter)
	client2Buffer.WriteString(strconv.Itoa(int(match.Client1.Avatar)))

	match.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchData, makeMessageString(InstructionOpponentData, client2Buffer.String())))
}

// SetMatchStart sets the phase + start time for the current match
// Fails silently but logs errors
//
// Database update is performed in a goroutine
func (match *Match) SetMatchStart() {
	// Update the local match state
	match.SetPhase(Play)

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

		// Update the MMR's

	}()
}

// SetPhase sets the match phase
func (match *Match) SetPhase(phase Phase) {
	match.phaseLock.Lock()
	defer match.phaseLock.Unlock()
	match.State.Phase = phase

}

// GetPhase gets the match phase
func (match *Match) GetPhase() Phase {
	match.phaseLock.Lock()
	defer match.phaseLock.Unlock()
	return match.State.Phase
}

func (match *Match) isValidMove(move Move, player Player) bool {
	// Early exit if the player tried to make a move during the other players turn
	if match.State.Turn != player && match.State.Turn != PlayerUndecided {
		return false
	}

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
func (match *Match) tickClient(client *GClient, other *GClient, player Player) {
	for len(client.connection.ReceiveQueue) > 0 {
		message := client.connection.GetNextReceiveMessage()
		if message.Type == protocol.Type(protocol.WSMTText) {
			if message.Payload.Code == protocol.WSCMatchMove {
				move, err := MoveFromString(message.Payload.Message)
				if err == nil && match.isValidMove(move, player) {
					// Update game state
					match.updateGameState(player, move)

					// Forward to other client
					other.SendMessage(message)
				} else {
					// Remove the offending client (this will also end the game)
					match.Server.Remove(client, protocol.WSCMatchIllegalMove, "")
				}
			} else if message.Type == protocol.Type(protocol.WSCMatchForfeit) {
				// Remove the forfeiting client (this will also end the game)
				match.Server.Remove(client, protocol.WSCMatchForfeit, "")
			} else if message.Type == protocol.Type(protocol.WSCMatchRelayMessage) {
				other.SendMessage(message)
			}
		} else {
			// Handle non-text messages?
		}
	}
}

func (match *Match) updateGameState(player Player, move Move) {
	var targetHand []Card
	var targetField []Card
	var targetDeck []Card
	var targetDiscard []Card

	var oppositeHand []Card
	var oppositeField []Card
	var oppositeDeck []Card
	var oppositeDiscard []Card

	if player == Player1 {
		targetHand = match.State.Cards.Player1Hand
		targetField = match.State.Cards.Player1Field
		targetDeck = match.State.Cards.Player1Deck
		targetDiscard = match.State.Cards.Player1Discard

		oppositeHand = match.State.Cards.Player2Hand
		oppositeField = match.State.Cards.Player2Field
		oppositeDeck = match.State.Cards.Player2Deck
		oppositeDiscard = match.State.Cards.Player2Discard
	} else {
		targetHand = match.State.Cards.Player2Hand
		targetField = match.State.Cards.Player2Field
		targetDeck = match.State.Cards.Player2Deck
		targetDiscard = match.State.Cards.Player2Discard

		oppositeHand = match.State.Cards.Player1Hand
		oppositeField = match.State.Cards.Player1Field
		oppositeDeck = match.State.Cards.Player1Deck
		oppositeDiscard = match.State.Cards.Player1Discard
	}

	inCard := move.Instruction.ToCard()
	if !removeFirstOfType(targetHand, inCard) {
		// Early exit to bad move
		return
	}

	usedRodEffect := inCard == ElliotsOrbalStaff && len(targetField) > 0 && isBolted(targetField[len(targetField)-1])
	usedBoltEffect := inCard == Bolt && len(oppositeField) > 0 && !isBolted(oppositeField[len(oppositeField)-1])
	usedMirrorEffect := inCard == Mirror && len(targetField) > 0 && len(oppositeField) > 0
	usedBlastEffect := inCard == Blast && len(oppositeHand) > 0
	usedNormalOrForceCard := !usedRodEffect && !usedBoltEffect && !usedMirrorEffect && !usedBlastEffect

	// If the selected card was a normal or force card, and the target players latest field card is flipped (bolted) remove it
	if usedNormalOrForceCard && len(targetField) > 0 && isBolted(targetField[len(targetField)-1]) {
		removedCard := targetField[len(targetField)-1]

		if !removeFirstOfType(targetField, removedCard) {
			// Early exit to bad move
			return
		}

		targetDiscard = append(targetDiscard, removedCard)
	}

	if len(targetField) > 0 && !usedNormalOrForceCard {
		if usedBlastEffect {
			blastedCardInt, err := strconv.Atoi(move.Payload)
			if err != nil {
				// Early exit to bad move
				return
			}

			blastedCard := Card(blastedCardInt)

			if !removeFirstOfType(oppositeHand, blastedCard) {
				// Early exit to bad move
				return
			}

			oppositeDiscard = append(oppositeDiscard, blastedCard)

		} else if usedRodEffect {
			unBolt(targetField)
		} else if usedBoltEffect {
			bolt(oppositeField)
		} else if usedMirrorEffect {
			tempTargetField := targetField
			targetField = oppositeField
			oppositeField = tempTargetField
		}

		targetDiscard = append(targetDiscard, inCard)
	} else {
		targetField = append(targetField, inCard)
	}

	match.State.Player1Score = calculateScore(match.State.Cards.Player1Field)
	match.State.Player2Score = calculateScore(match.State.Cards.Player2Field)

	if match.State.Player1Score == match.State.Player2Score {
		match.State.Turn = PlayerUndecided
	} else if match.State.Player1Score < match.State.Player2Score {
		match.State.Turn = Player1
	} else {
		match.State.Turn = Player2
	}

	// Based on whos turn it is, start a timer
	// If the turn is undecided, wait for both
	// TODO implement undecided turn message sending to the server (only for network games?)
}

func isBolted(card Card) bool {
	return card > Force
}

func bolt(targetField []Card) {
	if len(targetField) > 0 {
		if targetField[len(targetField)-1] <= Force {
			targetField[len(targetField)-1] = Card(uint8(targetField[len(targetField)-1]) + boltedCardOffset)
		}
	}
}

func unBolt(targetField []Card) {
	if len(targetField) > 0 {
		if targetField[len(targetField)-1] >= InactiveElliotsOrbalStaff {
			targetField[len(targetField)-1] = Card(uint8(targetField[len(targetField)-1]) - boltedCardOffset)
		}
	}
}

func calculateScore(targetCards []Card) uint16 {
	var total uint16 = 0

	for i := 0; i < len(targetCards); i++ {
		card := targetCards[i]

		if !isBolted(card) {
			if card == Force && i > 0 {
				total *= 2
			} else {
				total += uint16(card.Value())
			}
		}
	}

	return total
}
