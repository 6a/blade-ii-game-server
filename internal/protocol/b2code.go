package protocol

// B2Code is a typedef for non-system websocket messages
type B2Code uint16

// WSCInfo is a generic all-purpose code. Try to avoid using this unless the message can be safely ignored
const WSCInfo = 0

// Connection
const (
	WSCConnectionTimeOut B2Code = 100
	WSCUnknownError      B2Code = 101
)

// Auth
const (
	WSCAuthRequest        B2Code = 200
	WSCAuthBadFormat      B2Code = 201
	WSCAuthBadCredentials B2Code = 202
	WSCAuthExpired        B2Code = 203
	WSCAuthBanned         B2Code = 204
	WSCAuthExpected       B2Code = 205
	WSCAuthNotReceived    B2Code = 206
	WSCAuthSuccess        B2Code = 207
)

// MatchMaking
const (
	WSCMatchMakingGameFound B2Code = 300
	WSCMatchMakingReady     B2Code = 301
	WSCMatchConfirmed       B2Code = 302
	WSCReadyCheckFailed     B2Code = 303
	WSCJoinedQueue          B2Code = 304
)

// Match
const (
	WSCMatchID                  B2Code = 400
	WSCMatchIDExpected          B2Code = 401
	WSCMatchIDBadFormat         B2Code = 402
	WSCMatchInvalid             B2Code = 403
	WSCMatchIDNotReceived       B2Code = 404
	WSCMatchMultipleConnections B2Code = 405
	WSCMatchFull                B2Code = 406
	WSCMatchIllegalMove         B2Code = 407
	WSCMatchJoined              B2Code = 408
	WSCMatchData                B2Code = 409
	WSCMatchOpponentData        B2Code = 410
	WSCMatchMoveUpdate          B2Code = 411
	WSCMatchForfeit             B2Code = 412
	WSCMatchOpponentForfeit     B2Code = 413
	WSCMatchMessage             B2Code = 414
)
