package nats_service

import (
	"shared/grpc/clients/nats_service"
	"shared/grpc/tests/integration/clients/nats_service/server"
	"testing"
)

// TestEnvironment encapsulates the mock server and gRPC client for integration tests.
type TestEnvironment struct {
	Server *server.TestServerContainer
	Client *nats_service.NatsClient
}

// SetupTestEnvironment initializes the test environment.
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	container := NewTestContainer()

	grpcServer := container.TestServerContainer.Get()
	grpcClient := container.NatsClient.Get()

	t.Cleanup(func() {
		_ = grpcClient.Close()
		grpcServer.Stop()
	})

	return &TestEnvironment{
		Server: grpcServer,
		Client: grpcClient,
	}
}
