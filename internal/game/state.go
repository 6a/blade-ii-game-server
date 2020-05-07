// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game implements the Blade II Online game server.
package game

// Player is a typedef for the two different players
type Player uint8

// Player enums representing each type of player. PlayerUndecided is an edge case for indicating
// that there is no player specified, such as when the game just cleared all the cards from the
// field and is waiting to see when each player draw onto the field.
const (
	PlayerUndecided Player = 0
	Player1         Player = 1
	Player2         Player = 2
)

// Phase is a typedef for the different states that a match can be in.
type Phase uint8

// Match phase enums.
const (
	WaitingForPlayers Phase = 0
	Play              Phase = 1
	Finished          Phase = 2
)

// MatchState represents the current state of a match.
type MatchState struct {

	// Winner is the database ID for the winner. A value of zero indicates that there is no player,
	// or the game ended in a draw.
	Winner uint64

	// Whos turn it currently is.
	Turn Player

	// The cards for this match.
	Cards Cards

	// Player scores.
	Player1Score uint16
	Player2Score uint16

	// Match phase (used by the server to determine whether to, for example, tick the match or not).
	Phase Phase
}
