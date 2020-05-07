// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game implements the Blade II Online game server.
package game

const (

	// serverMoveUpdateToCardOffset is used to convert a server message that represents a card update, into a Card enum, and vice versa.
	serverMoveUpdateToCardOffset uint8 = 1

	// serverMoveUpdateMin is the minimum value a server message that represents a card update can have.
	serverMoveUpdateMin uint8 = 1

	// serverMoveUpdateMax is the maximum value a server message that represents a card update can have.
	serverMoveUpdateMax uint8 = 11
)

// B2MatchInstruction is an enum type that represents different different types of match instruction for Blade II Online.
type B2MatchInstruction uint8

// Definitions of server updates. Explicitly numbered so that their numerical value can be
// seen when hovering them in vscode.
const (

	// None is a special case used to indicate that this instruction should result in a noop.
	None B2MatchInstruction = 0

	// Instructions 1 through 11 indicate that a card was selected. The position from which it
	// should be taken, and where it should end up, is determined by the current board state.
	CardElliotsOrbalStaff   B2MatchInstruction = 1
	CardFiesTwinGunswords   B2MatchInstruction = 2
	CardAlisasOrbalBow      B2MatchInstruction = 3
	CardJusisSword          B2MatchInstruction = 4
	CardMachiasOrbalShotgun B2MatchInstruction = 5
	CardGaiusSpear          B2MatchInstruction = 6
	CardLaurasGreatsword    B2MatchInstruction = 7
	CardBolt                B2MatchInstruction = 8
	CardMirror              B2MatchInstruction = 9
	CardBlast               B2MatchInstruction = 10
	CardForce               B2MatchInstruction = 11

	// Messages that can be sent to and from the server.
	InstructionForfeit B2MatchInstruction = 12
	InstructionMessage B2MatchInstruction = 13

	// Messages that can only be received from the server.
	InstructionCards              B2MatchInstruction = 14
	InstructionPlayerData         B2MatchInstruction = 15
	InstructionOpponentData       B2MatchInstruction = 16
	InstructionConnectionProgress B2MatchInstruction = 17
	InstructionConnectionClosed   B2MatchInstruction = 18

	// Error messages from the server grouped so we can check for errors by equality (> the lowest value error).
	InstructionConnectionError    B2MatchInstruction = 19
	InstructionAuthError          B2MatchInstruction = 20
	InstructionMatchCheckError    B2MatchInstruction = 21
	InstructionMatchSetupError    B2MatchInstruction = 22
	InstructionMatchIllegalMove   B2MatchInstruction = 23
	InstructionMatchMutualTimeOut B2MatchInstruction = 24
	InstructionMatchTimeOut       B2MatchInstruction = 25
)

// ToCard returns this instruction as a card. Invalid cards are returned with the default value of 0 (ElliotsOrbalStaff).
func (i B2MatchInstruction) ToCard() (card Card) {

	// Set the return value to the default value, ensuring a valid value is returned even if the instruction was not valid.
	card = ElliotsOrbalStaff

	// if the incoming instruction is within the valid range for a move update, convert it to a Card enum.
	if uint8(i) >= serverMoveUpdateMin && uint8(i) <= serverMoveUpdateMax {

		// Conversion is performed by subtracting the move update to card offset after casting to a uint8 (to allow
		// the subtraction to be valid) and then casting the result to a Card enum.
		card = Card(uint8(i) - serverMoveUpdateToCardOffset)
	}

	return card
}
