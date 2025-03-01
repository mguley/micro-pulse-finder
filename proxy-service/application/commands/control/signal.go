package control

import (
	"fmt"
	"log/slog"
	"proxy-service/infrastructure/proxy"
)

// SignalCommand handles sending signals (e.g., "NEWNYM") to the proxy control port.
type SignalCommand struct {
	BaseCommand
	signal string // signal is the signal to be sent (e.g., "NEWNYM").
}

// NewSignalCommand creates a new instance of SignalCommand.
func NewSignalCommand(adapter *proxy.Connection, signal string, logger *slog.Logger) *SignalCommand {
	return &SignalCommand{
		BaseCommand: BaseCommand{adapter: adapter, logger: logger},
		signal:      signal,
	}
}

// Execute sends the signal command to the proxy control port.
func (c *SignalCommand) Execute() (err error) {
	c.logger.Info("Starting signal command", "signal", c.signal)
	if err = c.Initialize(); err != nil {
		c.logger.Error("Failed to initialize signal command", "error", err)
		return fmt.Errorf("%w", err)
	}

	command := fmt.Sprintf("SIGNAL %s\r\n", c.signal)
	c.logger.Info("Sending SIGNAL command", "command", command)
	if err = c.SendCommand(command); err != nil {
		c.logger.Error("Could not send SIGNAL command", "error", err)
		return fmt.Errorf("could not send SIGNAL command: %w", err)
	}

	if err = c.ReadResponse(); err != nil {
		c.logger.Error("Error reading response for SIGNAL command", "error", err)
	}
	return err
}
