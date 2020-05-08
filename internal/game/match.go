// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game implements the Blade II Online game server.
package game

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/6a/blade-ii-game-server/internal/apiinterface"
	"github.com/6a/blade-ii-game-server/pkg/mathplus"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

const (

	// clientDataDelimiter is the delimiter character used for client data payloads (handle, public ID etc.).
	clientDataDelimiter string = "."

	// payloadDelimiter is the delimiter character used for creating message payloads that contain both a
	// match instruction and some metadata.
	payloadDelimiter string = ":"

	// debugGameID is the match ID for the debug game - this is a special game that is never removed from
	// the server, and never has the result or state updated.
	debugGameID uint64 = 20

	// turnMaxWait is the maximum time to wait for a move from a client before they are considered to have lost by
	// default - that is, they did not play a move within the turn time limit.
	turnMaxWait = time.Millisecond * 21000

	// cardDrawDelay is extra time that is added to the wait timer for the first turn (after the match starts), that
	// takes into account the time taken for the card animation to finish client side, as well as a few extra
	// second to allow for slower computers or networks.
	cardDrawDelay = time.Millisecond * 15000

	// tiedScoreAdditionalWait is an additional delay that is added to the wait timer for a turn when clearing the
	// field after the score is tied, that takes into account the time taken for the card animation to finish client
	// side.
	tiedScoreAdditionalWait = time.Millisecond * 4500

	// tiedScoreAdditionalWait is an additional delay that is added to the wait timer for a turn when a blast
	// card is played, that takes into account the time taken for the card animation to finish client side.
	blastCardAdditionalWait = time.Millisecond * 4500
)

// Match is a wrapper for a matches data and client connections etc
type Match struct {

	// The Match ID
	ID uint64

	// Pointers to both players.
	Client1 *GClient
	Client2 *GClient

	// Match state
	State MatchState

	// A pointer to the game server.
	Server *Server

	// Mutex lock to protect the critical section that can occur when reading/writing to
	// the phase of the game (in State).
	phaseLock sync.Mutex

	// Timer for each player's turn - used to determine if a player has made a move within the alloted time.
	turnTimer *time.Timer
}

// Tick reads any incoming messages and passes outgoing messages to the queue, as well as handling
// any timed out players (players that did not make a move within the turn time limit).
func (match *Match) Tick() {

	// Tick client 1.
	match.tickClient(match.Client1, match.Client2, Player1)

	// Tick client 2.
	match.tickClient(match.Client2, match.Client1, Player2)

	// Check for timeouts for each client if the turn timer channel has something in it (which
	// indicates that the timer has fired). If a player has timed out, end the game, update the
	// state, and terminate the connection(s) accordingly.
	if len(match.turnTimer.C) > 0 {
		select {

		// Read from the channel to drain it.
		case <-match.turnTimer.C:

			// Determine which player(s) timed out.

			if match.Client1.WaitingForMove && match.Client2.WaitingForMove {

				// Both players timed out (such as failing to perform the first draw when the match starts).
				match.Server.Remove(match.Client1, protocol.WSCMatchMutualTimeout, "Both players timed out")
			} else if match.Client1.WaitingForMove {

				// Player 1 was timed out - Set Player 2 as the winner, and remove the match from the server.
				match.State.Winner = match.Client2.DBID
				match.Server.Remove(match.Client1, protocol.WSCMatchTimeOut, "Player 1 timed out")
			} else {

				// Player 2 was timed out - Set Player 1 as the winner, and remove the match from the server.
				match.State.Winner = match.Client1.DBID
				match.Server.Remove(match.Client2, protocol.WSCMatchTimeOut, "Player 2 timed out")
			}

			// Set the match phase to finished.
			match.SetPhase(Finished)
		}
	}
}

