package game

// B2MatchInstruction is an enum type that represents different different types of match instruction for Blade II Online
type B2MatchInstruction uint8

// Definitions of server updates
const (
	None                    B2MatchInstruction = 0
	CardElliotsOrbalStaff   B2MatchInstruction = 1
	CardFiesTwinGunswords   B2MatchInstruction = 2
	CardAlisasOrbalBow      B2MatchInstruction = 3
	CardJusisSword          B2MatchInstruction = 4
	CardMachiasOrbalShotgun B2MatchInstruction = 5
	CardGaiusSpear          B2MatchInstruction = 6
	CardLaurasGreatsword    B2MatchInstruction = 7
	CardBolt                B2MatchInstruction = 8
	CardMirror              B2MatchInstruction = 9
	CardBlast               B2MatchInstruction = 10
	CardForce               B2MatchInstruction = 11

	// Messages that can be sent to and from the serve
	InstructionForfeit B2MatchInstruction = 12
	InstructionMessage B2MatchInstruction = 13

	// Messages that can only be received from the serve
	InstructionCards              B2MatchInstruction = 14
	InstructionPlayerData         B2MatchInstruction = 15
	InstructionOpponentData       B2MatchInstruction = 16
	InstructionConnectionProgress B2MatchInstruction = 17
	InstructionConnectionClosed   B2MatchInstruction = 18

	// Error messages from the server grouped so we can check for errors by equality (> the lowest value error)
	InstructionConnectionError  B2MatchInstruction = 19
	InstructionAuthError        B2MatchInstruction = 20
	InstructionMatchCheckError  B2MatchInstruction = 21
	InstructionMatchSetupError  B2MatchInstruction = 22
	InstructionMatchIllegalMove B2MatchInstruction = 23
)
