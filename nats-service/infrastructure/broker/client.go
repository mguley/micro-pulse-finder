package broker

import (
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
)

// Client represents a NATS client that manages a connection to the NATS server.
type Client struct {
	conn    *nats.Conn    // conn is the underlying NATS connection.
	mu      sync.Mutex    // mu is the mutex to ensure thread-safe operations.
	options *nats.Options // options are the connection options.
}

// NewClient creates a new NATS client instance with the specified options.
func NewClient(options *nats.Options) *Client {
	return &Client{options: options}
}

// Connect establishes connection to the NATS server.
func (c *Client) Connect() (connection *nats.Conn, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn, nil
	}

	if c.conn, err = c.options.Connect(); err != nil {
		return nil, fmt.Errorf("could not connect to NATS: %w", err)
	}
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
	return nil
}

// IsConnected checks if the client is currently connected to the NATS server.
func (c *Client) IsConnected() (isConnected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.conn.IsClosed() {
		return false
	}
	return c.conn.IsConnected()
}
