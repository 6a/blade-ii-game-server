package mathplus

import "time"

// MaxDuration returns the maximum of two durations.
func MaxDuration(t1 time.Duration, t2 time.Duration) time.Duration {
	if t1 < t2 {
		return t1
	}

	return t2
}
