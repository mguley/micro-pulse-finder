package control

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"proxy-service/application/commands/control/codes"
	"proxy-service/infrastructure/proxy"
	"strings"
)

// BaseCommand provides common functionality for commands interacting with the proxy control port.
type BaseCommand struct {
	adapter *proxy.Connection // adapter is a common adapter for proxy commands.
	conn    net.Conn          // conn is the underlying TCP connection.
	logger  *slog.Logger      // logger for structured logging.
}

// Initialize establishes a connection to the proxy control port.
func (b *BaseCommand) Initialize() (err error) {
	if b.adapter == nil {
		b.logger.Error("Connection adapter is not initialized")
		return errors.New("connection adapter is not initialized")
	}

	b.logger.Info("Attempting to establish connection to proxy")
	if b.conn, err = b.adapter.Dial(); err != nil {
		b.logger.Error("Could not connect to proxy", "error", err)
		return fmt.Errorf("could not connect to proxy: %w", err)
	}

	b.logger.Info("Connection established")
	return nil
}

// Close closes the established connection.
func (b *BaseCommand) Close() (err error) {
	if err = b.adapter.Close(); err != nil {
		b.logger.Error("Could not close connection", "error", err)
	}
	return err
}

// SendCommand sends a formatted command to the proxy server.
func (b *BaseCommand) SendCommand(command string) (err error) {
	if !b.adapter.IsConnected() {
		b.logger.Error("Connection is not established")
		return errors.New("connection is not established")
	}

	var (
		writer       *bufio.Writer
		bytesWritten int
	)

	writer = bufio.NewWriter(b.conn)
	if bytesWritten, err = writer.WriteString(command); err != nil {
		b.logger.Error("Could not send command to proxy", "error", err)
		return fmt.Errorf("could not send command to proxy: %w", err)
	}

	if err = writer.Flush(); err != nil {
		b.logger.Error("Could not flush writer", "error", err)
		return fmt.Errorf("could not flush writer: %w", err)
	}

	b.logger.Info("Command sent", "command", command, "bytesWritten", bytesWritten)
	return nil
}

// ReadResponse reads and processes the response from the proxy server.
func (b *BaseCommand) ReadResponse() (err error) {
	var response string
	if response, err = b.adapter.ReadLine(); err != nil {
		b.logger.Error("Could not read response", "error", err)
		return fmt.Errorf("could not read response: %w", err)
	}

	switch {
	case strings.HasPrefix(response, codes.SuccessResponse):
		b.logger.Info("Successful response received")
		return nil
	case strings.HasPrefix(response, codes.AuthenticationInvalidPassword):
		b.logger.Error("Invalid password provided")
		return errors.New("invalid password")
	case strings.HasPrefix(response, codes.AuthenticationRequired):
		b.logger.Error("Authentication is required")
		return errors.New("authentication is required")
	default:
		b.logger.Error("Unexpected response from server", "response", response)
		return fmt.Errorf("unexpected response from server: %s", response)
	}
}
