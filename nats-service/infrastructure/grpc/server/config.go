package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config holds the server configuration.
type Config struct {
	TLSEnabled bool   // TLSEnabled is used to indicate whether to use TLS.
	CertFile   string // CertFile is a path to the TLS certificate file.
	KeyFile    string // KeyFile is a path to the TLS key file.
	Port       string // Port is the port the server listens on.
}

// Option defines a functional option for configuring the server.
type Option func(*Config)

// WithTLS enables TLS and sets the certificate and key files.
func WithTLS(certFile, keyFile string) Option {
	return func(config *Config) {
		config.TLSEnabled = true
		config.CertFile = certFile
		config.KeyFile = keyFile
	}
}

// WithPort sets the server's listening port.
func WithPort(port string) Option {
	return func(config *Config) {
		config.Port = port
	}
}

// NewGRPCServer initializes gRPC server with the provided options.
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
