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

// State enums
const (
	WaitingForPlayers Phase = 0
	Play              Phase = 1
	Finished          Phase = 2
)

// MatchState represents the current state of a match
type MatchState struct {
	Winner Player
	Turn   Player
	Cards  Cards
	Phase  Phase
}
