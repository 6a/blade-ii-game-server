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
