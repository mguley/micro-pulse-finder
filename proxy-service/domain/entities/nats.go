package entities

import (
	"fmt"
	"proxy-service/application/config"
	"sync"
)

var (
	nats     *Nats
	natsOnce sync.Once
)

// GetNats retrieves the NATS configuration.
func GetNats() *Nats {
	natsOnce.Do(func() {
		cfg := config.GetConfig()
		nats = &Nats{
			Host: cfg.Nats.Host,
			Port: cfg.Nats.Port,
		}
	})
	return nats
}

// Nats represents NATS configuration details.
type Nats struct {
	Host string // Host is the hostname of the NATS server.
	Port string // Port is the port number of the NATS server.
}

// Address returns the full address of the NATS server.
func (b *Nats) Address() (address string, err error) {
	if b.Host == "" || b.Port == "" {
		return "", fmt.Errorf("invalid address: host or port is empty")
	}
	return fmt.Sprintf("%s:%s", b.Host, b.Port), nil
}
