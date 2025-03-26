package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config holds the gRPC server configuration.
//
// Fields:
//   - TLSEnabled: Indicates whether TLS is enabled.
//   - CertFile:   Path to the TLS certificate file.
//   - KeyFile:    Path to the TLS key file.
//   - Port:       Port on which the server listens.
type Config struct {
	TLSEnabled bool
	CertFile   string
	KeyFile    string
	Port       string
}

// Option defines a functional option for configuring the server.
type Option func(*Config)

// WithTLS enables TLS for the server and sets the certificate and key files.
//
// Parameters:
//   - certFile: Path to the TLS certificate file.
//   - keyFile:  Path to the TLS key file.
//
// Returns:
//   - Option: A functional option that modifies the server configuration.
func WithTLS(certFile, keyFile string) Option {
	return func(config *Config) {
		config.TLSEnabled = true
		config.CertFile = certFile
		config.KeyFile = keyFile
	}
}

// WithPort sets the server's listening port.
//
// Parameters:
//   - port: The port number as a string.
//
// Returns:
//   - Option: A functional option that modifies the server configuration.
func WithPort(port string) Option {
	return func(config *Config) {
		config.Port = port
	}
}

// NewGRPCServer initializes a new gRPC server with the provided options.
//
// Parameters:
//   - opts: A variadic list of Option functions to customize the server configuration.
//
// Returns:
//   - grpcServer: A pointer to the initialized gRPC server.
//   - config:     A pointer to the configuration used for the server.
//   - err:        An error if server initialization fails, or nil if successful.
func NewGRPCServer(opts ...Option) (grpcServer *grpc.Server, config *Config, err error) {
	config = &Config{
		TLSEnabled: false,
	}

	// Apply options to configure the server
	for _, opt := range opts {
		opt(config)
	}

	var transportCredentials credentials.TransportCredentials
	if config.TLSEnabled {
		if transportCredentials, err = credentials.NewServerTLSFromFile(config.CertFile, config.KeyFile); err != nil {
			return nil, nil, err
		}
		grpcServer = grpc.NewServer(grpc.Creds(transportCredentials))
	} else {
		grpcServer = grpc.NewServer()
	}

	return grpcServer, config, nil
}
