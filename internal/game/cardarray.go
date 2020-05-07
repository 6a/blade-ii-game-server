// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game provides implements the Blade II Online game server.
package game

// reverseCardArray reverses a Card array in place. This modified the underlying
// data for the specified array, so two variables pointing to the same data
// would, for example, have their data modified.
func reverseCardArray(inArray []Card) {

	// Method described here: https://stackoverflow.com/a/19239850.
	// Moves from the the outermost members (0th and len()th) towards the center, swapping them simultaneously
	// until the indices overlap.
	//
	// Example - For an array [A, B, C]:
	// 	Iteration 0:
	//		inArray[i] = A
	// 		inArray[j] = C
	//
	//		Perform swap:
	//			inArray = [A, B, C]
	//			    		\   /
	//			    		 \ /
	//			    		  X
	//			    		 / \
	//			    		/   \
	//			inArray = [C, B, A]
	//		inArray[i] = C
	// 		inArray[j] = A
	// 	Iteration 1:
	//		Exits (as i is now NOT less than j)

	for i, j := 0, len(inArray)-1; i < j; i, j = i+1, j-1 {

		// Using go's multiple return values we can swap values without using temporary variables!
		inArray[i], inArray[j] = inArray[j], inArray[i]
	}
}

// removeFirstOfType removes the first card of the specified type, from the specified
// card array. The specified array is passed in as a pointer, and thus the array that
// is being pointed to, is the one that will be modified. Returns true unless it was
// not possible to perform the removal, such as when an instance of the specified card
// was not present in the specified array.
//
// [ WARNING ] THIS DOES NOT PRESERVE THE ORDER OF THE CARD ARRAY - DO NOT USE ON THE
// DECK.
func removeFirstOfType(cards *[]Card, toRemove Card) (success bool) {

	// Default value - if no matching card is found in the specified array, this will
	// remain unchanged.
	var indexToRemove = -1

	// Iterate over all the cards in the array. Assign the array index to (indexToRemove)
	// if card at the current index matches the type of (toRemove), and break out of the
	// loop.
	for i := 0; i < len(*cards); i++ {
		if (*cards)[i] == toRemove {
			indexToRemove = i
			break
		}
	}

	// If the parameters were valid (a card of the specified type is present in the
	// specified card array) remove the card at (indexToRemove).
	// removing the card in this fashion does NOT preserve the order of the array.
	if indexToRemove != -1 {

		// Copy the value of the last card in the array, into the element at the index of
		// the card that should be "removed".
		(*cards)[indexToRemove] = last(*cards)

		// Remove the last card in the array, as this card was copied to the element mentioned above.
		removeLast(cards)

		// In summary this overwrites the card at the index to remove with the card
		// at the end of the array, and then removes the last card.
	}

	// Return a boolean based on whether (indexToRemove) was modified after initialisation (which
	// implies that a matching card was found).
	return indexToRemove != -1
}

// removeFirstOfType removes the last card from the specified card array. The specified array is
// passed in as a pointer, and thus the array that is being pointed to, is the one that will be
// modified. Returns true, unless the specified array was empty or nil.
//
// Note that the removed array member still exists in memory, though there is no guarantee how
// long it will remain if there are no other references to it.
func removeLast(cards *[]Card) bool {

	// Check to ensure that the array is not empty or nil
	if len(*cards) > 0 {

		// Reduce the size of the array by 1, essentially trimming off the last member.
		*cards = (*cards)[:len(*cards)-1]
		return true
	}

	// Reaching this point means that the array was empty or nil.
	return false
}

// containsOnlyEffectCards returns true if all the cards in thethe specified array of cards
// contains are ALL effect cards.
//
// [ WARNING ] DOES NOT PROPERLY HANDLE INACTIVE CARDS - USE THIS ONLY ON A PLAYERS HAND.
func containsOnlyEffectCards(cardSet []Card) bool {

	// Loop over all the cards in the specified array, and early exit with false if
	// a standard basic card is found.
	for i := 0; i < len(cardSet); i++ {
		if cardSet[i] < Bolt {
			return false
		}
	}

	// Reaching this point means that all the cards were effect cards.
	return true
}

