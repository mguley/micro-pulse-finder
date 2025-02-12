package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignalCommand_Success validates that the SIGNAL command succeeds after authentication.
func TestSignalCommand_Success(t *testing.T) {
	container := SetupTestContainer()
	authCmd := container.AuthenticateCommand.Get()
	signalCmd := container.SignalCommand.Get()

	// Authenticate
	err := authCmd.Execute()
	require.NoError(t, err, "Authentication should succeed before sending SIGNAL command")

	// Send signal
	err = signalCmd.Execute()
	require.NoError(t, err, "SIGNAL command should succeed after authentication")
}

// TestSignalCommand_WithoutAuthentication ensures that an unauthenticated SIGNAL command fails.
func TestSignalCommand_WithoutAuthentication(t *testing.T) {
	container := SetupTestContainer()
	signalCmd := container.SignalCommand.Get()

	// Attempt to send a SIGNAL command without authentication
	err := signalCmd.Execute()
	require.Error(t, err, "SIGNAL command should fail without authentication")
	assert.Equal(t, "authentication is required", err.Error(), "Expected authentication required error")
}

// TestSignalCommand_CloseConnectionAfterExecution ensures the connection closes after sending a SIGNAL command.
func TestSignalCommand_CloseConnectionAfterExecution(t *testing.T) {
	container := SetupTestContainer()
	authCmd := container.AuthenticateCommand.Get()
	signalCmd := container.SignalCommand.Get()

	// Authenticate
	err := authCmd.Execute()
	require.NoError(t, err, "Authentication should succeed")

	// Send SIGNAL command
	err = signalCmd.Execute()
	require.NoError(t, err, "SIGNAL command should succeed")

	// Close the connection
	err = signalCmd.Close()
	require.NoError(t, err, "Closing connection should succeed")

	// Ensure the connection is closed
	assert.False(t, container.PortConnection.Get().IsConnected(), "Expected connection to be closed after execution")
}
