package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Connect(t *testing.T) {
	container := SetupTestContainer(t)
	client := container.NatsClient.Get()

	// Connect
	conn, err := client.Connect()
	require.NoError(t, err, "Failed to connect to NATS server")
	assert.NotNil(t, conn, "Connection should not be nil")

	t.Logf("Connected to NATS at: %s", conn.ConnectedUrl())
}

func TestClient_IsConnected(t *testing.T) {
	container := SetupTestContainer(t)
	client := container.NatsClient.Get()

	// Connect
	conn, err := client.Connect()
	require.NoError(t, err, "Failed to connect to NATS server")
	assert.NotNil(t, conn, "Connection should not be nil")

	// Verify the connection status
	isConnected := client.IsConnected()
	assert.True(t, isConnected, "Client should be connected to NATS server")

	t.Log("Client connection status verified")
}

func TestClient_Close(t *testing.T) {
	container := SetupTestContainer(t)
	client := container.NatsClient.Get()

	// Connect
	conn, err := client.Connect()
	require.NoError(t, err, "Failed to connect to NATS server")
	assert.NotNil(t, conn, "Connection should not be nil")

	// Close connection
	err = client.Close()
	require.NoError(t, err, "Failed to close NATS connection")

	// Verify the connection status
	isConnected := client.IsConnected()
	assert.False(t, isConnected, "Client should not be connected after close")

	t.Log("Client successfully disconnected")
}