// tickClient performs the tick actions for the specified client (client), relative to
// the (other) client. Specify which player this is (player 1 or player 2) by setting
// a value for (player).
func (match *Match) tickClient(client *GClient, other *GClient, player Player) {

	// If the inbound message queue contains messages...
	for len(client.connection.InboundMessageQueue) > 0 {

		// Read the next message from the receive queue.
		message := client.connection.GetNextInboundMessage()

		// If the message is a text message...
		if message.Type == protocol.Type(protocol.WSMTText) {

			// If the message is a move update...
			if message.Payload.Code == protocol.WSCMatchMove {

				// Set the client (the one that is being ticked) to NOT be waiting for a move,
				// preventing the move timer from timing this client out for now.
				client.WaitingForMove = false

				// Parse the incoming move message. Errors will end the game, causing this client
				// to lose (handles in the else branch below).
				move, err := MoveFromString(message.Payload.Message)

				// If there was no error, and the incoming move is considered to be valid given
				// the current state of the game...
				if err == nil && match.isValidMove(move, player) {

					// Update the state of the game. The return values are used below to determine
					// how to continue.
					valid, matchEnded, winner := match.updateMatchState(player, move)

					// If the game state was successfully updated, forward the move to the other client.
					// When (valid) is false, this means that the received move was not valid in the context
					// of the current game state - either the player did something (like fiddling with their data packets?)
					// or something caused some moves to be received out of order.
					if valid {
						// Forward the original message to other client.
						other.SendMessage(message)

						// If the match is determined to have ended...
						if matchEnded {

							// Determine which player won (if any).
							if winner == Player1 {

								// Player 1 was the winner - set the winner and remove this match from the server.
								match.State.Winner = match.Client1.DBID
								match.Server.Remove(match.Client1, protocol.WSCMatchWin, "")
							} else if winner == Player2 {

								// Player 2 was the winner - set the winner and remove this match from the server.
								match.State.Winner = match.Client2.DBID
								match.Server.Remove(match.Client2, protocol.WSCMatchWin, "")
							} else {

								// Neither player won - that match ended in a draw. Remove this match from the server,
								// without setting a winner, so that the server can correctly identify that the game
								// ended in a draw.
								match.Server.Remove(match.Client1, protocol.WSCMatchDraw, "")
							}

							// Set the match phase to finished.
							match.SetPhase(Finished)
						}
					} else {

						// Remove the offending client (this will also end the game) and set the winner
						// to the other client.
						match.State.Winner = other.DBID
						match.Server.Remove(client, protocol.WSCMatchIllegalMove, "")
					}
				} else {

					// Remove the offending client (this will also end the game) and set the winner
					// to the other client.
					match.State.Winner = other.DBID
					match.Server.Remove(client, protocol.WSCMatchIllegalMove, "")
				}
			} else if message.Type == protocol.Type(protocol.WSCMatchForfeit) {

				// Remove the forfeiting client (this will also end the game) and set the winner
				// to the other client.
				match.State.Winner = other.DBID
				match.Server.Remove(client, protocol.WSCMatchForfeit, "")
			} else if message.Type == protocol.Type(protocol.WSCMatchRelayMessage) {

				// If we reach this point, the payload was just a message that should be
				// relayed to the other client.

				// TODO add filtering? Profanity check? Something to ensure nothing naughty
				// reaches the other client...

				other.SendMessage(message)
			}
		} else {
			// Handle non-text messages?
		}
	}
}

// BroadCast sends the specified message to both clients.
func (match *Match) BroadCast(message protocol.Message) {

	// Add the same message to both clients.
	match.Client1.SendMessage(message)
	match.Client2.SendMessage(message)
}

// SendCardData sends starting card data to each client.
func (match *Match) SendCardData(cards string) {

	// Create two string builders, one for each player.
	var client1Buffer strings.Builder
	var client2Buffer strings.Builder

	// Write the player number, card data delimiter, and then the serialized card data, to player 1's string builder.
	client1Buffer.WriteString("0")
	client1Buffer.WriteString(SerializedCardsDelimiter)
	client1Buffer.WriteString(cards)

	// Write the player number, card data delimiter, and then the serialized card data, to player 2's string builder.
	client2Buffer.WriteString("1")
	client2Buffer.WriteString(SerializedCardsDelimiter)
	client2Buffer.WriteString(cards)

	// Send the data to both players
	match.sendMatchData(client1Buffer, client2Buffer, InstructionCards)
}

