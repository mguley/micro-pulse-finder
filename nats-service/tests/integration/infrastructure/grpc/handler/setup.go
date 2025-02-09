package handler

import (
	"net"
	natsservicev1 "shared/proto/nats-service/gen"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SetupTestContainer initializes TestContainer.
func SetupTestContainer(t *testing.T) (client natsservicev1.BusServiceClient) {
	container := NewTestContainer()

	// Create a listener for the in-process gRPC server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "Failed to create listener")

	// Initialize the gRPC server and register the BusService
	server := grpc.NewServer()
	busService := container.BusService.Get()
	natsservicev1.RegisterBusServiceServer(server, busService)

	// Start the server
	go func() {
		err = server.Serve(listener)
		require.NoError(t, err, "Failed to start gRPC server")
	}()

	// Ensure server stops after tests
	t.Cleanup(func() {
		server.GracefulStop()
	})

	// Set up a gRPC client to interact with the server
	target := listener.Addr().String()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(target, opts...)
	require.NoError(t, err, "Failed to connect to gRPC server")

	// Ensure client connection is closed
	t.Cleanup(func() {
		err = conn.Close()
		require.NoError(t, err, "Failed to close gRPC connection")
	})

	return natsservicev1.NewBusServiceClient(conn)
}
