package game

import (
	"bytes"
	"math"
	"math/rand"
	"strconv"
	"time"
)

// SerializedCardsDelimiter is the delimiter for serialized cards objects
const SerializedCardsDelimiter string = "."
const maxDrawsOnStart uint8 = 3
const postInitialisationDeckSize uint8 = 5
const startingHandSize uint8 = 10

// Cards is a container for all the cards on the field
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

// Serialized returns the string representation of the DECKS ONLY, as the other data is never required to be sent
func (c *Cards) Serialized() string {

	var buffer bytes.Buffer

	// Player 1's deck is added
	for _, card := range c.Player1Deck {
		buffer.WriteString(strconv.FormatUint(uint64(card), 16))
	}

	// Another delimiter
	buffer.WriteString(SerializedCardsDelimiter)

	// Player 2's deck is added
	for _, card := range c.Player2Deck {
		buffer.WriteString(strconv.FormatUint(uint64(card), 16))
	}

	// Resultant string read from buffer
	return buffer.String()
}

// GenerateCards generates a new set of cards for a match - has additional checks to ensure that the match is not unwinnable from the first move etc
func GenerateCards() (cards Cards, drawsUntilValid uint) {

	// This probably doenst need to be done every time so..
	// TODO determine if this is better off being called just once
	rand.Seed(time.Now().UTC().UnixNano())

	// Add all the cards (ref: https://www.reddit.com/r/Falcom/comments/fxt5nq/can_i_buy_the_card_game_blade_anywhere/fmxo8qo/)
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

	var success = false
	for !success {
		permutation := rand.Perm(len(pool))
		cards = Cards{}

		// Fill player 1 deck using the permutation
		for i := 0; i < 15; i++ {
			cards.Player1Deck = append(cards.Player1Deck, pool[permutation[i]])
		}

		// Fill player 2 deck using the permutation
		for i := 15; i < 30; i++ {
			cards.Player2Deck = append(cards.Player2Deck, pool[permutation[i]])
		}

		// Check the cards validity - a result of true will cause the loop to exit
		success, drawsUntilValid = validateCards(&cards)
	}

	return cards, drawsUntilValid
}

// InitialiseCards simulates the first moves of the game until a playable state is reached - dont call this on a deck before its validated
func InitialiseCards(inCards Cards, drawsUntilValid uint) (outCards Cards) {

	// Copy into out value
	outCards = inCards.Copy()

	// Player 1 deck to player 1 hand

	outCards.Player1Hand = outCards.Player1Deck[postInitialisationDeckSize:]
	reverseCardArray(outCards.Player1Hand)
	outCards.Player1Deck = outCards.Player1Deck[:postInitialisationDeckSize]

	// Player 2 deck to player 2 hand
	outCards.Player2Hand = outCards.Player2Deck[postInitialisationDeckSize:]
	reverseCardArray(outCards.Player2Hand)
	outCards.Player2Deck = outCards.Player2Deck[:postInitialisationDeckSize]

	// if drawsUntilValid > 1 {
	// 	// Player 1 deck to player 1 discard
	// 	p1Index := uint(len(outCards.Player1Deck)) - (drawsUntilValid - 2) - 1
	// 	outCards.Player1Discard = outCards.Player1Deck[p1Index:]
	// 	outCards.Player1Deck = outCards.Player1Deck[:p1Index]

	// 	// Player 2 deck to player 2 discard
	// 	p2Index := uint(len(outCards.Player2Deck)) - (drawsUntilValid - 2) - 1
	// 	outCards.Player2Discard = outCards.Player2Deck[p2Index:]
	// 	outCards.Player2Deck = outCards.Player2Deck[:p2Index]
	// }

	// // Player 1 deck to player 1 field
	// outCards.Player1Field = outCards.Player1Deck[len(outCards.Player1Deck)-1:]
	// outCards.Player1Deck = outCards.Player1Deck[:len(outCards.Player1Deck)-1]

	// // Player 2 deck to player 2 field
	// outCards.Player2Field = outCards.Player2Deck[len(outCards.Player2Deck)-1:]
	// outCards.Player2Deck = outCards.Player2Deck[:len(outCards.Player2Deck)-1]

	return outCards
}