// SendPlayerData sends each player's (their own) name to the respective client.
func (match *Match) SendPlayerData() {

	// Create two string builders, one for each player.
	var client1Buffer strings.Builder
	var client2Buffer strings.Builder

	// Write the player's display name, client data delimiter, and then the player's avatar ID (as a string), to player 1's string builder.
	client1Buffer.WriteString(match.Client1.DisplayName)
	client1Buffer.WriteString(clientDataDelimiter)

	// Note the conversion to an int before the call to Itoa.
	client1Buffer.WriteString(strconv.Itoa(int(match.Client1.Avatar)))

	// Write the player's display name, client data delimiter, and then the player's avatar ID (as a string), to player 2's string builder.
	client2Buffer.WriteString(match.Client2.DisplayName)
	client2Buffer.WriteString(clientDataDelimiter)

	// Note the conversion to an int before the call to Itoa.
	client2Buffer.WriteString(strconv.Itoa(int(match.Client2.Avatar)))

	// Send the data to both players
	match.sendMatchData(client1Buffer, client2Buffer, InstructionPlayerData)
}

// SendOpponentData sends each player the opponents data.
func (match *Match) SendOpponentData() {

	// Create two string builders, one for each player.
	var client1Buffer strings.Builder
	var client2Buffer strings.Builder

	// Build the two string builders with the following format:
	// <other player display name><delim><other player public ID><delim><other player avatar ID>

	// Player 1's string builder.
	client1Buffer.WriteString(match.Client2.DisplayName)
	client1Buffer.WriteString(clientDataDelimiter)
	client1Buffer.WriteString(match.Client2.PublicID)
	client1Buffer.WriteString(clientDataDelimiter)

	// Note the conversion to an int before the call to Itoa.
	client1Buffer.WriteString(strconv.Itoa(int(match.Client2.Avatar)))

	// Player 2's string builder.
	client2Buffer.WriteString(match.Client1.DisplayName)
	client2Buffer.WriteString(clientDataDelimiter)
	client2Buffer.WriteString(match.Client1.PublicID)
	client2Buffer.WriteString(clientDataDelimiter)

	// Note the conversion to an int before the call to Itoa.
	client2Buffer.WriteString(strconv.Itoa(int(match.Client1.Avatar)))

	// Send the data to both players
	match.sendMatchData(client1Buffer, client2Buffer, InstructionOpponentData)
}

// SetMatchStart sets the phase + start time for the current match.
//
// Fails silently but logs errors.
//
// Database update is performed in a goroutine to avoid a delay when updating
// the database.
func (match *Match) SetMatchStart() {

	// Set the match to the play state.
	match.SetPhase(Play)

	// Start turn timer to a suitable value, that should allow for loading, drawing, and any network delays
	// client side.
	match.turnTimer = time.NewTimer(turnMaxWait + cardDrawDelay)

	// Set both players to be waiting for a move - as we are waiting for their initial draw from the deck.
	match.Client1.WaitingForMove = true
	match.Client2.WaitingForMove = true

	// Early exit if we are currently in the debug match (don't write to the db).
	if match.ID == debugGameID {
		return
	}

	// Using a goroutine, update the match phase.
	go func() {

		// Update the match phase in the database,
		err := database.SetMatchStart(match.ID)
		if err != nil {

			// On error, print to log but don't handle it.
			log.Printf("Failed to update match phase: %s", err.Error())
		}
	}()
}

// SetMatchResult updates the database with the match result, and also
// updates the match stats for each player via the Blade II Online REST API.
//
// Fails silently but logs errors.
//
// Performed in a goroutine.
func (match *Match) SetMatchResult() {

	// Early exit if we are currently in the debug match (don't write to the db).
	if match.ID == debugGameID {
		return
	}

	// Using a goroutine, update the database and send off the match stats update request.
	go func() {

		// Update the match in the database.
		err := database.SetMatchResult(match.ID, match.State.Winner)
		if err != nil {

			// On error, print to log but don't handle it.
			log.Printf("Failed to update match result: %s", err.Error())
		}

		// Determine the winner of the match.
		var winner apiinterface.Winner
		if match.State.Winner == match.Client1.DBID {
			winner = apiinterface.Player1
		} else if match.State.Winner == match.Client2.DBID {
			winner = apiinterface.Player2
		} else {
			winner = apiinterface.Draw
		}

		// Send the match update request to the Blade II Online REST API. This blocks,
		// hence the goroutine.
		apiinterface.UpdateMatchStats(match.Client1.DBID, match.Client2.DBID, winner)
	}()
}

