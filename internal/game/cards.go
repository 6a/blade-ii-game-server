// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game implements the Blade II Online game server.
package game

import (
	"bytes"
	"math"
	"math/rand"
	"strconv"
)

const (

	// SerializedCardsDelimiter is the delimiter for serialized cards objects.
	SerializedCardsDelimiter string = "."

	// maxDrawsOnStart is the maximum number of times the initial draw from deck to field can result in a tied score
	// before the set of cards is considered to be invalid.
	maxDrawsOnStart uint8 = 3

	// postInitialisationDeckSize is the size of the deck after the intitial state of the match is initialised -
	// i.e. after all the cards are placed onto the field, and each player draws from their deck into their hands.
	postInitialisationDeckSize uint8 = 5

	// startingHandSize is the initial size of a players hand when the match starts.
	startingHandSize uint8 = 10

	// startingDeckSize is the intitial size of a players deck when the match starts, before the cards are dealt.
	startingDeckSize uint8 = 15
)

// Cards is a container for all the cards on the field.
type Cards struct {
	Player1Deck    []Card
	Player1Hand    []Card
	Player1Field   []Card
	Player1Discard []Card

	Player2Deck    []Card
	Player2Hand    []Card
	Player2Field   []Card
	Player2Discard []Card
}

// Serialized returns the string representation of the DECKS ONLY, as the other data is never required to be sent.
//
// The cards are serialized as hexadecimal numbers, with the following format:
//
// NNNNNNNNNNNNNNN.NNNNNNNNNNNNNNN
//
// Where each "N" is the hexadecimal representation of a card.
func (c *Cards) Serialized() string {

	// Create an empty buffer to save on string operation costs.
	var buffer bytes.Buffer

	// For each card in player 1's deck, write a hex string representation of the card
	// to the buffer.
	for _, card := range c.Player1Deck {
		buffer.WriteString(strconv.FormatUint(uint64(card), 16))
	}

	// Append the appropriate delimiter.
	buffer.WriteString(SerializedCardsDelimiter)

	// For each card in player 2's deck, write a hex string representation of the card
	// to the buffer.
	for _, card := range c.Player2Deck {
		buffer.WriteString(strconv.FormatUint(uint64(card), 16))
	}

	// Return the contents of the buffer as a string.
	return buffer.String()
}

// GenerateCards generates a new set of cards for a match - has additional checks to ensure that the match is not
// unwinnable from the first move etc.
func GenerateCards() (cards Cards) {

	// Generate all the cards (ref: https://www.reddit.com/r/Falcom/comments/fxt5nq/can_i_buy_the_card_game_blade_anywhere/fmxo8qo/)
	// that will be used to create the deck for a match. This is not stored as a const as a. its pretty large and unsightly, and b.
	// it would need to get copied each time anyway as we will be modifying arrays that point to sections within this pool.
	pool := []Card{
		ElliotsOrbalStaff, ElliotsOrbalStaff,
		FiesTwinGunswords, FiesTwinGunswords, FiesTwinGunswords, FiesTwinGunswords, FiesTwinGunswords,
		AlisasOrbalBow, AlisasOrbalBow, AlisasOrbalBow, AlisasOrbalBow, AlisasOrbalBow,
		JusisSword, JusisSword, JusisSword, JusisSword, JusisSword,
		MachiasOrbalShotgun, MachiasOrbalShotgun, MachiasOrbalShotgun, MachiasOrbalShotgun,
		GaiusSpear, GaiusSpear, GaiusSpear,
		LaurasGreatsword, LaurasGreatsword,
		Bolt, Bolt, Bolt, Bolt,
		Mirror, Mirror, Mirror, Mirror,
		Blast, Blast, Blast, Blast,
		Force, Force,
	}

	// Iterate until a valid set of cards is generated. While there is always a danger of infinite looping here,
	// the chances of the algorithm failing to find a deck more than a few times is infinitesimally small.
	var success = false
	for !success {

		// Generate a permutation based on the size of the card pool. This gives us an array with a set of
		// integers representing each index of the pool array, in random order.
		permutation := rand.Perm(len(pool))

		// Create an empty Card object to fill later.
		cards = Cards{}

		// Fill player 1's deck using the first 15 members (0 ->14) of the permutation array.
		for i := uint8(0); i < startingDeckSize; i++ {
			cards.Player1Deck = append(cards.Player1Deck, pool[permutation[i]])
		}

		// Fill player 2's deck using the next 15 members (15 -> 29) of the permutation array.
		for i := uint8(15); i < startingDeckSize*2; i++ {
			cards.Player2Deck = append(cards.Player2Deck, pool[permutation[i]])
		}

		// Check the validity of the cards that were selected. A result of true will cause the loop to exit.
		success = validateCards(&cards)
	}

	// Reaching this point means a valid set of cards has been found - so return the set.
	return cards
}

