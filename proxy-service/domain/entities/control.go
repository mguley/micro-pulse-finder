package entities

import (
	"fmt"
	"proxy-service/application/config"
	"sync"
)

var (
	port     *ControlPort
	portOnce sync.Once
)

// GetControlPort retrieves the ControlPort configuration.
func GetControlPort() *ControlPort {
	portOnce.Do(func() {
		cfg := config.GetConfig()
		port = &ControlPort{
			Host:     cfg.Proxy.Host,
			Port:     cfg.Proxy.ControlPort,
			Password: cfg.Proxy.ControlPassword,
		}
	})
	return port
}

// ControlPort represents the details of a proxy control port configuration.
type ControlPort struct {
	Host     string // Host is the hostname of the proxy server.
	Port     string // Port is the port number of the proxy control port.
	Password string // Password is a password used to authenticate with the proxy control port.
}

// Address returns the full address of the proxy control port.
func (c *ControlPort) Address() (uri string, err error) {
	if c.Host == "" || c.Port == "" {
		return "", fmt.Errorf("invalid address: host or port is empty")
	}
	return fmt.Sprintf("%s:%s", c.Host, c.Port), nil
}
