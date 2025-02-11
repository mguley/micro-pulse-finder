package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPortConnection_Dial tests the connection establishment.
func TestPortConnection_Dial(t *testing.T) {
	container := SetupTestContainer(t)
	conn := container.PortConnection.Get()

	// Attempt to establish connection
	connection, err := conn.Dial()
	require.NoError(t, err, "Failed to establish connection")
	require.NotNil(t, connection, "Connection should not be nil")

	// Verify connection status
	assert.True(t, conn.IsConnected(), "Expected connection to be active")
}

// TestPortConnection_Close tests closing the connection.
func TestPortConnection_Close(t *testing.T) {
	container := SetupTestContainer(t)
	conn := container.PortConnection.Get()

	// Establish connection
	connection, err := conn.Dial()
	require.NoError(t, err, "Failed to establish connection")
	require.NotNil(t, connection, "Connection should not be nil")

	// Close connection
	err = conn.Close()
	require.NoError(t, err, "Failed to close connection")

	// Verify connection is closed
	assert.False(t, conn.IsConnected(), "Expected connection to be closed")
}
