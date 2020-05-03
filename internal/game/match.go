package game

import (
	"bytes"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/6a/blade-ii-game-server/internal/apiinterface"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

const clientDataDelimiter string = "."
const payloadDelimiter string = ":"
const debugGameID uint64 = 20
const boltedCardOffset = 11
const turnMaxWait = time.Millisecond * 21000
const tiedScoreAdditionalWait = time.Millisecond * 4500 // additional time to allow for clearing the field etc client side
const blastCardAdditionalWait = time.Millisecond * 4500 // additional time to allow for the long-ass blast animation

// Match is a wrapper for a matches data and client connections etc
type Match struct {
	ID        uint64
	Client1   *GClient
	Client2   *GClient
	State     MatchState
	Server    *Server
	phaseLock sync.Mutex
	turnTimer *time.Timer
}

// Tick reads any incoming messages and passes outgoing messages to the queue
func (match *Match) Tick() {
	// Client 1
	match.tickClient(match.Client1, match.Client2, Player1)

	// Client 2
	match.tickClient(match.Client2, match.Client1, Player2)

	// Check for timeouts

	if len(match.turnTimer.C) > 0 {
		select {
		case _ = <-match.turnTimer.C:
			if match.Client1.WaitingForMove && match.Client2.WaitingForMove {
				match.Server.Remove(match.Client1, protocol.WSCMatchMutualTimeout, "Both players timed out")
			} else if match.Client1.WaitingForMove {
				match.State.Winner = match.Client2.DBID
				match.Server.Remove(match.Client1, protocol.WSCMatchTimeOut, "Player 1 timed out")
			} else {
				match.State.Winner = match.Client1.DBID
				match.Server.Remove(match.Client2, protocol.WSCMatchTimeOut, "Player 2 timed out")
			}

			match.SetPhase(Finished)
		}
	}
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

	// Start turn timers
	var nextTurnPeriod time.Duration
	if match.Client1.connection.Latency > match.Client2.connection.Latency {
		nextTurnPeriod = turnMaxWait + match.Client1.connection.Latency
	} else {
		nextTurnPeriod = turnMaxWait + match.Client2.connection.Latency
	}

	match.turnTimer = time.NewTimer(nextTurnPeriod)
	match.Client1.WaitingForMove = true
	match.Client2.WaitingForMove = true

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

		// Update match stats
		var winner apiinterface.Winner
		if match.State.Winner == match.Client1.DBID {
			winner = apiinterface.Player1
		} else if match.State.Winner == match.Client2.DBID {
			winner = apiinterface.Player2
		} else {
			winner = apiinterface.Draw
		}

		apiinterface.UpdateMatchStats(match.Client1.DBID, match.Client2.DBID, winner)
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
				client.WaitingForMove = false
				move, err := MoveFromString(message.Payload.Message)
				if err == nil && match.isValidMove(move, player) {
					// Update game state

					valid, matchEnded, winner := match.updateGameState(player, move)
					if valid {
						// Forward to other client
						other.SendMessage(message)

						if matchEnded {
							if winner == Player1 {
								match.State.Winner = match.Client1.DBID
								match.Server.Remove(match.Client1, protocol.WSCMatchWin, "")
							} else if winner == Player2 {
								match.State.Winner = match.Client2.DBID
								match.Server.Remove(match.Client2, protocol.WSCMatchWin, "")
							} else {
								match.Server.Remove(match.Client1, protocol.WSCMatchDraw, "")
							}

							match.SetPhase(Finished)
						}
					} else {
						// Remove the offending client (this will also end the game)
						match.State.Winner = other.DBID
						match.Server.Remove(client, protocol.WSCMatchIllegalMove, "")
					}
				} else {
					// Remove the offending client (this will also end the game)
					match.State.Winner = other.DBID
					match.Server.Remove(client, protocol.WSCMatchIllegalMove, "")
				}
			} else if message.Type == protocol.Type(protocol.WSCMatchForfeit) {
				// Remove the forfeiting client (this will also end the game)
				match.State.Winner = other.DBID
				match.Server.Remove(client, protocol.WSCMatchForfeit, "")
			} else if message.Type == protocol.Type(protocol.WSCMatchRelayMessage) {
				other.SendMessage(message)
			}
		} else {
			// Handle non-text messages?
		}
	}
}

