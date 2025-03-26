package broker

import (
	"fmt"
	"log/slog"
	"shared/observability/nats-service/metrics"
	"sync"

	"github.com/nats-io/nats.go"
)

// Client represents a NATS client that manages connections to a NATS server.
//
// Fields:
//   - conn:    The underlying NATS connection.
//   - mu:      Mutex to ensure thread-safe operations.
//   - options: Connection options used to configure the NATS connection.
//   - metrics: Metrics instance for monitoring connection status and performance.
//   - logger:  Logger for structured logging of connection events.
type Client struct {
	conn    *nats.Conn
	mu      sync.Mutex
	options *nats.Options
	metrics *metrics.Metrics
	logger  *slog.Logger
}

// NewClient creates a new NATS client instance with the specified options.
//
// Parameters:
//   - options: A pointer to the NATS options for connection configuration.
//   - metrics: Metrics instance for monitoring.
//   - logger:  Logger instance for logging.
//
// Returns:
//   - *Client: A pointer to the newly created Client instance.
func NewClient(options *nats.Options, metrics *metrics.Metrics, logger *slog.Logger) *Client {
	return &Client{options: options, metrics: metrics, logger: logger}
}

// Connect establishes a connection to the NATS server.
//
// Returns:
//   - connection: A pointer to the established NATS connection.
//   - err:        An error if the connection fails, or nil if successful.
func (c *Client) Connect() (connection *nats.Conn, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info("Attempting to connect to NATS...", "url", c.options.Url)
	c.metrics.Connection.NATSConnectionAttempts.Inc()

	if c.conn != nil && !c.conn.IsClosed() {
		c.logger.Debug("NATS connection is already active")
		return c.conn, nil
	}

	if c.conn, err = c.options.Connect(); err != nil {
		c.logger.Error("Failed to connect to NATS", "url", c.options.Url, "error", err)
		c.metrics.Connection.NATSConnectionFailures.Inc()
		c.metrics.Connection.NATSConnectionStatus.Set(0)
		return nil, fmt.Errorf("could not connect to NATS: %w", err)
	}

	c.logger.Info("Successfully connected to NATS", "url", c.options.Url)
	c.metrics.Connection.NATSConnectionStatus.Set(1)

	return c.conn, nil
}

// Close terminates the connection to the NATS server.
//
// Returns:
//   - err: An error if closing the connection fails, or nil if successful.
func (c *Client) Close() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.conn.IsClosed() {
		return nil
	}

	c.conn.Close()
	c.conn = nil
	c.metrics.Connection.NATSConnectionStatus.Set(0)
	c.logger.Info("NATS connection closed")

	return nil
}

// IsConnected checks if the client is currently connected to the NATS server.
//
// Returns:
//   - isConnected: A boolean value indicating whether the client is connected to the NATS server.
func (c *Client) IsConnected() (isConnected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.conn.IsClosed() {
		c.metrics.Connection.NATSConnectionStatus.Set(0)
		return false
	}

	return c.conn.IsConnected()
}
