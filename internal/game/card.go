// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game implements the Blade II Online game server.
package game

// Card is a typedef for different Blade II card types.
type Card uint8

// Card enums - explicitly numbered so that their numerical value can be
// seen when hovering them in vscode.
const (

	// Standard cards (basic).
	ElliotsOrbalStaff   Card = 0
	FiesTwinGunswords   Card = 1
	AlisasOrbalBow      Card = 2
	JusisSword          Card = 3
	MachiasOrbalShotgun Card = 4
	GaiusSpear          Card = 5
	LaurasGreatsword    Card = 6

	// Standard cards (effects).
	Bolt   Card = 7
	Mirror Card = 8
	Blast  Card = 9
	Force  Card = 10

	// Bolted Cards (basic).
	InactiveElliotsOrbalStaff   Card = 11
	InactiveFiesTwinGunswords   Card = 12
	InactiveAlisasOrbalBow      Card = 13
	InactiveJusisSword          Card = 14
	InactiveMachiasOrbalShotgun Card = 15
	InactiveGaiusSpear          Card = 16
	InactiveLaurasGreatsword    Card = 17

	// Bolted cards (effects).
	InactiveBolt   Card = 18
	InactiveMirror Card = 19
	InactiveBlast  Card = 20
	InactiveForce  Card = 21
)

const (

	// cardEnumToValueOffset is used to determine the score value of a Card enum.
	cardEnumToValueOffset = 1

	// effectCardDefaultValue is the default score value for all effect cards, when they are played as a non effect card
	// (i.e. after drawing from the deck onto the field).
	effectCardDefaultValue = 1
)

// Value returns the point value of the specified card, if it where to be played on the field.
// Effect cards always return 1.
func (c Card) Value() (value uint8) {

	// Determine the score value based on the enum value of the card.
	if c < Bolt {

		// If this is less than a bolt card (i.e. it's a standard basic card), cast it to
		// a uint8 so that we can add (cardEnumToValueOffset) to it
		value = uint8(c) + cardEnumToValueOffset

	} else if c >= Bolt && c <= Force {

		// If this is a standard effect card, it's value is always one.
		value = effectCardDefaultValue
	} else {

		// If this card was not a standard basic or effect card, it should have a value of zero
		// indicating that it is a bolted card that adds no points to a player's score, OR
		// the card that this function was called on is invalid
		value = 0
	}

	return value
}