func (match *Match) updateGameState(player Player, move Move) (validMove bool, matchEnded bool, winner Player) {
	var targetHand *[]Card
	var targetField *[]Card
	var targetDeck *[]Card
	var targetDiscard *[]Card

	var oppositeHand *[]Card
	var oppositeField *[]Card
	var oppositeDiscard *[]Card

	var updateTurn = false

	if player == Player1 {
		targetHand = &match.State.Cards.Player1Hand
		targetField = &match.State.Cards.Player1Field
		targetDeck = &match.State.Cards.Player1Deck
		targetDiscard = &match.State.Cards.Player1Discard

		oppositeHand = &match.State.Cards.Player2Hand
		oppositeField = &match.State.Cards.Player2Field
		oppositeDiscard = &match.State.Cards.Player2Discard
	} else {
		targetHand = &match.State.Cards.Player2Hand
		targetField = &match.State.Cards.Player2Field
		targetDeck = &match.State.Cards.Player2Deck
		targetDiscard = &match.State.Cards.Player2Discard

		oppositeHand = &match.State.Cards.Player1Hand
		oppositeField = &match.State.Cards.Player1Field
		oppositeDiscard = &match.State.Cards.Player1Discard
	}

	inCard := move.Instruction.ToCard()

	// Hack - need to know if the move was a blast so that we can add additional delay leeway
	var usedBlastEffect bool = false

	// Edge case - if we are waiting for draws from both player, we remove from the deck/hand onto the field
	if match.State.Turn == PlayerUndecided {
		if len(*targetDeck) > 0 {
			if !removeLast(*targetDeck) {
				return false, false, PlayerUndecided
			}
		} else {
			if !removeFirstOfType(*targetHand, inCard) {
				return false, false, PlayerUndecided
			}
		}

		*targetField = append(*targetField, inCard)

		if player == Player1 {
			match.Client1.WaitingForMove = false
		} else {
			match.Client2.WaitingForMove = false
		}

		if len(*targetField) == 1 && len(*oppositeField) == 1 {
			updateTurn = true
		} else {
			return true, false, PlayerUndecided
		}
	} else {
		if !removeFirstOfType(*targetHand, inCard) {
			return false, false, PlayerUndecided
		}

		usedRodEffect := inCard == ElliotsOrbalStaff && len(*targetField) > 0 && isBolted(last(*targetField))
		usedBoltEffect := inCard == Bolt && len(*oppositeField) > 0 && !isBolted(last(*oppositeField))
		usedMirrorEffect := inCard == Mirror && len(*targetField) > 0 && len(*oppositeField) > 0
		usedBlastEffect = inCard == Blast && len(*oppositeHand) > 0
		usedNormalOrForceCard := !usedRodEffect && !usedBoltEffect && !usedMirrorEffect && !usedBlastEffect

		// If the selected card was a normal or force card, and the target players latest field card is flipped (bolted) remove it
		if usedNormalOrForceCard && len(*targetField) > 0 && isBolted(last(*targetField)) {
			removedCard := last(*targetField)

			if !removeFirstOfType(*targetField, removedCard) {
				return false, false, PlayerUndecided
			}

			*targetDiscard = append(*targetDiscard, removedCard)
		}

		if len(*targetField) > 0 && !usedNormalOrForceCard {
			if usedBlastEffect {
				blastedCardInt, err := strconv.Atoi(move.Payload)
				if err != nil {
					return false, false, PlayerUndecided
				}

				blastedCard := Card(blastedCardInt)

				if !removeFirstOfType(*oppositeHand, blastedCard) {
					return false, false, PlayerUndecided
				}

				*oppositeDiscard = append(*oppositeDiscard, blastedCard)

			} else if usedRodEffect {
				unBolt(*targetField)
			} else if usedBoltEffect {
				bolt(*oppositeField)
			} else if usedMirrorEffect {
				tempTargetField := *targetField
				*targetField = *oppositeField
				*oppositeField = tempTargetField
			}

			*targetDiscard = append(*targetDiscard, inCard)
		} else {
			*targetField = append(*targetField, inCard)
		}

		// Set wait flags + turn (edge case for blast as the turn doesnt change)
		if usedBlastEffect {
			if match.State.Turn == Player1 {
				match.Client1.WaitingForMove = true
			} else {
				match.Client2.WaitingForMove = true
			}
		} else {
			updateTurn = true
		}
	}

	match.State.Player1Score = calculateScore(match.State.Cards.Player1Field)
	match.State.Player2Score = calculateScore(match.State.Cards.Player2Field)

	if match.State.Turn != PlayerUndecided {
		matchEnded, winner = match.checkForMatchEnd(usedBlastEffect)
		if matchEnded {
			return true, matchEnded, winner
		}
	}

	if updateTurn {
		// Disable wait flags
		match.Client1.WaitingForMove = false
		match.Client2.WaitingForMove = false

		if match.State.Player1Score == match.State.Player2Score {
			match.State.Turn = PlayerUndecided
			match.Client1.WaitingForMove = true
			match.Client2.WaitingForMove = true

			*targetDiscard = append(*targetDiscard, (*targetField)...)
			*targetField = nil

			*oppositeDiscard = append(*oppositeDiscard, (*oppositeField)...)
			*oppositeField = nil

		} else if match.State.Player1Score < match.State.Player2Score {
			match.State.Turn = Player1
			match.Client1.WaitingForMove = true
		} else {
			match.State.Turn = Player2
			match.Client2.WaitingForMove = true
		}
	}

	var nextTurnPeriod time.Duration
	if match.Client1.connection.Latency > match.Client2.connection.Latency {
		nextTurnPeriod = turnMaxWait + match.Client1.connection.Latency
	} else {
		nextTurnPeriod = turnMaxWait + match.Client2.connection.Latency
	}

	if match.State.Player1Score == match.State.Player2Score {
		nextTurnPeriod += tiedScoreAdditionalWait
	} else if usedBlastEffect {
		nextTurnPeriod += blastCardAdditionalWait
	}

	match.turnTimer.Stop()
	match.turnTimer.Reset(nextTurnPeriod)

	return true, false, PlayerUndecided
}

