package nats_service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	natsservicev1 "shared/proto/nats-service/gen"

	"google.golang.org/grpc"
)

// NatsClient is a wrapper over the underlying gRPC client connection to BusService.
type NatsClient struct {
	conn      *grpc.ClientConn               // conn is the underlying gRPC client connection.
	client    natsservicev1.BusServiceClient // client is the generated BusService client.
	validator Validator                      // validator is the gRPC client requests validator.
	logger    *slog.Logger                   // logger for structured logging.
}

// NewNatsClient creates a new instance of NatsClient.
func NewNatsClient(env, address string, validator Validator, logger *slog.Logger) (natsClient *NatsClient, err error) {
	var (
		conn   *grpc.ClientConn
		config *Config
	)

	switch env {
	case "prod":
		conn, config, err = NewGRPCClient(WithAddress(address), WithTLS(""))
	case "dev":
		conn, config, err = NewGRPCClient(WithAddress(address))
	default:
		return nil, errors.New("unsupported environment; must be \"prod\" or \"dev\"")
	}

	if err != nil {
		logger.Error("Failed to create gRPC client", "error", err)
		return nil, fmt.Errorf("create NATS client: %w", err)
	}

	logger.Info("New channel established", "address", address, "tls_enabled", config.TLSEnabled)
	return &NatsClient{
		conn:      conn,
		client:    natsservicev1.NewBusServiceClient(conn),
		validator: validator,
		logger:    logger,
	}, nil
}

// Publish sends a message to the specified NATS subject.
func (c *NatsClient) Publish(ctx context.Context, subject string, data []byte) (err error) {
	var (
		request  = natsservicev1.PublishRequest{Subject: subject, Data: data}
		response *natsservicev1.PublishResponse
	)

	// Validate request before sending
	if err = c.validator.ValidatePublishRequest(&request); err != nil {
		c.logger.Error("Validation failed for publish request", "subject", subject, "error", err)
		return fmt.Errorf("validate publish request: %w", err)
	}

	// RPC call
	if response, err = c.client.Publish(ctx, &request); err != nil {
		c.logger.Error("Failed to publish message", "subject", subject, "error", err)
		return fmt.Errorf("message publish: %w", err)
	}

	if !response.GetSuccess() {
		c.logger.Error("Publish response indicates failure", "subject", subject, "message", response.GetMessage())
		return fmt.Errorf("could not publish message: %s", response.GetMessage())
	}

	c.logger.Info("Message published successfully", "subject", subject)
	return nil
}

// Subscribe listens for messages on a specified NATS subject and processes them via a callback function.
func (c *NatsClient) Subscribe(
	ctx context.Context,
	subject, queueGroup string,
	handler func(data []byte, subject string),
) (err error) {
	var (
		request = natsservicev1.SubscribeRequest{Subject: subject, QueueGroup: queueGroup}
		stream  grpc.ServerStreamingClient[natsservicev1.SubscribeResponse]
		message *natsservicev1.SubscribeResponse
	)

	// Validate request before subscribing
	if err = c.validator.ValidateSubscribeRequest(&request); err != nil {
		c.logger.Error("Validation failed for subscribe request",
			"subject", subject, "queueGroup", queueGroup, "error", err)
		return fmt.Errorf("validate subscribe request: %w", err)
	}

	// Open a gRPC streaming connection
	if stream, err = c.client.Subscribe(ctx, &request); err != nil {
		c.logger.Error("Failed to subscribe to subject", "subject", subject, "error", err)
		return fmt.Errorf("subscribe to subject %s: %w", subject, err)
	}
	c.logger.Info("Subscribed to NATS subject", "subject", subject)

	// Continuously listen for messages from the gRPC stream
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Subscription canceled", "subject", subject)
			return ctx.Err()
		default:
			if message, err = stream.Recv(); err != nil {
				switch err {
				case io.EOF:
					c.logger.Info("End of stream reached for subscription", "subject", subject)
					return nil
				default:
					c.logger.Error("Error receiving message from NATS", "subject", subject, "error", err)
					return fmt.Errorf("receive message from NATS: %w", err)
				}
			}
			// Process the received message using the provided handler
			handler(message.GetData(), message.GetSubject())
		}
	}
}

// Close closes the underlying gRPC connection.
func (c *NatsClient) Close() (err error) {
	if err = c.conn.Close(); err != nil {
		c.logger.Error("Failed to close NATS client connection", "error", err)
		return err
	}
	return nil
}