// InitializeCards simulates the first moves of the game until a playable state is reached.
//
// Returns a COPY of the input cards.
func InitializeCards(inCards Cards) (outCards Cards) {

	// Make a copy of the the input so that the original cards object is not modified.
	// While the parameter is passed as a copy, it contains arrays which must be deep copied.
	outCards = inCards.Copy()

	// Copy the all the cards after the first 5, from player 1's deck to player 1's hand.
	outCards.Player1Hand = outCards.Player1Deck[postInitialisationDeckSize:]

	// Reverse the cards in player 1's hand.
	reverseCardArray(outCards.Player1Hand)

	// Trim player 1's deck so that it contains only the first 5 cards.
	outCards.Player1Deck = outCards.Player1Deck[:postInitialisationDeckSize]

	// Copy the all the cards after the first 5, from player 2's deck to player 2's hand.
	outCards.Player2Hand = outCards.Player2Deck[postInitialisationDeckSize:]

	// Reverse the cards in player 2's hand.
	reverseCardArray(outCards.Player2Hand)

	// Trim player 2's deck so that it contains only the first 5 cards.
	outCards.Player2Deck = outCards.Player2Deck[:postInitialisationDeckSize]

	// Return the initialised cards
	return outCards
}

// Copy returns a deep copy of this cards object.
func (c Cards) Copy() (outCards Cards) {

	// Initialise all the fields - this is required for the copy operation
	// to work properly, as it will only copy (len(destination)) values from
	// the source array into the destination array.
	outCards.Player1Deck = make([]Card, len(c.Player1Deck))
	outCards.Player1Hand = make([]Card, len(c.Player1Hand))
	outCards.Player1Field = make([]Card, len(c.Player1Field))
	outCards.Player1Discard = make([]Card, len(c.Player1Discard))
	outCards.Player2Deck = make([]Card, len(c.Player2Deck))
	outCards.Player2Hand = make([]Card, len(c.Player2Hand))
	outCards.Player2Field = make([]Card, len(c.Player2Field))
	outCards.Player2Discard = make([]Card, len(c.Player2Discard))

	// Copy all of player 1's cards from the source, into the (outCards) object.
	copy(outCards.Player1Deck, c.Player1Deck)
	copy(outCards.Player1Hand, c.Player1Hand)
	copy(outCards.Player1Field, c.Player1Field)
	copy(outCards.Player1Discard, c.Player1Discard)

	// Copy all of player 2's cards from the source, into the (outCards) object.
	copy(outCards.Player2Deck, c.Player2Deck)
	copy(outCards.Player2Hand, c.Player2Hand)
	copy(outCards.Player2Field, c.Player2Field)
	copy(outCards.Player2Discard, c.Player2Discard)

	// return the copy of this cards object.
	return outCards
}