// SetPhase sets the match phase, using a mutex lock to protect the critical section,
// as multiple goroutines may be trying to read the matches phase.
func (match *Match) SetPhase(phase Phase) {

	// Lock the mutex lock, and then defer unlocking.
	match.phaseLock.Lock()
	defer match.phaseLock.Unlock()

	// Set the value of State.Phase. After the function exits, the lock will be
	// released.
	match.State.Phase = phase

}

// GetPhase gets the match phase
func (match *Match) GetPhase() Phase {

	// Lock the mutex lock, and then defer unlocking.
	match.phaseLock.Lock()
	defer match.phaseLock.Unlock()

	// Return the value of State.Phase. After the function exits, the lock will be
	// released.
	return match.State.Phase
}

// sendMatchData is a helper function that sends match data, based on the two string builders provided, to the respective clients, with
// the specified instruction.
func (match *Match) sendMatchData(client1Buffer strings.Builder, client2Buffer strings.Builder, instruction B2MatchInstruction) {

	// Package player 1's string builder as a string along with the specified B2MatchInstruction.
	client1MessageString := makeMessageString(instruction, client1Buffer.String())

	// Send the packaged message to player 1.
	match.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchData, client1MessageString))

	// Package player 2's string builder as a string along with the specified B2MatchInstruction.
	client2MessageString := makeMessageString(instruction, client2Buffer.String())

	// Send the packaged message to player 2.
	match.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchData, client2MessageString))
}

// makeMessageString is a helper function that returns a string representation of a message payload
// to send to a client.
//
// Format: <instruction><delim><data>
func makeMessageString(instruction B2MatchInstruction, data string) string {

	// Create a string builder.
	var builder strings.Builder

	// Write the instruction number, the delimiter, and the data to the string builder. Note the conversion
	// to an int before the call to Itoa.
	builder.WriteString(strconv.Itoa(int(instruction)))
	builder.WriteString(payloadDelimiter)
	builder.WriteString(data)

	// Return the string builder as a string.
	return builder.String()
}

