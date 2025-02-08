package server

import (
	"fmt"
	"log"
	"net"
	natsservicev1 "shared/proto/nats-service/gen"

	"google.golang.org/grpc"
)

// TestServerContainer manages the lifecycle of a test gRPC server.
type TestServerContainer struct {
	grpcServer *grpc.Server
	listener   net.Listener
	Address    string
}

// NewTestServerContainer initializes and starts a test gRPC server.
func NewTestServerContainer(busServer natsservicev1.BusServiceServer) (container *TestServerContainer, err error) {
	var (
		listener   net.Listener
		grpcServer *grpc.Server
		network    = "tcp"
		address    = "127.0.0.1:0"
	)

	if listener, err = net.Listen(network, address); err != nil {
		return nil, fmt.Errorf("could not listen : %w", err)
	}

	grpcServer = grpc.NewServer()
	natsservicev1.RegisterBusServiceServer(grpcServer, busServer)

	go func() {
		if err = grpcServer.Serve(listener); err != nil {
			log.Fatalf("could not serve : %v", err)
		}
	}()

	return &TestServerContainer{
		grpcServer: grpcServer,
		listener:   listener,
		Address:    listener.Addr().String(),
	}, nil
}

// Stop gracefully stops the test gRPC server.
func (s *TestServerContainer) Stop() {
	s.grpcServer.GracefulStop()
}
