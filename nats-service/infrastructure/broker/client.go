package broker

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"
)

// Client represents a NATS client that manages connections to a NATS server.
//
// Fields:
//   - conn:    The underlying NATS connection.
//   - mu:      A read/write mutex for ensuring thread safety when accessing or modifying the connection.
//   - options: Connection options used to configure the NATS connection.
//   - logger:  Logger for structured logging of connection events.
type Client struct {
	conn    *nats.Conn
	mu      sync.RWMutex
	options *nats.Options
	logger  *slog.Logger
}

// NewClient creates a new NATS client instance with the specified options and logger.
//
// Parameters:
//   - options: A pointer to the NATS options for connection configuration.
//   - logger:  Logger instance for logging.
//
// Returns:
//   - *Client: A pointer to the newly created Client instance.
func NewClient(options *nats.Options, logger *slog.Logger) *Client {
	return &Client{options: options, logger: logger}
}

// Connect establishes a connection to the NATS server.
//
// Returns:
//   - connection: A pointer to the established NATS connection.
//   - err:        An error if the connection fails, or nil if successful.
func (c *Client) Connect() (connection *nats.Conn, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info("Connecting to NATS service", slog.String("url", c.options.Url))

	if c.conn != nil && !c.conn.IsClosed() {
		c.logger.Info("NATS connection is already active", slog.String("url", c.options.Url))
		return c.conn, nil
	}

	if c.conn, err = c.options.Connect(); err != nil {
		c.logger.Error("Failed to connect to NATS",
			slog.String("url", c.options.Url), slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not connect to NATS: %w", err)
	}

	c.logger.Info("Successfully connected to NATS", "url", c.options.Url)
	return c.conn, nil
}

// Close terminates the connection to the NATS server.
//
// Returns:
//   - err: An error if closing the connection fails, or nil if successful.
func (c *Client) Close() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info("Closing NATS connection", slog.String("url", c.options.Url))

	if c.conn == nil || c.conn.IsClosed() {
		c.logger.Info("NATS connection is already closed")
		return nil
	}

	c.conn.Close()
	c.conn = nil

	c.logger.Info("NATS connection is closed")
	return nil
}

// IsConnected checks if the client is currently connected to the NATS server.
//
// Returns:
//   - isConnected: A boolean value indicating whether the client is connected to the NATS server.
func (c *Client) IsConnected() (isConnected bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil || c.conn.IsClosed() {
		return false
	}

	return c.conn.IsConnected()
}
