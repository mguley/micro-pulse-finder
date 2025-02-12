package control

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"proxy-service/application/commands/control/codes"
	"proxy-service/infrastructure/proxy"
	"strings"
)

// BaseCommand provides common functionality for commands interacting with the proxy control port.
type BaseCommand struct {
	adapter *proxy.Connection // adapter is a common adapter for proxy commands.
	conn    net.Conn          // conn is the underlying TCP connection.
}

// Initialize establishes a connection to the proxy control port.
func (b *BaseCommand) Initialize() (err error) {
	if b.adapter == nil {
		return errors.New("connection adapter is not initialized")
	}

	if b.conn, err = b.adapter.Dial(); err != nil {
		return fmt.Errorf("could not connect to proxy: %w", err)
	}
	return nil
}

// Close closes the established connection.
func (b *BaseCommand) Close() (err error) {
	return b.adapter.Close()
}

// SendCommand sends a formatted command to the proxy server.
func (b *BaseCommand) SendCommand(command string) (err error) {
	if !b.adapter.IsConnected() {
		return errors.New("connection is not established")
	}

	var (
		writer       *bufio.Writer
		bytesWritten int
	)

	writer = bufio.NewWriter(b.conn)
	if bytesWritten, err = writer.WriteString(command); err != nil {
		return fmt.Errorf("could not send command to proxy: %w", err)
	}

	if err = writer.Flush(); err != nil {
		return fmt.Errorf("could not flush writer: %w", err)
	}

	fmt.Printf("[INFO]: command %s wrote %d bytes to proxy\n", command, bytesWritten)
	return nil
}

// ReadResponse reads and processes the response from the proxy server.
func (b *BaseCommand) ReadResponse() (err error) {
	var response string
	if response, err = b.adapter.ReadLine(); err != nil {
		return fmt.Errorf("could not read response: %w", err)
	}

	switch {
	case strings.HasPrefix(response, codes.SuccessResponse):
		return nil
	case strings.HasPrefix(response, codes.AuthenticationInvalidPassword):
		return errors.New("invalid password")
	case strings.HasPrefix(response, codes.AuthenticationRequired):
		return errors.New("authentication is required")
	default:
		return fmt.Errorf("unexpected response from server: %s", response)
	}
}