// updateMatchState takes a move that the specified player made, and updates the match state accordingly.
//
// Returns a bool indicating whether the operation was a success, another indicating whether the match ended due
// to the most, and one more that indicates which player won (if any).
func (match *Match) updateMatchState(player Player, move Move) (validMove bool, matchEnded bool, winner Player) {

	// Declare variables that will store pointers to the target player's (the one that made the move) cards.
	var targetHand *[]Card
	var targetField *[]Card
	var targetDeck *[]Card
	var targetDiscard *[]Card

	// Declare variables that will store pointers to the other player's (the one that DID NOT make the move) cards.
	var oppositeHand *[]Card
	var oppositeField *[]Card
	var oppositeDiscard *[]Card

	// Note: using pointers allows us to use the same function for both players, and still be able to modify the
	// modify - Using non-pointers will change the local array, but the original array will remain unchanged for
	// operations that don't involve reordering the members by shifting array members.

	// A variable to store the score of the player that made the move.
	var targetScore uint16

	// Whether or not to update the turn. E.g. when a blast is played, the turn does not change.
	var updateTurn = false

	// Depending on which player made the move, set values for each card array pointer.

	if player == Player1 {
		targetHand = &match.State.Cards.Player1Hand
		targetField = &match.State.Cards.Player1Field
		targetDeck = &match.State.Cards.Player1Deck
		targetDiscard = &match.State.Cards.Player1Discard
		targetScore = match.State.Player1Score

		oppositeHand = &match.State.Cards.Player2Hand
		oppositeField = &match.State.Cards.Player2Field
		oppositeDiscard = &match.State.Cards.Player2Discard
	} else {
		targetHand = &match.State.Cards.Player2Hand
		targetField = &match.State.Cards.Player2Field
		targetDeck = &match.State.Cards.Player2Deck
		targetDiscard = &match.State.Cards.Player2Discard
		targetScore = match.State.Player2Score

		oppositeHand = &match.State.Cards.Player1Hand
		oppositeField = &match.State.Cards.Player1Field
		oppositeDiscard = &match.State.Cards.Player1Discard
	}

	// Get the type of card that was played.
	inCard := move.Instruction.ToCard()

	// Hack - need to know if the move was a blast so that we can add some additional time to the turn timer, to
	// account for the long blast animation.
	var usedBlastEffect bool = false

	// If the turn is currently undecided, this means that the board was cleared and we are waiting for both, or
	// just one players draw from the deck.
	if match.State.Turn == PlayerUndecided {

		// If the target7s deck has some cards in it, try to remove one. If it fails (it shouldnt), return false.
		// Otherwise, try to draw from the player hand. Again, it shouldnt fail, but it if does, return false.
		if len(*targetDeck) > 0 {
			if !removeLast(targetDeck) {
				return false, false, PlayerUndecided
			}
		} else {
			if !removeFirstOfType(targetHand, inCard) {
				return false, false, PlayerUndecided
			}
		}

		// Now that a card has been selected, add it to the target player's field.
		*targetField = append(*targetField, inCard)

		// For the player that made the move, set them to be longer waiting for a move so
		// that they do not get timed out.
		if player == Player1 {
			match.Client1.WaitingForMove = false
		} else {
			match.Client2.WaitingForMove = false
		}

		// Determine if both players have made their draw from the deck/hand to the field,
		// by checking that both fields have one card on them. If this is the case, set the update
		// flag to true, so that further processing occurs below.
		// Otherwise exit early, indicating that the match has not ended.
		if len(*targetField) == 1 && len(*oppositeField) == 1 {
			updateTurn = true
		} else {
			return true, false, PlayerUndecided
		}
	} else {

		// Reaching this point means that the turn is NOT undecided - i.e. it is someones turn. Try to remove the
		// first instance of the played card from the target players hand. If this fails, the player sent some bad
		// data, or the game state on their client was wrong / messed with, and we return false.
		if !removeFirstOfType(targetHand, inCard) {
			return false, false, PlayerUndecided
		}

		// Initialise boolean values that, based on the state of the game, are set to true if a particular effect
		// card has been played AND THE EFFECT ACTUALLY ACTIVATED. Note that (usedBlastEffect) is declared earlier,
		// as a hack to ensure that the value can be reused later for the blast edge case.
		usedRodEffect := inCard == ElliotsOrbalStaff && len(*targetField) > 0 && isBolted(last(*targetField))
		usedBoltEffect := inCard == Bolt && len(*oppositeField) > 0 && !isBolted(last(*oppositeField))
		usedMirrorEffect := inCard == Mirror && len(*targetField) > 0 && len(*oppositeField) > 0
		usedBlastEffect = inCard == Blast && len(*oppositeHand) > 0 // Note: Variable declared above -> See above comment.
		usedForceEffect := inCard == Force && targetScore > 0

		// Set a separate bool that is used to quickly check if a force, or a normal card was played.
		usedNormalOrForceCard := (!usedRodEffect && !usedBoltEffect && !usedMirrorEffect && !usedBlastEffect) || usedForceEffect

		// If the selected card was a normal or force card, and the target players latest field card is flipped (bolted),
		// remove it.
		if usedNormalOrForceCard && len(*targetField) > 0 && isBolted(last(*targetField)) {

			// Get a copy of the last card on the target player's field (the one about to be removed).
			removedCard := last(*targetField)

			// Attempt to remove the last card from the target player's field. A failure indicates that a bad instruction was
			// received, so return false.
			if !removeFirstOfType(targetField, removedCard) {
				return false, false, PlayerUndecided
			}

			// Add the removed card to the target player's discard pile.
			*targetDiscard = append(*targetDiscard, removedCard)
		}

		// Determine if the card was an effect card that had its effect activated. Otherwise
		// just treat it like a normal card, placing it on the target player's field.
		if len(*targetField) > 0 && !usedNormalOrForceCard {

			// As mentioned earlier - the blast flag is checked here to handle the blast edge case.
			if usedBlastEffect {

				// Parse the move payload, as it should contain the type (as a string) of the
				// card that the target player selected to blast (from the other player's) hand.
				// If it errors, the payload was improperly formatted, empty, or otherwise invalid.
				blastedCardInt, err := strconv.Atoi(move.Payload)
				if err != nil {
					return false, false, PlayerUndecided
				}

				// If the above parse was successful, determine which card the payload contained by
				// casting it to a Card.
				blastedCard := Card(blastedCardInt)

				// Using the card determined aboved, attempt to remove the first instance of that
				// card from the other players hand. A failure here suggests that the payload had
				// the wrong value for whatever reason, so we return false.
				if !removeFirstOfType(oppositeHand, blastedCard) {
					return false, false, PlayerUndecided
				}

				// If the above removal call was a success, append the card that was blasted to the other
				// player's discard pile.
				*oppositeDiscard = append(*oppositeDiscard, blastedCard)

			} else if usedRodEffect {

				// If a rod effect was detected, unbolt the bolted card on the target player's field.
				unBolt(targetField)
			} else if usedBoltEffect {

				// If a bolt effect was detected, bolt the bolted card on the other player's field.
				bolt(oppositeField)
			} else if usedMirrorEffect {

				// If a mirror effect was detected, switch the fields for each player. To do so, the
				// target player's field is first stored in a temporary variable. Then, the target player's
				// field is overwritten with the other player's field. Finally, the other player's field
				// is overwritten with the cards stored in the temporary variable.
				tempTargetField := *targetField
				*targetField = *oppositeField
				*oppositeField = tempTargetField
			}

			// Finally, add the card that the target player played to the target player's discard pile.
			*targetDiscard = append(*targetDiscard, inCard)
		} else {

			// Reaching this point means we handle the incoming move as a standard play, and add it directly
			// to the target player's field.
			*targetField = append(*targetField, inCard)
		}

		// If a blast effect was used, set the appropriate wait flag. Otherwise, it was NOT a blast card, and the
		// update turn flag is set to true.
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

	// Update the score for both players.
	match.State.Player1Score = calculateScore(match.State.Cards.Player1Field)
	match.State.Player2Score = calculateScore(match.State.Cards.Player2Field)

	// If the match state is NOT undecided, see if one of the players won. This is done here, and not in the previous
	// if else statement because we need to update the score first.
	if match.State.Turn != PlayerUndecided {
		matchEnded, winner = match.checkForMatchEnd(usedBlastEffect)
		if matchEnded {
			return true, matchEnded, winner
		}
	}

	// If the update turn flag was set to true, we need to determine which player's turn it now is.
	if updateTurn {

		// Disable wait flags for both players.
		match.Client1.WaitingForMove = false
		match.Client2.WaitingForMove = false

		// If the scores are tied, clear the board and enter the undecided state. Otherwise determine
		// who's turn it now is based on the scores.
		if match.State.Player1Score == match.State.Player2Score {

			// Set the turn to undecided.
			match.State.Turn = PlayerUndecided

			// Set both wait flags to true.
			match.Client1.WaitingForMove = true
			match.Client2.WaitingForMove = true

			// Dump the target player's field into the target player's discard pile.
			*targetDiscard = append(*targetDiscard, (*targetField)...)
			*targetField = nil

			// And do the same for the other player.
			*oppositeDiscard = append(*oppositeDiscard, (*oppositeField)...)
			*oppositeField = nil

		} else if match.State.Player1Score < match.State.Player2Score {

			// Set the turn to player 1 and set their wait flag.
			match.State.Turn = Player1
			match.Client1.WaitingForMove = true
		} else {

			// Set the turn to player 2 and set their wait flag.
			match.State.Turn = Player2
			match.Client2.WaitingForMove = true
		}
	}

	// Calculate how long the next turn timeout should be, be taking the base value
	// and adding the maximum latency of the two clients. If one player has a particularly
	// high latency, this will give them some leeway to account for it.
	var nextTurnPeriod = turnMaxWait + mathplus.MaxDuration(match.Client1.connection.Latency, match.Client2.connection.Latency)

	// If the scores are drawn, add some extra time to account for clearing the board. Or, the move was a blast card, add
	// some time to account for the client side animations.
	if match.State.Player1Score == match.State.Player2Score {
		nextTurnPeriod += tiedScoreAdditionalWait
	} else if usedBlastEffect {
		nextTurnPeriod += blastCardAdditionalWait
	}

	// Reset the turn timer with the newly calculated turn wait time.
	match.turnTimer.Stop()
	match.turnTimer.Reset(nextTurnPeriod)

	// Return true, with no winner.
	return true, false, PlayerUndecided
}

// checkForMatchEnd returns true, and the player who won, if the match is no longer in a playable
// state (though not due to error - rather, due to someone winning, or a draw). Pass in a bool
// that indicates whether or not a blast effect was used on this turn (as it doesnt cause the
// turn to change, it has to be handled as an edge case).
func (match *Match) checkForMatchEnd(usedBlastEffect bool) (matchEnded bool, player Player) {

	// Return true and undecided if the match is drawn.
	if match.isDrawn() {
		return true, PlayerUndecided
	}

	// Return true, and player 1 if player 1 won.
	if match.playerHasWon(Player1, usedBlastEffect) {
		return true, Player1
	}

	// Return true, and player 2 if player 2 won.
	if match.playerHasWon(Player2, usedBlastEffect) {
		return true, Player2
	}

	// If none of the win conditions were met, the match is still in progress - return false.
	return false, PlayerUndecided
}

// Based on the state of the game (and whether or not a blast cased was played on this
// turn) determine if the specified player won the game.
func (match *Match) playerHasWon(player Player, usedBlastEffect bool) bool {

	// Similar to the updateMatchState function, this function takes in a player
	// as an argument, so we first declare some variables to set once we have
	// determined which player is the target, and which is the "opposite" player.

	// Unlike the updateMatchState function, however, the arrays are not pointers, as
	// the arrays will not be modified.

	var targetPlayerScore uint16
	var targetField []Card

	var oppositePlayerScore uint16
	var oppositePlayerHand []Card
	var oppositePlayerField []Card

	var oppositePlayerDeckCount int
	var isOppositePlayersTurn bool

	// Determine which player's turn it is, and store it in a local variable.
	if match.State.Turn == player {
		isOppositePlayersTurn = false
	} else {
		isOppositePlayersTurn = true
	}

	// Depending on which player is currently beind checked to see if they won, set values
	// for the previously declared variables.
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

		// Edge case - if a non player (undecided) was passed in, exit early.
		return false
	}

	// Early exit if the opponent only has effect cards left, as this is an auto win regardless.
	// Skips check if the opposite players hand is empty.
	if len(oppositePlayerHand) > 0 && containsOnlyEffectCards(oppositePlayerHand) {
		return true
	}

	// Early exit if the scores are equal, and the opposite player has no more cards left to
	// try and break a tie. Ensure that the game is checked for a draw state BEFORE this function, as
	// this check does not check the target player's side of the field.
	if targetPlayerScore == oppositePlayerScore {
		if oppositePlayerDeckCount+len(oppositePlayerHand) == 0 {
			return true
		}
	}

	// Early exit if a blast was used, and the opposite player now has no cards left in their hand (and are
	// therefore unable to continue).
	if usedBlastEffect {
		if len(oppositePlayerHand) == 0 {
			return true
		}
	}

	// If the target player's score is greater than the opposite player's score...
	if targetPlayerScore > oppositePlayerScore {

		// Determine the score gap that must be overcome, or equalled, in order for the opposite player
		// to be able to continue. No need to abs or cast to signed values, as we can only enter
		// this clause if (targetPlayerScore) is greater than (oppositePlayerScore).
		scoreGap := targetPlayerScore - oppositePlayerScore

		// If the opposite players score is lower than the target player's score, and the opposite player
		// did NOT player a blast card, they have lost as they failed to beat the score for the their turn.
		// Blast effects are an edge case, as it does not change the turn.
		if isOppositePlayersTurn && !usedBlastEffect {
			return true
		}

		// If the opposite player's hand is empty, they will not be able to counter the most recent move, and
		// therefore have lost.
		if len(oppositePlayerHand) == 0 {
			return true
		}

		// From here we check various conditions to see if the opposite player is able to make a valid move.

		// If the opposite player has a card in their hand that will overcome or match the target player's
		// score, they are ok to continue.
		if canOvercomeDifference(oppositePlayerHand, scoreGap) {
			return false
		}

		// If opposite player has an rod card in their hand, and are able to play it, and playing it would cause their new score to
		// be equal to or greater than the target score, they are ok to continue.
		if contains(oppositePlayerHand, ElliotsOrbalStaff) {

			// If the opposite player's field has at least one card, and the last card is bolted...
			if len(oppositePlayerField) > 0 && isBolted(last(oppositePlayerField)) {

				// If the bolted card is a force card, and applying the force effect would overcome the
				// difference, they are ok. Or, if the bolted card has a high enough value to overcome
				// the difference, that's also ok.
				if last(oppositePlayerField) == InactiveForce {
					if oppositePlayerScore*2 >= targetPlayerScore {
						return false
					}
				} else if uint16(getBoltedCardrealValue(last(oppositePlayerField))) >= scoreGap {
					return false
				}
			}
		}

		// If the opposite player has a bolt card in their hand, and the target player's last field card
		// can be bolted, they are ok to continue.
		if contains(oppositePlayerHand, Bolt) {
			if len(targetField) > 0 && !isBolted(last(targetField)) {
				return false
			}
		}

		// If the opposite player has a mirror card in their hand, they are ok.
		if contains(oppositePlayerHand, Mirror) {
			return false
		}

		// If the opposite player has a blast card in their hand, they are ok.
		if contains(oppositePlayerHand, Blast) {
			return false
		}

		// If the opposite player has a force card in their hand, and playing it would increase their
		// score so that it matches or beats the target player's score, they are ok.
		if contains(oppositePlayerHand, Force) {
			if oppositePlayerScore*2 > targetPlayerScore {
				return false
			}
		}
	}

	// Reaching this point indicates that none of the conditions were event explored, and the target player
	// has not won in any fashion.
	return false
}

