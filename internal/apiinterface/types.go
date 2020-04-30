package apiinterface

// Winner represents the winner for a match
type Winner uint8

// Enum values that represent the potential winners for a match
const (
	Draw    Winner = 0
	Player1        = 1
	Player2        = 2
)

// MMRUpdateRequest describes the data needed to update the MMR for a pair of users after a match
type MMRUpdateRequest struct {
	Player1ID uint64
	Player2ID uint64
	Winner    Winner
}
