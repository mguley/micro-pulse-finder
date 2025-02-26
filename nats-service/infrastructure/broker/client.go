package broker

import (
	"fmt"
	"shared/observability/nats-service/metrics"
	"sync"

	"github.com/nats-io/nats.go"
)

// Client represents a NATS client that manages a connection to the NATS server.
type Client struct {
	conn    *nats.Conn    // conn is the underlying NATS connection.
	mu      sync.Mutex    // mu is the mutex to ensure thread-safe operations.
	options *nats.Options // options are the connection options.
	metrics *metrics.Metrics
}

// NewClient creates a new NATS client instance with the specified options.
func NewClient(options *nats.Options, metrics *metrics.Metrics) *Client {
	return &Client{options: options, metrics: metrics}
}

// Connect establishes connection to the NATS server.
func (c *Client) Connect() (connection *nats.Conn, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.Connection.NATSConnectionAttempts.Inc()

	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn, nil
	}

	if c.conn, err = c.options.Connect(); err != nil {
		c.metrics.Connection.NATSConnectionFailures.Inc()
		c.metrics.Connection.NATSConnectionStatus.Set(0)
		return nil, fmt.Errorf("could not connect to NATS: %w", err)
	}

	c.metrics.Connection.NATSConnectionStatus.Set(1)

	return c.conn, nil
}

// Close terminates the connection to the NATS server.
func (c *Client) Close() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.conn.IsClosed() {
		return nil
	}

	c.conn.Close()
	c.conn = nil
	c.metrics.Connection.NATSConnectionStatus.Set(0)

	return nil
}

// IsConnected checks if the client is currently connected to the NATS server.
func (c *Client) IsConnected() (isConnected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.conn.IsClosed() {
		c.metrics.Connection.NATSConnectionStatus.Set(0)
		return false
	}

	return c.conn.IsConnected()
}
