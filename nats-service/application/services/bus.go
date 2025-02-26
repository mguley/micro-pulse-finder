package services

import (
	"errors"
	"fmt"
	"shared/observability/nats-service/metrics"
	"time"

	"github.com/nats-io/nats.go"
)

// errNatsConnection indicates that the NATS connection is not established or closed.
var errNatsConnection = errors.New("nats connection is not established")

// Operations provides methods for interacting with a NATS messaging system.
type Operations struct {
	conn    *nats.Conn
	metrics *metrics.Metrics
}

// NewOperations creates a new instance of Operations.
func NewOperations(conn *nats.Conn, metrics *metrics.Metrics) *Operations {
	return &Operations{conn: conn, metrics: metrics}
}

// Publish sends a message to the specified NATS subject.
func (ops *Operations) Publish(subject string, data []byte) (err error) {
	if ops.conn == nil || ops.conn.IsClosed() {
		ops.metrics.Message.FailedPublishAttempts.Inc()
		return errNatsConnection
	}

	start := time.Now()
	if err = ops.conn.Publish(subject, data); err != nil {
		ops.metrics.Message.FailedPublishAttempts.Inc()
		return fmt.Errorf("could not publish: %w", err)
	}

	ops.metrics.Message.TotalMessagesPublished.Inc()
	ops.metrics.Message.MessagePublishLatency.Observe(time.Since(start).Seconds())

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

	var msgHandler = func(msg *nats.Msg) {
		start := time.Now()
		handler(msg)
		ops.metrics.Message.TotalMessagesReceived.Inc()
		ops.metrics.Message.MessageProcessingDuration.Observe(time.Since(start).Seconds())
	}

	switch queueGroup {
	case "":
		sub, err = ops.conn.Subscribe(subject, msgHandler)
	default:
		sub, err = ops.conn.QueueSubscribe(subject, queueGroup, msgHandler)
	}

	if err != nil {
		return nil, fmt.Errorf("could not subscribe to subject %s: %w", subject, err)
	}

	ops.metrics.Subscription.ActiveSubscriptions.Inc()

	return sub, nil
}
