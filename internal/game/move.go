package game

import (
	"errors"
	"strconv"
	"strings"
)

// Move represents a client match data packet
type Move struct {
	Instruction B2MatchInstruction
	Payload     string
	Original    string
}

// MoveFromString attempts to parse a move for the specified string
// Non nil error means something went wrong
func MoveFromString(stringMove string) (move Move, err error) {
	move = Move{
		Original: stringMove,
	}

	data := strings.Split(stringMove, payloadDelimiter)

	if len(data) != 2 {
		return move, errors.New("Could not parse incoming move")
	}

	outInt, err := strconv.Atoi(data[1])
	if err != nil {
		return move, errors.New("Could not parse the code for the incoming move (wrong type)")
	}

	if outInt < 0 || outInt > 17 {
		return move, errors.New("Could not parse the code for the incoming move (not valid b2serverupdate)")
	}

	move.Instruction = B2MatchInstruction(outInt)
	move.Payload = data[1]

	return move, nil
}
