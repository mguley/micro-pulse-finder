package control

import (
	"fmt"
	"proxy-service/infrastructure/proxy"
)

// AuthenticateCommand handles the authentication process with the proxy control port.
type AuthenticateCommand struct {
	BaseCommand
	password string // password is used to authenticate with the proxy control port.
}

// NewAuthenticateCommand creates a new instance of AuthenticateCommand.
func NewAuthenticateCommand(adapter *proxy.Connection, password string) *AuthenticateCommand {
	return &AuthenticateCommand{
		BaseCommand: BaseCommand{adapter: adapter},
		password:    password,
	}
}

// Execute performs the authentication process by sending the AUTHENTICATE command and reading the response
// from the proxy control port.
func (c *AuthenticateCommand) Execute() (err error) {
	if err = c.Initialize(); err != nil {
		return fmt.Errorf("%w", err)
	}

	command := fmt.Sprintf("AUTHENTICATE %q\n", c.password)
	if err = c.SendCommand(command); err != nil {
		return fmt.Errorf("could not send AUTHENTICATE command: %w", err)
	}

	return c.ReadResponse()
}
