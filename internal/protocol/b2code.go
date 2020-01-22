package protocol

// B2Code is a typedef for non-system websocket messages
type B2Code uint16

// Offsets for various types of message
const (
	connectionOffset  B2Code = 100
	authOffset               = 200
	matchMakingOffset        = 300
)

// WSCInfo is a generic all-purpose code. Try to avoid using this unless the message can be safely ignored
const WSCInfo = 0

// Connection
const (
	WSCConnectionTimeOut = iota + connectionOffset
	WSCUnknownError
)

// Auth
const (
	WSCAuthRequest = iota + authOffset
	WSCAuthBadFormat
	WSCAuthBadCredentials
	WSCAuthExpired
	WSCAuthBanned
	WSCAuthExpected
	WSCAuthNotReceived
	WSCAuthSuccess
)

// MatchMaking
const (
	WSCMatchMakingGameFound = iota + matchMakingOffset
)
