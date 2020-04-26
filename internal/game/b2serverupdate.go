package game

// B2MatchInstruction is an enum type that represents different different types of match instruction for Blade II Online
type B2MatchInstruction uint8

// Definitions of server updates
const (
	None                        B2MatchInstruction = 0
	CardElliotsOrbalStaff       B2MatchInstruction = 1
	CardFiesTwinGunswords       B2MatchInstruction = 2
	CardAlisasOrbalBow          B2MatchInstruction = 3
	CardJusisSword              B2MatchInstruction = 4
	CardMachiasOrbalShotgun     B2MatchInstruction = 5
	CardGaiusSpear              B2MatchInstruction = 6
	CardLaurasGreatsword        B2MatchInstruction = 7
	CardBolt                    B2MatchInstruction = 8
	CardMirror                  B2MatchInstruction = 9
	CardBlast                   B2MatchInstruction = 10
	CardForce                   B2MatchInstruction = 11
	InstructionCards            B2MatchInstruction = 12
	InstructionForfeit          B2MatchInstruction = 13
	InstructionMessage          B2MatchInstruction = 14
	InstructionConnectionFailed B2MatchInstruction = 15
	InstructionPlayerData       B2MatchInstruction = 16
	InstructionOpponentData     B2MatchInstruction = 17
)
