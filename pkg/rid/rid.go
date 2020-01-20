// Adapted from https://gist.github.com/denisbrodbeck/635a644089868a51eccd6ae22b2eb800

package rid

import (
	"crypto/rand"
	"math/big"
)

// RandomString returns a string of random characters of n length
func RandomString(n int) (r string, err error) {
	r = ""
	for {
		if len(r) >= n {
			return r, nil
		}

		num, err := rand.Int(rand.Reader, big.NewInt(int64(127)))
		if err != nil {
			return "", err
		}

		n := num.Int64()
		if isValidChar(n) {
			r += string(n)
		}
	}
}

func isValidChar(c int64) bool {
	// ASCII numbers, and upper/lower case characters
	return (c >= 48 && c <= 57) || (c >= 65 && c <= 90) || (c >= 97 && c <= 122)
}