func (match *Match) checkForMatchEnd(usedBlastEffect bool) (matchEnded bool, player Player) {
	if match.isDrawn() {
		return true, PlayerUndecided
	}

	if match.playerHasWon(Player1, usedBlastEffect) {
		return true, Player1
	} else if match.playerHasWon(Player2, usedBlastEffect) {
		return true, Player2
	}

	return false, PlayerUndecided
}

func (match *Match) playerHasWon(player Player, usedBlastEffect bool) bool {
	var targetPlayerScore uint16
	var targetField []Card

	var oppositePlayerScore uint16
	var oppositePlayerHand []Card
	var oppositePlayerField []Card

	var oppositePlayerDeckCount int
	var isOppositePlayersTurn bool

	if match.State.Turn == player {
		isOppositePlayersTurn = false
	} else {
		isOppositePlayersTurn = false
	}

	if player == Player1 {
		targetPlayerScore = match.State.Player1Score
		targetField = match.State.Cards.Player1Field

		oppositePlayerScore = match.State.Player2Score
		oppositePlayerHand = match.State.Cards.Player2Hand
		oppositePlayerField = match.State.Cards.Player2Field
		oppositePlayerDeckCount = len(match.State.Cards.Player2Deck)

	} else if player == Player2 {
		targetPlayerScore = match.State.Player2Score
		targetField = match.State.Cards.Player2Field

		oppositePlayerScore = match.State.Player1Score
		oppositePlayerHand = match.State.Cards.Player1Hand
		oppositePlayerField = match.State.Cards.Player1Field
		oppositePlayerDeckCount = len(match.State.Cards.Player1Deck)
	} else {
		return false
	}

	if len(oppositePlayerHand) > 0 && containsOnlyEffectCards(oppositePlayerHand) {
		return true
	}

	if targetPlayerScore == oppositePlayerScore {
		if oppositePlayerDeckCount+len(oppositePlayerHand) == 0 {
			return true
		}
	}

	if targetPlayerScore > oppositePlayerScore {
		scoreGap := targetPlayerScore - oppositePlayerScore

		if isOppositePlayersTurn && !usedBlastEffect {
			return true
		}

		if len(oppositePlayerHand) == 0 {
			return true
		}

		if canBeatScore(oppositePlayerHand, scoreGap) {
			return false
		}

		if contains(oppositePlayerHand, ElliotsOrbalStaff) {
			if len(oppositePlayerField) > 0 && isBolted(last(oppositePlayerField)) {
				if last(oppositePlayerField) == InactiveForce {
					if oppositePlayerScore*2 >= targetPlayerScore {
						return false
					}
				} else if uint16(getBoltedCardrealValue(last(oppositePlayerField))) >= scoreGap {
					return false
				}
			}
		}

		if contains(oppositePlayerHand, Bolt) {
			if len(targetField) > 0 && !isBolted(last(targetField)) {
				return false
			}
		}

		if contains(oppositePlayerHand, Mirror) {
			return false
		}

		if contains(oppositePlayerHand, Blast) {
			return false
		}

		if contains(oppositePlayerHand, Force) {
			if oppositePlayerScore*2 > targetPlayerScore {
				return false
			}
		}

	}

	return false
}

func (match *Match) isDrawn() bool {
	if len(match.State.Cards.Player1Deck)+len(match.State.Cards.Player2Deck) == 0 {
		if len(match.State.Cards.Player1Hand)+len(match.State.Cards.Player2Hand) == 0 {
			if match.State.Player1Score == match.State.Player2Score {
				return true
			}
		}
	}

	return false
}

func isBolted(card Card) bool {
	return card > Force
}

func bolt(targetField []Card) {
	if len(targetField) > 0 {
		if last(targetField) <= Force {
			targetField[len(targetField)-1] = Card(uint8(last(targetField)) + boltedCardOffset)
		}
	}
}

func unBolt(targetField []Card) {
	if len(targetField) > 0 {
		if last(targetField) >= InactiveElliotsOrbalStaff {
			targetField[len(targetField)-1] = Card(uint8(last(targetField)) - boltedCardOffset)
		}
	}
}

func getBoltedCardrealValue(card Card) uint8 {
	if card > Force {
		card = Card(card - 11)
	}

	return card.Value()
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
