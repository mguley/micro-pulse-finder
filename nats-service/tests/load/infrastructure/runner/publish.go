package runner

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"shared/grpc/clients/nats_service"
)

// NatsServicePublishRunner implements the core.Runner interface for testing NATS service publish operations.
//
// Fields:
//   - client:      The NatsRpcClient used to interact with the NATS service.
//   - payload:     The generated random payload used in publish operations.
//   - subject:     The NATS subject to which messages will be published.
//   - messageSize: The size of the message payload in bytes.
//   - logger:      Logger for structured logging.
type NatsServicePublishRunner struct {
	client      *nats_service.NatsClient
	payload     []byte
	subject     string
	messageSize int
	logger      *slog.Logger
}

// NewNatsServicePublishRunner creates a new instance of NatsServicePublishRunner.
//
// Parameters:
//   - client:      The NatsRpcClient used to perform gRPC calls for publishing.
//   - messageSize: The size (in bytes) of the message payload to be generated.
//   - subject:     The NATS subject to publish messages to.
//   - logger:      Logger for structured logging.
//
// Returns:
//   - *NatsServicePublishRunner: A pointer to the newly created NatsServicePublishRunner.
func NewNatsServicePublishRunner(
	client *nats_service.NatsClient,
	messageSize int,
	subject string,
	logger *slog.Logger,
) *NatsServicePublishRunner {
	return &NatsServicePublishRunner{
		client:      client,
		messageSize: messageSize,
		subject:     subject,
		logger:      logger,
	}
}

// Setup performs necessary initialization for the publish runner.
// It generates a random payload of the specified message size and logs the setup status.
//
// Parameters:
//   - ctx: The context for the setup process.
//
// Returns:
//   - err: An error if payload generation fails; otherwise, nil.
func (r *NatsServicePublishRunner) Setup(ctx context.Context) (err error) {
	r.payload = make([]byte, r.messageSize)
	if _, err = rand.Read(r.payload); err != nil {
		return fmt.Errorf("failed to generate payload: %w", err)
	}

	r.logger.Info("NatsServicePublishRunner setup complete",
		slog.String("subject", r.subject),
		slog.Int("messageSize", r.messageSize))

	return nil
}

// Run publishes a message to the configured NATS subject using the underlying gRPC client.
//
// Parameters:
//   - ctx: The context for the publishing operation.
//
// Returns:
//   - err: An error if the publishing operation fails; otherwise, nil.
func (r *NatsServicePublishRunner) Run(ctx context.Context) (err error) {
	return r.client.Publish(ctx, r.subject, r.payload)
}

// Teardown cleans up resources by closing the underlying gRPC connection.
//
// Parameters:
//   - ctx: The context for the teardown process.
//
// Returns:
//   - err: An error if closing the connection fails; otherwise, nil.
func (r *NatsServicePublishRunner) Teardown(ctx context.Context) (err error) {
	if r.client != nil {
		if err = r.client.Close(); err != nil {
			return fmt.Errorf("failed to close grpc connection: %w", err)
		}
	}
	return err
}

// Name returns the descriptive name of the runner.
//
// Returns:
//   - string: The name of the runner ("NATS Service Publish Runner").
func (r *NatsServicePublishRunner) Name() (name string) {
	return "NATS Service Publish Runner"
}
