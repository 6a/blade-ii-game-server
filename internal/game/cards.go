package game

import (
	"bytes"
	"math"
	"math/rand"
	"strconv"
)

const delimiter string = "."
const maxDrawsOnStart uint8 = 3
const postInitialisationDeckSize uint8 = 4

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

// Serialize returns the string representation of the DECKS ONLY, as the other data is never required to be sent
func (c *Cards) Serialize() string {

	var buffer bytes.Buffer

	// Player 1's deck is added
	for _, card := range c.Player1Deck {
		buffer.WriteString(strconv.FormatUint(uint64(card), 16))
	}

	// Another delimiter
	buffer.WriteString(delimiter)

	// Player 2's deck is added
	for _, card := range c.Player2Deck {
		buffer.WriteString(strconv.FormatUint(uint64(card), 16))
	}

	// Resultant string read from buffer
	return buffer.String()
}

// Generate generates a new set of cards for a match - has additional checks to ensure that the match is not unwinnable from the first move etc
func Generate() (cards Cards) {

	// Add all the cards (ref: https://www.reddit.com/r/Falcom/comments/fxt5nq/can_i_buy_the_card_game_blade_anywhere/fmxo8qo/)
	pool := []Card{
		Bolt, Bolt,
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

		// Fill player 1 deck using the permutation
		for i := 0; i < 15; i++ {
			cards.Player1Deck = append(cards.Player1Deck, pool[permutation[i]])
		}

		// Fill player 2 deck using the permutation
		for i := 15; i < 30; i++ {
			cards.Player2Deck = append(cards.Player1Deck, pool[permutation[i]])
		}

		// Check the cards validity - a result of true will cause the loop to exit
		success = validateCards(&cards)
	}

	return cards
}

// Initialise simulates the first moves of the game until a playable state is reached
func Initialise(inCards Cards) (outCards Cards) {

	return outCards
}

// validateCards returns true if the current cards will NOT result in a bad game state, such as an insta-loss, or more than "maxDrawsOnStart" draws on the first turn
func validateCards(cards *Cards) bool {
	for i := uint8(0); i < maxDrawsOnStart; i++ {
		cardIndex := postInitialisationDeckSize - i

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
			var cardSet *[]Card
			var opponentCard Card
			var scoreToCheck uint8

			if player1Score < player2Score {
				cardSet = &cards.Player1Hand
				scoreToCheck = player1Score
			} else {
				cardSet = &cards.Player2Hand
			}

			// If, after the first legal draw, the player that goes first has a valid move to play, the cards are valid
			if validFirstMoveAvailable(*cardSet, opponentCard, scoreToCheck) {
				return true
			}

			// If we get to this point, the first draw was legal, but the player that goes first was immediately put into a position where they
			// could not continue (aside from playing blast cards, but these are ignored)
			break
		}
	}

	return false
}

// validFirstMoveAvailable returns true if the specified set of cards contains one that can beat the specified card (first move only, i.e. after initial draw from deck)
func validFirstMoveAvailable(cardSet []Card, card Card, currentScore uint8) bool {
	if len(cardSet) == 1 {
		return true
	}

	var cardSetWithoutCurrent []Card
	for i := 0; i < len(cardSet); i++ {
		cardSetWithoutCurrent = append(cardSet[:i], cardSet[i+1:]...)

		// If there will be at least one non-effect card available if this one is played
		if !containsOnlyEffectCards(cardSetWithoutCurrent) {

			// If playing the card will beat the score, its valid
			// Note blast cards are skipped as they dont change the score
			// Note force cards are further checked to ensure that the resultant score is greater than or equal to the specified card
			// Checking like this works for mirro and bolt as they change the score
			if cardSet[i].Value() >= card.Value() {
				if cardSet[i] != Blast {
					if cardSet[i] == Force {
						if currentScore*2 >= card.Value() {
							return true
						}
					}

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
