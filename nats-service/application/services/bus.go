package services

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

// errNatsConnection indicates that the NATS connection is not established or closed.
var errNatsConnection = errors.New("nats connection is not established")

// Operations provides methods for interacting with a NATS messaging system.
//
// Fields:
//   - conn:    The active NATS connection used to send/receive messages.
//   - logger:  Logger used for logging operation statuses and errors.
type Operations struct {
	conn   *nats.Conn
	logger *slog.Logger
}

// NewOperations creates a new instance of Operations.
//
// Parameters:
//   - conn:    A pointer to the NATS connection.
//   - logger:  Logger instance for logging.
//
// Returns:
//   - *Operations: A pointer to the newly created Operations instance.
func NewOperations(conn *nats.Conn, logger *slog.Logger) *Operations {
	return &Operations{conn: conn, logger: logger}
}

// Publish sends a message to a specified NATS topic.
//
// Parameters:
//   - subject: The NATS subject to which the message will be published.
//   - data:    The byte slice representing the message payload.
//
// Returns:
//   - err: An error if the publish operation fails, or nil if successful.
func (ops *Operations) Publish(subject string, data []byte) (err error) {
	if ops.conn == nil || ops.conn.IsClosed() {
		ops.logger.Error("NATS connection is not available", "subject", subject)
		return errNatsConnection
	}

	if err = ops.conn.Publish(subject, data); err != nil {
		ops.logger.Error("Failed to publish message", "subject", subject, "error", err)
		return fmt.Errorf("could not publish: %w", err)
	}

	return nil
}

// Subscribe listens for messages on the specified NATS subject.
//
// Parameters:
//   - subject:    The NATS subject to subscribe to.
//   - queueGroup: (Optional) The queue group for load-balanced message processing.
//   - handler:    The message handler function that will process incoming messages.
//
// Returns:
//   - sub: A pointer to the NATS subscription if the subscription is successful.
//   - err: An error if the subscription operation fails; otherwise, nil.
func (ops *Operations) Subscribe(
	subject, queueGroup string,
	handler func(message *nats.Msg),
) (sub *nats.Subscription, err error) {
	if ops.conn == nil || ops.conn.IsClosed() {
		return nil, errNatsConnection
	}

	switch queueGroup {
	case "":
		sub, err = ops.conn.Subscribe(subject, handler)
	default:
		sub, err = ops.conn.QueueSubscribe(subject, queueGroup, handler)
	}

	if err != nil {
		ops.logger.Error("Failed to subscribe to queue", "subject", subject, "error", err)
		return nil, fmt.Errorf("could not subscribe to subject %s: %w", subject, err)
	}

	ops.logger.Info("Subscribed to queue", "subject", subject)

	return sub, nil
}
