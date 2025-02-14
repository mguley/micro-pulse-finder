package control

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAuthenticateCommand_Success validates successful authentication.
func TestAuthenticateCommand_Success(t *testing.T) {
	container := SetupTestContainer()
	authCmd := container.AuthenticateCommand.Get()

	// Execute authentication
	err := authCmd.Execute()
	require.NoError(t, err, "Authenticate should succeed")
}
