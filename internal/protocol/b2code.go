package protocol

// B2Code is a typedef for non-system websocket messages
type B2Code uint16

// WSCNone is a generic all-purpose code. Try to avoid using this unless the message can be safely ignored
const WSCNone B2Code = 0

// Connection
const (
	WSCConnectionTimeOut      B2Code = 100
	WSCUnknownConnectionError B2Code = 101
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
	WSCAuthReceived       B2Code = 207
	WSCAuthSuccess        B2Code = 208
)

// MatchMaking
const (
	WSCMatchMakingGameFound B2Code = 300
	WSCMatchMakingAccept    B2Code = 301
	WSCMatchConfirmed       B2Code = 302
	WSCReadyCheckFailed     B2Code = 303
	WSCJoinedQueue          B2Code = 304
	WSCOpponentAccepted     B2Code = 305
	WSCOpponentDidNotAccept B2Code = 306
)

// Match
const (
	WSCMatchID                  B2Code = 400
	WSCMatchIDExpected          B2Code = 401
	WSCMatchIDBadFormat         B2Code = 402
	WSCMatchInvalid             B2Code = 403
	WSCMatchExpired             B2Code = 404
	WSCMatchIDReceived          B2Code = 405
	WSCMatchIDNotReceived       B2Code = 406
	WSCMatchIDConfirmed         B2Code = 407
	WSCMatchMultipleConnections B2Code = 408
	WSCMatchFull                B2Code = 409
	WSCMatchJoined              B2Code = 410
	WSCMatchIllegalMove         B2Code = 411
	WSCMatchRelayMessage        B2Code = 412
	WSCMatchMove                B2Code = 413
	WSCMatchData                B2Code = 414
	WSCMatchForfeit             B2Code = 415
	WSCMatchMutualTimeout       B2Code = 416
	WSCMatchTimeOut             B2Code = 417
	WSCMatchWin                 B2Code = 418
	WSCMatchDraw                B2Code = 419
	WSCMatchLoss                B2Code = 420
)
