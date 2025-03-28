package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

// Operations provides methods for interacting with the NATS message broker.
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
//   - conn:    A pointer to the active NATS connection.
//   - logger:  A pointer to the logger to be used for logging.
//
// Returns:
//   - *Operations: A pointer to the newly created Operations instance.
func NewOperations(conn *nats.Conn, logger *slog.Logger) *Operations {
	return &Operations{conn: conn, logger: logger}
}

// Publish sends a message to a specified NATS topic.
//
// Parameters:
//   - ctx:     Context for managing timeouts and cancellation signals.
//   - subject: The subject/topic to which the message will be published.
//   - data:    The byte slice representing the message payload.
//
// Returns:
//   - err: An error if the publish operation fails, or nil if successful.
func (o *Operations) Publish(ctx context.Context, subject string, data []byte) (err error) {
	if o.conn == nil || o.conn.IsClosed() {
		o.logger.Error("NATS connection is not established", slog.String("topic", subject))
		return fmt.Errorf("connection is not established")
	}

	select {
	case <-ctx.Done():
		o.logger.Info("Context canceled before publishing", slog.String("topic", subject))
		return ctx.Err()
	default:
		if err = o.conn.Publish(subject, data); err != nil {
			o.logger.Error("NATS connection publish failed",
				slog.String("topic", subject), slog.String("error", err.Error()))
			return fmt.Errorf("could not send message to NATS: %w", err)
		}
	}

	return err
}

// Subscribe listens for messages on the specified NATS subject.
//
// Parameters:
//   - ctx:        Context for managing timeouts and cancellation signals.
//   - subject:    The subject/topic to subscribe to.
//   - queueGroup: (Optional) The queue group for load-balanced message processing.
//   - handler:    The message handler function that will process incoming messages.
//
// Returns:
//   - sub: A pointer to the NATS subscription if the subscription is successful.
//   - err: An error if the subscription operation fails; otherwise, nil.
func (o *Operations) Subscribe(
	ctx context.Context,
	subject, queueGroup string,
	handler func(message *nats.Msg),
) (sub *nats.Subscription, err error) {
	if o.conn == nil || o.conn.IsClosed() {
		o.logger.Error("NATS connection is not established", slog.String("topic", subject))
		return nil, fmt.Errorf("connection is not established")
	}

	select {
	case <-ctx.Done():
		o.logger.Info("Context canceled before subscription", slog.String("topic", subject))
		return nil, ctx.Err()
	default:
		switch queueGroup {
		case "":
			sub, err = o.conn.Subscribe(subject, handler)
		default:
			sub, err = o.conn.QueueSubscribe(subject, queueGroup, handler)
		}

		if err != nil {
			o.logger.Error("NATS connection subscribe failed",
				slog.String("topic", subject), slog.String("error", err.Error()))
			return nil, fmt.Errorf("could not subscribe to NATS subject: %w", err)
		}

		return sub, nil
	}
}
