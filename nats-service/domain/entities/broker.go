package entities

import (
	"fmt"
	"nats-service/application/config"
	"sync"
)

var (
	broker     *Nats
	brokerOnce sync.Once
)

// GetBroker retrieves the Nats configuration.
func GetBroker() *Nats {
	brokerOnce.Do(func() {
		cfg := config.GetConfig()
		broker = &Nats{
			Host: cfg.Nats.Host,
			Port: cfg.Nats.Port,
		}
	})
	return broker
}

// Nats represents the details of a message broker configuration.
type Nats struct {
	Host string // Host is the hostname of the NATS server.
	Port string // Port is the port number of the NATS server.
}

// Address returns the full address of the message broker.
func (b *Nats) Address() (uri string, err error) {
	if b.Host == "" || b.Port == "" {
		return "", fmt.Errorf("invalid address: host or port is empty")
	}
	return fmt.Sprintf("nats://%s:%s", b.Host, b.Port), nil
}
