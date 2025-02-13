package interfaces

import "time"

// RetryStrategy defines a contract for retry strategy.
type RetryStrategy interface {
	// WaitDuration calculates the duration to wait before the next retry attempt.
	WaitDuration(attempt int) (sleepDuration time.Duration, err error)
}
