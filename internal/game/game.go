package game

// Match is a wrapper for a matche's data and client connections etc
type Match struct {
	MatchID uint64
	Client1 *GClient
	Client2 *GClient
	State   State
}
