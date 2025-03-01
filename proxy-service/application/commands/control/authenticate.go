package control

import (
	"fmt"
	"log/slog"
	"proxy-service/infrastructure/proxy"
)

// AuthenticateCommand handles the authentication process with the proxy control port.
type AuthenticateCommand struct {
	BaseCommand
	password string // password is used to authenticate with the proxy control port.
}

// NewAuthenticateCommand creates a new instance of AuthenticateCommand.
func NewAuthenticateCommand(adapter *proxy.Connection, password string, logger *slog.Logger) *AuthenticateCommand {
	return &AuthenticateCommand{
		BaseCommand: BaseCommand{adapter: adapter, logger: logger},
		password:    password,
	}
}

// Execute performs the authentication process by sending the AUTHENTICATE command and reading the response
// from the proxy control port.
func (c *AuthenticateCommand) Execute() (err error) {
	c.logger.Info("Starting authentication command", "command", "AUTHENTICATE")
	if err = c.Initialize(); err != nil {
		c.logger.Error("Failed to initialize authentication command", "error", err)
		return fmt.Errorf("%w", err)
	}

	command := fmt.Sprintf("AUTHENTICATE %q\n", c.password)
	c.logger.Info("Sending AUTHENTICATE command", "command", command)
	if err = c.SendCommand(command); err != nil {
		c.logger.Error("Could not send AUTHENTICATE command", "error", err)
		return fmt.Errorf("could not send AUTHENTICATE command: %w", err)
	}

	if err = c.ReadResponse(); err != nil {
		c.logger.Error("Error reading response for AUTHENTICATE command", "error", err)
	}
	return err
}
