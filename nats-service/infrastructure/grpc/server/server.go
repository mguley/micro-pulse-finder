package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	natsservicev1 "shared/proto/nats-service/gen"
	"syscall"

	"google.golang.org/grpc"
)

// BusServer is a wrapper for the gRPC BusService server.
// It manages server initialization, registration and lifecycle.
type BusServer struct {
	grpcServer *grpc.Server // The gRPC server to serve RPC requests.
	listener   net.Listener // The network listener for the server.
	port       string       // The port the server listens on.
	env        string       // The environment (e.g., "prod" or "dev").
}

// NewBusServer creates a new instance of BusServer based on the provided configuration.
func NewBusServer(env, port, certFile, keyFile string) (busServer *BusServer, err error) {
	var (
		grpcServer   *grpc.Server
		serverConfig *Config
		listener     net.Listener
		network      = "tcp"
	)

	switch env {
	case "dev":
		grpcServer, serverConfig, err = NewGRPCServer(WithPort(port))
	case "prod":
		grpcServer, serverConfig, err = NewGRPCServer(WithTLS(certFile, keyFile), WithPort(port))
	default:
		return nil, errors.New("unsupported environment; must be \"prod\" or \"dev\"")
	}

	if err != nil {
		return nil, fmt.Errorf("create gRPC server: %w", err)
	}

	if listener, err = net.Listen(network, fmt.Sprintf(":%s", serverConfig.Port)); err != nil {
		return nil, fmt.Errorf("create gRPC listener: %w", err)
	}

	return &BusServer{
		grpcServer: grpcServer,
		listener:   listener,
		port:       port,
		env:        env,
	}, nil
}

// RegisterService registers the BusService implementation with the gRPC server.
func (s *BusServer) RegisterService(service natsservicev1.BusServiceServer) {
	natsservicev1.RegisterBusServiceServer(s.grpcServer, service)
}

// Start starts the gRPC server and begins listening for incoming requests.
func (s *BusServer) Start() {
	log.Printf("Starting the Bus gRPC server on %s (env: %s)...", s.listener.Addr(), s.env)
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			log.Fatalf("Bus gRPC server failed to serve: %v", err)
		}
	}()
}

// WaitForShutdown gracefully shuts down the server upon receiving termination signals.
func (s *BusServer) WaitForShutdown() {
	var (
		signalChan = make(chan os.Signal, 1)
		done       = make(chan struct{})
	)

	// Notify
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		defer close(signalChan)

		log.Printf("Received signal: %s. Initiating shutdown...", sig)
		s.grpcServer.GracefulStop()
		log.Println("Proxy gRPC server stopped gracefully.")

		defer close(done)
	}()

	<-done
}

// GracefulStop stops the gRPC server gracefully.
func (s *BusServer) GracefulStop() {
	s.grpcServer.GracefulStop()
}
