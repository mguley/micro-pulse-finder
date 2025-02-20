package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_Connect verifies that the MongoDB client can connect.
func TestClient_Connect(t *testing.T) {
	container := SetupTestContainer()
	client := container.MongoClient.Get()

	// Connect to MongoDB.
	mongoClient, err := client.Connect()
	require.NoError(t, err, "Expected no error on Connect()")
	require.NotNil(t, mongoClient, "Expected a non-nil mongo client after Connect()")

	// Verify that the client reports a connected state.
	assert.True(t, client.IsConnected(), "Expected IsConnected() to return true")
	t.Logf("Connected to MongoDB at: %t", client.IsConnected())

	err = client.Close()
	require.NoError(t, err, "Expected no error on Close()")
}

// TestClient_DoubleConnect ensures that calling Connect() multiple times returns the same client instance.
func TestClient_DoubleConnect(t *testing.T) {
	container := SetupTestContainer()
	client := container.MongoClient.Get()

	// First connection attempt.
	mongoClient1, err := client.Connect()
	require.NoError(t, err, "Expected no error on Connect()")
	require.NotNil(t, mongoClient1, "Expected a non-nil mongo client after Connect()")

	// Second connection attempt.
	mongoClient2, err := client.Connect()
	require.NoError(t, err, "Expected no error on Connect()")
	require.NotNil(t, mongoClient2, "Expected a non-nil mongo client after Connect()")

	// Both attempts should return the same underlying client.
	assert.Equal(t, mongoClient1, mongoClient2, "Expected both calls to return the same client instance")

	err = client.Close()
	require.NoError(t, err, "Expected no error on Close()")
}

// TestClient_Close demonstrates the behavior of closing the client.
func TestClient_Close(t *testing.T) {
	container := SetupTestContainer()
	client := container.MongoClient.Get()

	mongoClient, err := client.Connect()
	require.NoError(t, err, "Expected no error on Connect()")
	require.NotNil(t, mongoClient, "Expected a non-nil mongo client after Connect()")

	// Close the connection.
	err = client.Close()
	require.NoError(t, err, "Expected no error on Close()")

	// Verify that IsConnected() now returns false.
	assert.False(t, client.IsConnected(), "Expected IsConnected() to return false")
}
