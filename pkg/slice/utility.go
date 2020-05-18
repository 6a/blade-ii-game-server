// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package slice implements slice helper functions
package slice

import "errors"

// RemoveAtIndexUInt64 removes the element from slice s at index i. Returns an
// error if out of bounds.
func RemoveAtIndexUInt64(s *[]uint64, i int) error {

	// Perform a bounds check and return an error if invalid
	if i > len(*s) {
		return errors.New("Slice index out of bounds")
	}

	// Shift everything in slice s to the left by 1, at the index to be removed
	copy((*s)[i:], (*s)[i+1:])

	// Truncate the slice (cutting off the last one, which is now a duplicate of the second to last)
	*s = (*s)[:len(*s)-1]

	return nil
}
