package queue

// ClientPair is a light wrapper for a pair of client connections
type ClientPair struct {
	C1 *Client
	C2 *Client
}

// NewPair creates a new ClientPair
func NewPair(c1 *Client, c2 *Client) ClientPair {
	return ClientPair{
		C1: c1,
		C2: c2,
	}
}
