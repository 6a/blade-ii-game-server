// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game provides implements the Blade II Online game server.
package game

// Player is a typedef for the two different players
type Player uint8

// Player enums
const (
	PlayerUndecided Player = 0
	Player1         Player = 1
	Player2         Player = 2
)

// Phase is a typedef for the different states that a match can be in
type Phase uint8

// Match phase enums
const (
	WaitingForPlayers Phase = 0
	Play              Phase = 1
	Finished          Phase = 2
)

// MatchState represents the current state of a match
type MatchState struct {
	Winner       uint64
	Turn         Player
	Cards        Cards
	Player1Score uint16
	Player2Score uint16
	Phase        Phase
}
