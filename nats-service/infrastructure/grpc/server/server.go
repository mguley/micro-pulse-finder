package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	natsservicev1 "shared/proto/nats-service/gen"
	"syscall"

	"google.golang.org/grpc"
)

// BusServer is a wrapper around the underlying gRPC server for the BusService.
//
// It manages server initialization, registration, and lifecycle.
//
// Fields:
//   - grpcServer: The gRPC server instance that handles RPC requests.
//   - listener:   The network listener for incoming connections.
//   - port:       Port on which the server listens.
//   - env:        The environment in which the server is running (e.g., "prod" or "dev").
//   - logger:     Logger for structured logging of server events.
type BusServer struct {
	grpcServer *grpc.Server
	listener   net.Listener
	port       string
	env        string
	logger     *slog.Logger
}

// NewBusServer creates a new instance of BusServer based on the provided configuration.
//
// Parameters:
//   - env:      The environment (e.g., "prod" or "dev").
//   - port:     The port on which the server should listen.
//   - certFile: Path to the TLS certificate file (used in "prod").
//   - keyFile:  Path to the TLS key file (used in "prod").
//   - logger:   Logger instance for logging.
//
// Returns:
//   - busServer: A pointer to the newly created BusServer.
//   - err:       An error if server creation fails, or nil if successful.
func NewBusServer(env, port, certFile, keyFile string, logger *slog.Logger) (busServer *BusServer, err error) {
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
		logger:     logger,
	}, nil
}

// RegisterService registers the BusService with the gRPC server.
//
// Parameters:
//   - service: The BusServiceServer implementation to register.
func (s *BusServer) RegisterService(service natsservicev1.BusServiceServer) {
	natsservicev1.RegisterBusServiceServer(s.grpcServer, service)
}

// Start begins serving incoming gRPC requests.
//
// It starts the server in a separate goroutine.
func (s *BusServer) Start() {
	s.logger.Info("Starting the Bus gRPC server...", "address", s.listener.Addr(), "env", s.env)
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			s.logger.Error("Bus gRPC server failed to serve", "error", err)
			panic(err)
		}
	}()
}

// WaitForShutdown gracefully shuts down the gRPC server upon receiving termination signals.
//
// It waits for a shutdown signal and then gracefully stops the server.
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

		s.logger.Info("Received shutdown signal. Initiating shutdown...", "signal", sig)
		s.grpcServer.GracefulStop()
		s.logger.Info("Bus gRPC server stopped gracefully")

		defer close(done)
	}()

	<-done
}

// GracefulStop stops the gRPC server gracefully.
func (s *BusServer) GracefulStop() {
	s.grpcServer.GracefulStop()
}
