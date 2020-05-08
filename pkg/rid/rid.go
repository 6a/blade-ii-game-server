// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package rid implements cryptographically secure random strings. Adapted from
// https://gist.github.com/denisbrodbeck/635a644089868a51eccd6ae22b2eb800
package rid

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// RandomString returns a string of random characters of n length.
func RandomString(n int) (r string, err error) {

	// initialise a stringbuilder with default values (empty, ready to use).
	var stringBuilder strings.Builder

	// Iterate the string builder buffer is size (n), or an error causes an early exit (such as if the crypto package cannot
	// cannot find a proper native crypto module).
	for {

		// Exit once the string builder buffer is the correct size.
		if stringBuilder.Len() >= n {

			// Return the string builder as a string.
			return stringBuilder.String(), nil
		}

		// Generate a random int using the crypto rand package.
		num, err := rand.Int(rand.Reader, big.NewInt(int64(127)))
		if err != nil {
			return "", err
		}

		// Store the random int as a int64.
		n := num.Int64()

		// If the int is a valid char, add it to the string builder.
		if isValidChar(n) {
			stringBuilder.WriteString(string(n))
		}
	}
}

// isValidChar returns true if the specified int64 representation of an ASII character is within the
// valid range.
func isValidChar(c int64) bool {

	// ASCII numbers, and upper/lower case characters.
	return (c >= 48 && c <= 57) || (c >= 65 && c <= 90) || (c >= 97 && c <= 122)
}
