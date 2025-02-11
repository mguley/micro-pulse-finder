package proxy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// SetupTestContainer initializes the TestContainer.
func SetupTestContainer(t *testing.T) *TestContainer {
	container := NewTestContainer()
	conn := container.PortConnection.Get()

	t.Cleanup(func() {
		err := conn.Close()
		require.NoError(t, err, "Failed to close connection")
	})

	return container
}
