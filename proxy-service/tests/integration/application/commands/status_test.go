package commands

import (
	"encoding/json"
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
	ipAddresses := make(map[string]bool)
	var list []string

	// Execute the StatusCommand multiple times and collect information.
	for i := 0; i < 3; i++ {
		statusCmd := container.StatusCommand.Get()

		response, err := statusCmd.Execute()
		require.NoError(t, err, "Expected no error when executing status command")
		require.NotEmpty(t, response, "Expected a non-empty response")

		var data ipResponse
		err = json.Unmarshal([]byte(response), &data)
		require.NoError(t, err, "Expected a valid response")
		require.NotEmpty(t, data.Origin, "IP address should not be empty")

		ipAddresses[data.Origin] = true
		list = append(list, data.Origin)
	}

	assert.Equal(t, 3, len(ipAddresses), "Expected each client to have a unique IP address")
	t.Logf("Collected IP addresses: %v", list)
}