// Copy returns a deep copy of this cards object
func (c Cards) Copy() (outCards Cards) {
	// Size all the fields
	outCards.Player1Deck = make([]Card, len(c.Player1Deck))
	outCards.Player1Hand = make([]Card, len(c.Player1Hand))
	outCards.Player1Field = make([]Card, len(c.Player1Field))
	outCards.Player1Discard = make([]Card, len(c.Player1Discard))
	outCards.Player2Deck = make([]Card, len(c.Player2Deck))
	outCards.Player2Hand = make([]Card, len(c.Player2Hand))
	outCards.Player2Field = make([]Card, len(c.Player2Field))
	outCards.Player2Discard = make([]Card, len(c.Player2Discard))

	// Player 1 cards
	copy(outCards.Player1Deck, c.Player1Deck)
	copy(outCards.Player1Hand, c.Player1Hand)
	copy(outCards.Player1Field, c.Player1Field)
	copy(outCards.Player1Discard, c.Player1Discard)

	// Player 2 cards
	copy(outCards.Player2Deck, c.Player2Deck)
	copy(outCards.Player2Hand, c.Player2Hand)
	copy(outCards.Player2Field, c.Player2Field)
	copy(outCards.Player2Discard, c.Player2Discard)

	return outCards
}

// validateCards returns true if the current cards will NOT result in a bad game state, such as an insta-loss, or more than "maxDrawsOnStart" draws on the first turn
func validateCards(cards *Cards) (valid bool, drawsUntilValid uint) {
	for i := uint8(0); i < maxDrawsOnStart; i++ {
		cardIndex := postInitialisationDeckSize - 1 - i

		// Just incase... This also means that a valid move was not found within the first [ maxDrawsOnStart ]
		if cardIndex < 0 {
			break
		}

		// Ensure that play starts within "maxDrawsOnStart" draws - a non zero value indicates that the scores will be different
		player1Score := cards.Player1Deck[cardIndex].Value()
		player2Score := cards.Player2Deck[cardIndex].Value()
		scoreDifference := math.Abs(float64(int16(player1Score) - int16(player2Score)))
		if scoreDifference != 0 {

			// Also ensure that the score difference from the opponent hand is beatable by the player that goes first (if their
			// hand does not have any cards of high enough value they will insta-lose otherwise)
			var cardSet = make([]Card, startingHandSize)
			var cardToBeatOrMatch Card
			var scoreToCheck uint8

			// Copy the cards that we will be examining. Note the explicit copy, as we modify the array later which would
			// otherwise effect the original set of cards
			if player1Score < player2Score {
				copy(cardSet, cards.Player1Deck[postInitialisationDeckSize:])
				scoreToCheck = player1Score
				cardToBeatOrMatch = cards.Player2Deck[cardIndex]
			} else {
				copy(cardSet, cards.Player2Deck[postInitialisationDeckSize:])
				scoreToCheck = player2Score
				cardToBeatOrMatch = cards.Player1Deck[cardIndex]
			}

			// Reverse the hand as we would in fact be drawing from the top of the deck to the start of the hand
			reverseCardArray(cardSet)

			// If, after the first legal draw, the player that goes first has a valid move to play, the cards are valid
			return validFirstMoveAvailable(cardSet, cardToBeatOrMatch, scoreToCheck), (uint(i) + 1)
		}
	}

	return false, 0
}

// validFirstMoveAvailable returns true if the specified set of cards contains one that can beat the specified card (first move only, i.e. after initial draw from deck)
func validFirstMoveAvailable(cardSet []Card, cardToBeatOrMatch Card, currentScore uint8) bool {
	if len(cardSet) == 1 {
		return true
	}

	var cardSetWithoutCurrent []Card
	for i := 0; i < len(cardSet); i++ {
		cardSetWithoutCurrent = append(cardSet[:i], cardSet[i+1:]...)

		// If there will be at least one non-effect card available if this one is played
		if !containsOnlyEffectCards(cardSetWithoutCurrent) {

			// Blasts are always invalid as they do not change the score
			if cardSet[i] != Blast {

				// if the card causes the score to beat or match, its valid
				if currentScore+cardSet[i].Value() >= cardToBeatOrMatch.Value() {
					return true
				}

				// If the force effect causes the score to beat or match, its valid
				if cardSet[i] == Force {
					if currentScore*2 >= cardToBeatOrMatch.Value() {
						return true
					}
				}

				// Bolts and mirrors are always valid
				if cardSet[i] == Bolt || cardSet[i] == Mirror {
					return true
				}
			}
		}
	}

	return false
}

func containsOnlyEffectCards(cardSet []Card) bool {
	for i := 0; i < len(cardSet); i++ {
		if cardSet[i] < Bolt {
			return false
		}
	}

	return true
}
