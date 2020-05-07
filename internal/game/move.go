// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game implements the Blade II Online game server.
package game

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// Regex to determine if a move string is valid.
var validMoveStringRegex = regexp.MustCompile("^[^:]+:[^:]*$")

// Move represents a client match data packet.
type Move struct {
	Instruction B2MatchInstruction
	Payload     string
}

// MoveFromString attempts to parse a move from the specified move string
// Non nil error means something went wrong.
func MoveFromString(moveString string) (move Move, err error) {

	// Create a new move
	move = Move{}

	// Check if the move string is valid, using the validation regex. If not, return an error.
	if !validMoveStringRegex.MatchString(moveString) {
		return move, errors.New("Serialised move format invalid")
	}

	// Attempt to split the move string using the payload delimiter, storing each part as
	// a string in an array.
	data := strings.Split(moveString, payloadDelimiter)

	// Attempt to parse the first value in the data array. This is the instruction code for the move.
	// A failure returns an error.
	outInt, err := strconv.Atoi(data[0])
	if err != nil {
		return move, errors.New("Could not parse the code for the incoming move (wrong type)")
	}

	// Ensure that it's a valid move update.
	if outInt < 0 || outInt > int(CardForce) {
		return move, errors.New("Could not parse the code for the incoming move (not valid b2serverupdate)")
	}

	// Cast the int value to an instruction code.
	move.Instruction = B2MatchInstruction(outInt)

	// If there is a second member in the array, that means that there is payload data. This should be
	// store as the Payload member of the move. Otherwise, the payload will remain as an empty string.
	if len(data) == 2 {
		move.Payload = data[1]
	}

	return move, nil
}