// validateCards returns true if the current cards will NOT result in a bad game state, such as an insta-loss, or more
// requires more than "maxDrawsOnStart" draws in order to reach a playable state.
func validateCards(cards *Cards) (valid bool) {

	// Iterate until (maxDrawsOnStart) is reached.
	for i := uint8(0); i < maxDrawsOnStart; i++ {

		// Get the index of the card that will be checked from each deck. This starts at i = 4.
		cardIndex := postInitialisationDeckSize - 1 - i

		// This should never be hit, but it's here incase somebody changes (postInitialisationDeckSize) to a bad value.
		if cardIndex < 0 {
			break
		}

		// Ensure that play starts within "maxDrawsOnStart" draws - a non zero value indicates that the scores will be different.
		// This is done by taking the values of each target card, and determining the score difference between them.
		player1Score := cards.Player1Deck[cardIndex].Value()
		player2Score := cards.Player2Deck[cardIndex].Value()

		// Note the cast to int16 to avoid underflow.
		// Go only has a float abs so the non-abs difference is cast to a float 64 first.
		scoreDifference := math.Abs(float64(int16(player1Score) - int16(player2Score)))

		// If the score difference is NOT zero, it means that the current draw will result in a playable state.
		if scoreDifference != 0 {

			// We also must ensure that the score difference from the opponent hand is beatable by the player that goes first
			// (if their hand does not have any cards of high enough value they will insta-lose otherwise).

			// First, initialize some variables to be passed into the validation function.

			// Create a card array to store the hand of the player that will be going first.
			var firstMovePlayerHand []Card

			// A variable to store the card that needs to be beaten by (handToCheck).
			var cardToBeatOrMatch Card

			// A variable to store the current score of the player that will be going first.
			var scoreToCheck uint8

			// Set the values for the three variables initialized above.
			if player1Score < player2Score {
				firstMovePlayerHand = cards.Player1Deck[postInitialisationDeckSize:]
				scoreToCheck = player1Score
				cardToBeatOrMatch = cards.Player2Deck[cardIndex]
			} else {
				firstMovePlayerHand = cards.Player2Deck[postInitialisationDeckSize:]
				scoreToCheck = player2Score
				cardToBeatOrMatch = cards.Player1Deck[cardIndex]
			}

			// Return with the result of the validation function.
			return validFirstMoveAvailable(firstMovePlayerHand, cardToBeatOrMatch, scoreToCheck)
		}
	}

	return false
}

// validFirstMoveAvailable returns true if the specified set of cards contains one that can beat the specified
// card (first move only, i.e. after initial draw from deck), when the player going first has a score of (currentScore)
func validFirstMoveAvailable(hand []Card, cardToBeatOrMatch Card, currentScore uint8) bool {

	// Declare a variable to store the hand with the target card (the card being checked to see if playing
	// it puts the game in a playable state) removed.
	var cardSetWithoutCurrent []Card

	// Iterate over the hand
	for i := 0; i < len(hand); i++ {

		// Store the hand with the target card (the card being checked to see if playing it puts the
		// game in a playable state) removed.
		cardSetWithoutCurrent = append(hand[:i], hand[i+1:]...)

		// If there will be at least one non-effect card available if this one is played...
		if !containsOnlyEffectCards(cardSetWithoutCurrent) {

			// And it the card to be played is NOT a blast card (blasts are always invalid as they do not change
			// the score)...
			if hand[i] != Blast {

				// If the value of the card causes the player's score to beat or match (cardToBeatOrMatch), it's valid.
				if currentScore+hand[i].Value() >= cardToBeatOrMatch.Value() {
					return true
				}

				// If the card is a force card and causes the player's score to beat or match (cardToBeatOrMatch),
				// it's valid.
				if hand[i] == Force {
					if currentScore*2 >= cardToBeatOrMatch.Value() {
						return true
					}
				}

				// Bolts and mirrors are always valid first cards, as they will always result in the turn being changed.
				if hand[i] == Bolt || hand[i] == Mirror {
					return true
				}
			}
		}
	}

	// Reaching this point indicates that the specified hand did not contain a card that could be played to put
	// the game in a playable state.
	return false
}
