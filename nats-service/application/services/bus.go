package services

import (
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
)

// errNatsConnection indicates that the NATS connection is not established or closed.
var errNatsConnection = errors.New("nats connection is not established")

// Operations provides methods for interacting with a NATS messaging system.
type Operations struct {
	conn *nats.Conn
}

// NewOperations creates a new instance of Operations.
func NewOperations(conn *nats.Conn) *Operations {
	return &Operations{conn: conn}
}

// Publish sends a message to the specified NATS subject.
func (ops *Operations) Publish(subject string, data []byte) (err error) {
	if ops.conn == nil || ops.conn.IsClosed() {
		return errNatsConnection
	}

	if err = ops.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("could not publish: %w", err)
	}

	return nil
}

// Subscribe listens for messages on the specified NATS subject.
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
		return nil, fmt.Errorf("could not subscribe to subject %s: %w", subject, err)
	}

	return sub, nil
}
