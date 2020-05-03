package game

// reverseCardArray reverses an array in place
func reverseCardArray(inArray []Card) {
	for i, j := 0, len(inArray)-1; i < j; i, j = i+1, j-1 {
		inArray[i], inArray[j] = inArray[j], inArray[i]
	}
}

func removeFirstOfType(cards []Card, toRemove Card) (success bool) {
	var indexToRemove = -1

	for i := 0; i < len(cards); i++ {
		if cards[i] == toRemove {
			indexToRemove = i
			break
		}
	}

	if indexToRemove != -1 {
		cards[indexToRemove] = cards[len(cards)-1]
		cards = cards[:len(cards)-1]
	}

	return indexToRemove != -1
}

func removeLast(cards []Card) bool {
	if len(cards) > 0 {
		cards = cards[:len(cards)-1]
		return true
	}

	return false
}

func containsOnlyEffectCards(cardSet []Card) bool {
	for i := 0; i < len(cardSet); i++ {
		if cardSet[i] < Bolt {
			return false
		}
	}

	return true
}

func canBeatScore(cardSet []Card, scoreToBeat uint16) bool {
	for i := 0; i < len(cardSet); i++ {
		if uint16(cardSet[i].Value()) >= scoreToBeat {
			return true
		}
	}

	return false
}

func contains(cardSet []Card, cardToCheck Card) bool {
	for i := 0; i < len(cardSet); i++ {
		if cardSet[i] == cardToCheck {
			return true
		}
	}

	return false
}

func last(cardSet []Card) Card {
	if len(cardSet) > 0 {
		return cardSet[len(cardSet)-1]
	}

	return ElliotsOrbalStaff
}