// isDrawn returns true if the scores are drawn, and both players are unable to make more moves.
func (match *Match) isDrawn() bool {

	// If both decks are empty...
	if len(match.State.Cards.Player1Deck)+len(match.State.Cards.Player2Deck) == 0 {

		// If both hands are empty...
		if len(match.State.Cards.Player1Hand)+len(match.State.Cards.Player2Hand) == 0 {

			// If both scores are equal...
			if match.State.Player1Score == match.State.Player2Score {

				// The match is drawn.
				return true
			}
		}
	}

	return false
}

// isBolted returns true if the specified card is bolted.
func isBolted(card Card) bool {

	// A enum value greater than force means that the card is one of the InactiveX cards, and is
	// therefore bolted.
	return card > Force
}

// getBoltedCardrealValue returns the value of a bolted card, if it was to be
// unbolted. If an unbolted card is passed in, its value will be returned
// as is.
func getBoltedCardrealValue(card Card) uint8 {

	// A enum value greater than force means that the card is one of the InactiveX cards, and is
	// therefore bolted. If the card is not bolted, it is unchanged.
	if card > Force {

		// The card is first "unbolted" by subtracting the bolted card offset from its value.
		// Note the cast to uint8 to allow the math operation, following by a cast back to a Card.
		card = Card(uint8(card) - boltedCardOffset)
	}

	// Return the value of the card.
	return card.Value()
}

// isValidMove returns true if the specified move is a valid move, for the specified player to make,
// based on the current state of the match.
//
// Note - only partially implemented, but a lot of the validity checking is performed in the state
// update function.
func (match *Match) isValidMove(move Move, player Player) bool {

	// Early exit if the player tried to make a move during the other players turn.
	if match.State.Turn != player && match.State.Turn != PlayerUndecided {
		return false
	}

	// Reaching this point means that the move is valid.
	return true
}

// NewMatch creates and returns a pointer to a new match, setting the specified client as player 1.
func NewMatch(matchID uint64, client *GClient, server *Server) *Match {

	// Create a new match, and store its address in a new variable
	match := &Match{
		ID:      matchID,
		Client1: client,
		Server:  server,
	}

	// Return the pointer to the new match.
	return match
}
