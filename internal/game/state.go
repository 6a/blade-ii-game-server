package game

// Player is a typedef for the two different players
type Player uint8

// Player enums
const (
	PlayerUndecided Player = 0
	Player1         Player = 1
	Player2         Player = 2
)

// State represents the current state of a match
type State struct {
	Winner   Player
	Turn     Player
	Cards    Cards
	Finished bool
}
