package commands

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusCommand_ValidResponse tests if the StatusCommand successfully fetches a response.
func TestStatusCommand_ValidResponse(t *testing.T) {
	container := SetupTestContainer()
	statusCmd := container.StatusCommand.Get()

	response, err := statusCmd.Execute()
	require.NoError(t, err, "Expected no error when executing status command")
	assert.NotEmpty(t, response, "Expected a non-empty response")
}

// TestStatusCommand_CircuitRotation ensures that each new StatusCommand execution results in a different exit IP.
func TestStatusCommand_CircuitRotation(t *testing.T) {
	type ipResponse struct {
		Origin string `json:"origin"`
	}

	container := SetupTestContainer()

	var (
		numRequests = 3
		results     = make(chan string, numRequests)
		errs        = make(chan error, numRequests)
		wg          sync.WaitGroup
	)

	wg.Add(numRequests)
	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()

			statusCmd := container.StatusCommand.Get()
			var (
				response string
				err      error
			)

			if response, err = statusCmd.Execute(); err != nil {
				errs <- err
				return
			}
			results <- response
		}()
	}
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		require.NoError(t, err, "Expected no error when executing status command concurrently")
	}

	// Process the responses and collect the unique IP addresses.
	var (
		ipAddresses = make(map[string]bool)
		list        = make([]string, numRequests)
	)
	for result := range results {
		var data ipResponse
		err := json.Unmarshal([]byte(result), &data)
		require.NoError(t, err, "Expected a valid JSON response")
		require.NotEmpty(t, data.Origin, "IP address should not be empty")
		ipAddresses[data.Origin] = true
		list = append(list, data.Origin)
	}

	// Assert that the number of unique IPs equals the number of requests.
	assert.Equal(t, numRequests, len(ipAddresses), "Expected each request to have a unique IP address")
	t.Logf("Collected IP addresses: %v", list)
}
