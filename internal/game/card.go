package game

// Card is a typedef for different card types
type Card uint8

// Card enums - explicitly numbered so we can see their values when hovered
const (
	ElliotsOrbalStaff           Card = 0
	FiesTwinGunswords           Card = 1
	AlisasOrbalBow              Card = 2
	JusisSword                  Card = 3
	MachiasOrbalShotgun         Card = 4
	GaiusSpear                  Card = 5
	LaurasGreatsword            Card = 6
	Bolt                        Card = 7
	Mirror                      Card = 8
	Blast                       Card = 9
	Force                       Card = 10
	InactiveElliotsOrbalStaff   Card = 11
	InactiveFiesTwinGunswords   Card = 12
	InactiveAlisasOrbalBow      Card = 13
	InactiveJusisSword          Card = 14
	InactiveMachiasOrbalShotgun Card = 15
	InactiveGaiusSpear          Card = 16
	InactiveLaurasGreatsword    Card = 17
	InactiveBolt                Card = 18
	InactiveMirror              Card = 19
	InactiveBlast               Card = 20
	InactiveForce               Card = 21
)

// Value returns the point value of the specified card
func (c *Card) Value() (value uint8) {
	if *c < Bolt {
		value = uint8(*c) + 1
	} else if *c >= Bolt && *c <= Force {
		value = 1
	} else {
		value = 0
	}

	return value
}
