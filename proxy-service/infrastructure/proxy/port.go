package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"proxy-service/domain/entities"
	"sync"
	"time"
)

// Connection represents a TCP connection to the proxy control port.
// It provides methods to establish, interact with and close the connection.
type Connection struct {
	conn    net.Conn      // conn is the underlying TCP connection.
	reader  *bufio.Reader // reader is the buffered reader for reading from the connection.
	mu      sync.Mutex    // mu is the mutex to ensure thread-safe operations.
	timeout time.Duration // timeout is the timeout duration for establishing the connection.
	network string        // network is the network type (e.g., "tcp").
	logger  *slog.Logger
}

// NewConnection creates a new Connection instance with the specified timeout.
func NewConnection(timeout time.Duration, logger *slog.Logger) *Connection {
	return &Connection{timeout: timeout, network: "tcp", logger: logger}
}

// Dial establishes a connection to the proxy control port.
func (c *Connection) Dial() (connection net.Conn, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.logger.Debug("Already connected to proxy control port")
		return c.conn, nil
	}

	var (
		port    = entities.GetControlPort() // Retrieve the control port configuration.
		address string                      // Address to connect to.
	)

	if address, err = port.Address(); err != nil {
		c.logger.Error("Failed to get proxy control port address", "error", err)
		return nil, fmt.Errorf("%w", err)
	}

	c.logger.Info("Dialing proxy control port", "address", address, "timeout", c.timeout)
	if c.conn, err = net.DialTimeout(c.network, address, c.timeout); err != nil {
		c.logger.Error("Could not connect to proxy control port", "address", address, "error", err)
		return nil, fmt.Errorf("could not connect to %s: %w", address, err)
	}

	c.reader = bufio.NewReader(c.conn)
	c.logger.Info("Successfully connected to proxy control port", "address", address)
	return c.conn, nil
}

// Close terminates the connection to the proxy control port.
func (c *Connection) Close() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		c.logger.Debug("No connection to close")
		return nil
	}

	c.logger.Info("Closing connection to proxy control port")
	if err = c.conn.Close(); err != nil {
		c.logger.Error("Could not close connection", "error", err)
		return fmt.Errorf("could not close connection: %w", err)
	}

	c.conn = nil
	c.reader = nil
	c.logger.Info("Connection closed")
	return nil
}

// IsConnected checks if the connection is currently active.
func (c *Connection) IsConnected() (isConnected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	isConnected = c.conn != nil
	c.logger.Debug("Checking connection status", "isConnected", isConnected)
	return isConnected
}

// ReadLine reads a single line from the connection.
func (c *Connection) ReadLine() (line string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.reader == nil {
		c.logger.Error("Attempted to read from an unestablished connection")
		return "", errors.New("connection is not established")
	}

	if line, err = c.reader.ReadString('\n'); err != nil {
		c.logger.Error("Failed to read line from connection", "error", err)
		return "", fmt.Errorf("could not read line: %w", err)
	}

	c.logger.Debug("Read line from connection", "line", line)
	return line, nil
}
