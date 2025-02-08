package nats_service

import (
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds client configuration.
type Config struct {
	TLSEnabled bool   // TLSEnabled is used to indicate whether to use TLS.
	Address    string // Address is a target server address.
	CertFile   string // CertFile is a path to the certificate file (TLS).
}

// Option defines a functional option for configuring the client.
type Option func(*Config)

// WithTLS enables TLS and sets the certificate file for the client.
func WithTLS(certFile string) Option {
	return func(config *Config) {
		config.TLSEnabled = true
		config.CertFile = certFile
	}
}

// WithAddress sets the target server address.
func WithAddress(address string) Option {
	return func(config *Config) {
		config.Address = address
	}
}

// NewGRPCClient initializes a gRPC client connection with the provided options.
func NewGRPCClient(opts ...Option) (client *grpc.ClientConn, config *Config, err error) {
	config = &Config{
		TLSEnabled: false,
	}

	// Apply options to configure the client
	for _, opt := range opts {
		opt(config)
	}

	var (
		dialOpts             []grpc.DialOption
		transportCredentials credentials.TransportCredentials
		conn                 *grpc.ClientConn
	)

	if config.TLSEnabled {
		if transportCredentials, err = getTransportCredentials(config.CertFile); err != nil {
			return nil, nil, fmt.Errorf("could not get transport credentials: %w", err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(transportCredentials))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if conn, err = grpc.NewClient(config.Address, dialOpts...); err != nil {
		return nil, nil, fmt.Errorf("could not create client: %w", err)
	}
	return conn, config, nil
}

// getTransportCredentials determines and returns the correct transport credentials.
func getTransportCredentials(certFile string) (transportCredentials credentials.TransportCredentials, err error) {
	if strings.TrimSpace(certFile) != "" {
		// Use client-side TLS with the provided certificate
		return credentials.NewClientTLSFromFile(certFile, "")
	}
	// Use system CA trust store for validation
	return credentials.NewTLS(nil), nil
}
