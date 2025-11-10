package utilities

import (
	"math/rand/v2"
	"time"
)

// ExponentialBackoff returns base * 2^attempt duration, capped at max.
// attempt < 0 is treated as 0; attempts ≥ 63 immediately returns max
// to avoid left-shift overflow.
func ExponentialBackoff(attempt uint, base, max time.Duration) time.Duration {
	// attempts ≥63 immediately return max to avoid left-shift overflow.
	if attempt >= 63 {
		return max
	}

	// more efficient than converting to float and using math.Exp2
	// shifting bits to the left by n bits is equivalent to multiplying base by 2^n
	waitTime := base << attempt
	if waitTime > max {
		return max
	}

	return waitTime
}

// JitteredExpBackoff returns a random duration in the range [ExponentialBackoff - base, ExponentialBackoff].
// Using jitter helps prevent thundering herd retries.
func JitteredExpBackoff(attempt uint, base, max time.Duration) time.Duration {
	backoffDur := ExponentialBackoff(attempt, base, max)
	duration := rand.Int64N(int64(backoffDur+1-base)) + int64(base)
	return time.Duration(duration)
}
