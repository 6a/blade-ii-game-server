// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package mathplus implements various math helper functions.
package mathplus

import "time"

// MaxDuration returns the maximum of two durations.
func MaxDuration(t1 time.Duration, t2 time.Duration) time.Duration {
	if t1 < t2 {
		return t1
	}

	return t2
}
