package services

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// ExponentialBackoffStrategy implements an exponential backoff retry strategy.
// It increases the wait time exponentially with each retry attempt, up to a maximum delay.
type ExponentialBackoffStrategy struct {
	baseDelay  time.Duration // baseDelay is a minimum delay between retries (e.g., 1s).
	maxDelay   time.Duration // maxDelay is a maximum delay between retries (e.g., 5s).
	attempts   int           // attempts is a maximum number of retry attempts.
	multiplier float64       // multiplier is a growth multiplier for exponential backoff (e.g., 2.0)
}

// NewExponentialBackoffStrategy creates a new instance of ExponentialBackoffStrategy.
func NewExponentialBackoffStrategy(
	baseDelay, maxDelay time.Duration,
	attempts int,
	multiplier float64,
) *ExponentialBackoffStrategy {
	if baseDelay <= 0 {
		baseDelay = time.Duration(5) * time.Second
	}
	if maxDelay <= 0 {
		maxDelay = time.Duration(45) * time.Second
	}
	if multiplier <= 0 {
		multiplier = 2.0
	}

	return &ExponentialBackoffStrategy{
		baseDelay:  baseDelay,
		maxDelay:   maxDelay,
		attempts:   attempts,
		multiplier: multiplier,
	}
}

// WaitDuration calculates the duration to wait before the next retry attempt.
// It uses an exponential backoff formula: BaseDelay * Multiplier^attempt.
func (s *ExponentialBackoffStrategy) WaitDuration(attempt int) (result time.Duration, err error) {
	if err = s.validate(attempt); err != nil {
		return 0, fmt.Errorf("exponential backoff strategy: %w", err)
	}

	// Calculate exponential delay
	var delay = float64(s.baseDelay) * math.Pow(s.multiplier, float64(attempt))

	if delay > float64(s.maxDelay) {
		delay = float64(s.maxDelay)
	}

	return time.Duration(delay), nil
}

// validate checks the validity of the configuration and the given attempt number.
func (s *ExponentialBackoffStrategy) validate(number int) (err error) {
	if number < 0 {
		return errors.New("number must be greater than zero")
	}
	if s.attempts > 0 && number > s.attempts {
		return errors.New("maximum number of attempts exceeded")
	}
	return nil
}
