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

// GetBroker retrieves the NATS broker configuration.
//
// Returns:
//   - *Nats: A pointer to the Nats struct containing the broker details.
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

// Nats represents the configuration details for a NATS message broker.
//
// Fields:
//   - Host: The hostname of the NATS server.
//   - Port: The port number of the NATS server.
type Nats struct {
	Host string
	Port string
}

// Address constructs and returns the full URI address for the NATS server.
//
// Returns:
//   - uri: A formatted URI string in the form "nats://<host>:<port>".
//   - err: An error if either the host or port is missing.
func (b *Nats) Address() (uri string, err error) {
	if b.Host == "" || b.Port == "" {
		return "", fmt.Errorf("invalid address: host or port is empty")
	}
	return fmt.Sprintf("nats://%s:%s", b.Host, b.Port), nil
}