// canOvercomeDifference returns true if the specified set of cards contains a card that, if
// added to the field, will increase the score by enough to either beat or match (difference).
// This does not handle effect edge case, which should be checked for separately.
func canOvercomeDifference(cardSet []Card, difference uint16) bool {

	// Loop over all the cards, and early exit with true if one of them can beat or match
	// (difference).
	for i := 0; i < len(cardSet); i++ {

		// The card value is casted to a uint16 so that the value check is valid.
		if uint16(cardSet[i].Value()) >= difference {
			return true
		}
	}

	// Reaching this point means that none of the cards would beat or match (difference).
	return false
}

// canOvercomeDifference returns true if the specified card array contains at least one
// instance of the specified card.
func contains(cardSet []Card, cardToCheck Card) bool {

	// Loop over all the cards, and early exit with true if one of them is the same type
	// as the specified card.
	for i := 0; i < len(cardSet); i++ {
		if cardSet[i] == cardToCheck {
			return true
		}
	}

	// Reaching this point means that none of the cards were of the same type as the
	// specified card.
	return false
}

// last returns a copy of the last card in a card array. Returns the default value of
// "ElliotsOrbalStaff" if the card array is nil or empty.
func last(cardSet []Card) Card {

	// Check to ensure that the card array is not nil or empty.
	if len(cardSet) > 0 {

		// Returns a copy of the last card.
		return cardSet[len(cardSet)-1]
	}

	// Reaching this point means that the array was empty or nil.
	return ElliotsOrbalStaff
}

// bolt "bolts" the last card of the specified card array, replacing the array member with
// an inactive version of the same card. The specified array is passed in as a pointer, and
// thus the array that is being pointed to, is the one that will be modified.
//
// Results in a noop if the last card in the specified card array is inactive (already bolted).
func bolt(targetField *[]Card) {

	// Check to ensure that the card array is not nil or empty.
	if len(*targetField) > 0 {

		// Check to ensure that the last card in the card array is not already bolted.
		if last(*targetField) <= Force {

			// Replace the last member of the card array that is pointed to, with the inactive
			// (bolted) equivalent of the original card. The original card is first cast to a
			// uint8, increased by the bolted card offset value, and then cast back to a Card.
			(*targetField)[len(*targetField)-1] = Card(uint8(last(*targetField)) + boltedCardOffset)
		}
	}
}

// unBolt "un-bolts" the last card of the specified card array, replacing the array member with
// a standard version of the same card. The specified array is passed in as a pointer, and
// thus the array that is being pointed to, is the one that will be modified.
//
// Results in a noop if the last card in the specified card array is active (not currently bolted).
func unBolt(targetField *[]Card) {

	// Check to ensure that the card array is not nil or empty.
	if len(*targetField) > 0 {

		// Check to ensure that the last card in the card array is not already active.
		if last(*targetField) >= InactiveElliotsOrbalStaff {

			// Replace the last member of the card array that is pointed to, with the active
			// (unbolted) equivalent of the original card. The original card is first cast to a
			// uint8, decreasd by the bolted card offset value, and then cast back to a Card.
			(*targetField)[len(*targetField)-1] = Card(uint8(last(*targetField)) - boltedCardOffset)
		}
	}
}

// calculateScore aggregates the values of all the cards in the specified card array, taking
// into consideration the edge case where a force card doubles the score of all the previous
// cards in the array.
func calculateScore(targetCards []Card) uint16 {

	// Start with a default value of zero of type uint16. I'm pretty sure that the total score
	// Can never exceed the max uint8, but just incase, a uint16 is used.
	var total uint16 = 0

	// For each of the cards in the card array...
	for i := 0; i < len(targetCards); i++ {

		// Check if the card is bolted, as bolted cards do not add any points to the score.
		if !isBolted(targetCards[i]) {

			// If the current card is a Force card, and it is NOT the first card in the array, double
			// the current total. If the card WAS the first card in the array, it is handled as a normal
			// card as it could only have come from the deck straight onto the field. Otherwise, the card
			// is handled like a normal card, and its value is added to the total.
			if targetCards[i] == Force && i > 0 {
				total *= 2
			} else {
				total += uint16(targetCards[i].Value())
			}
		}
	}

	return total
}
