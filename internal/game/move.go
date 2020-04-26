package game

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var validMoveStringRegex = regexp.MustCompile("^[^:]+:[^:]*$")

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

	if !validMoveStringRegex.MatchString(stringMove) {
		return move, errors.New("Serialised move format invalid")
	}

	data := strings.Split(stringMove, payloadDelimiter)

	outInt, err := strconv.Atoi(data[0])
	if err != nil {
		return move, errors.New("Could not parse the code for the incoming move (wrong type)")
	}

	if outInt < 0 || outInt > 17 {
		return move, errors.New("Could not parse the code for the incoming move (not valid b2serverupdate)")
	}

	move.Instruction = B2MatchInstruction(outInt)

	if len(data) == 2 {
		move.Payload = data[1]
	} else {
		move.Payload = ""
	}

	return move, nil
}
