package control

import (
	"fmt"
	"proxy-service/infrastructure/proxy"
)

// SignalCommand handles sending signals (e.g., "NEWNYM") to the proxy control port.
type SignalCommand struct {
	BaseCommand
	signal string // signal is the signal to be sent (e.g., "NEWNYM").
}

// NewSignalCommand creates a new instance of SignalCommand.
func NewSignalCommand(adapter *proxy.Connection, signal string) *SignalCommand {
	return &SignalCommand{
		BaseCommand: BaseCommand{adapter: adapter},
		signal:      signal,
	}
}

// Execute sends the signal command to the proxy control port.
func (c *SignalCommand) Execute() (err error) {
	if err = c.Initialize(); err != nil {
		return fmt.Errorf("%w", err)
	}

	command := fmt.Sprintf("SIGNAL %s\r\n", c.signal)
	if err = c.SendCommand(command); err != nil {
		return fmt.Errorf("could not send SIGNAL command: %w", err)
	}

	return c.ReadResponse()
}
