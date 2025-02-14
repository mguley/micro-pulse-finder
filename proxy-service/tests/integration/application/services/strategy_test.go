package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExponentialBackoff_WaitDuration tests that the exponential backoff calculates wait times correctly.
func TestExponentialBackoff_WaitDuration(t *testing.T) {
	container := SetupTestContainer()
	strategy := container.RetryStrategy.Get()

	// Define expected wait durations for each attempt
	expectedDurations := []time.Duration{
		time.Duration(5) * time.Second,  // Attempt 0: Base delay
		time.Duration(10) * time.Second, // Attempt 1: 5s * 2
		time.Duration(20) * time.Second, // Attempt 2: 10s * 2
		time.Duration(40) * time.Second, // Attempt 3: 20s * 2
		time.Duration(45) * time.Second, // Attempt 4: Should cap at maxDelay (45s)
	}

	for i, expected := range expectedDurations {
		duration, err := strategy.WaitDuration(i)
		require.NoError(t, err, "WaitDuration should not return an error for valid attempt")
		assert.Equal(t, expected, duration, "Unexpected wait duration for attempt %d", i)
	}
}

// TestExponentialBackoff_InvalidAttempts tests that invalid attempt numbers return errors.
func TestExponentialBackoff_InvalidAttempts(t *testing.T) {
	container := SetupTestContainer()
	strategy := container.RetryStrategy.Get()

	// Attempt with a negative number
	duration, err := strategy.WaitDuration(-1)
	require.Error(t, err, "Expected an error for negative attempt number")
	assert.Equal(t, time.Duration(0), duration, "Unexpected duration for negative attempt number")
	assert.Contains(t, err.Error(), "number must be greater than zero", "Unexpected error message")

	// Attempt exceeding allowed retries
	duration, err = strategy.WaitDuration(6)
	require.Error(t, err, "Expected an error for exceeding max attempts")
	assert.Equal(t, time.Duration(0), duration, "Unexpected duration for exceeding max attempts")
	assert.Contains(t, err.Error(), "maximum number of attempts exceeded", "Unexpected error message")
}
