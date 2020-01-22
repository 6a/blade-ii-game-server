package connection

// WSCode is a typedef for non-system websocket messages
type WSCode uint16

// Offsets for various types of message
const (
	connectionOffset  = 100
	authOffset        = 200
	matchMakingOffset = 300
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
