package game

// reverseCardArray reverses an array in place
func reverseCardArray(inArray []Card) {
	for i, j := 0, len(inArray)-1; i < j; i, j = i+1, j-1 {
		inArray[i], inArray[j] = inArray[j], inArray[i]
	}
}
