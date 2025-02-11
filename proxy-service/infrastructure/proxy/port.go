package proxy

import (
	"bufio"
	"errors"
	"fmt"
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
}

// NewConnection creates a new Connection instance with the specified timeout.
func NewConnection(timeout time.Duration) *Connection {
	return &Connection{timeout: timeout, network: "tcp"}
}

// Dial establishes a connection to the proxy control port.
func (c *Connection) Dial() (connection net.Conn, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn, nil
	}

	var (
		port    = entities.GetControlPort() // Retrieve the control port configuration.
		address string                      // Address to connect to.
	)

	if address, err = port.Address(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if c.conn, err = net.DialTimeout(c.network, address, c.timeout); err != nil {
		return nil, fmt.Errorf("could not connect to %s: %w", address, err)
	}

	c.reader = bufio.NewReader(c.conn)
	return c.conn, nil
}

// Close terminates the connection to the proxy control port.
func (c *Connection) Close() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	if err = c.conn.Close(); err != nil {
		return fmt.Errorf("could not close connection: %w", err)
	}

	c.conn = nil
	c.reader = nil
	return nil
}

// IsConnected checks if the connection is currently active.
func (c *Connection) IsConnected() (isConnected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn != nil
}

// ReadLine reads a single line from the connection.
func (c *Connection) ReadLine() (line string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.reader == nil {
		return "", errors.New("connection is not established")
	}

	if line, err = c.reader.ReadString('\n'); err != nil {
		return "", fmt.Errorf("could not read line: %w", err)
	}
	return line, nil
}
