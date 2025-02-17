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
			RpcHost: cfg.Nats.RpcHost,
			RpcPort: cfg.Nats.RpcPort,
		}
	})
	return nats
}

// Nats represents NATS configuration details.
type Nats struct {
	RpcHost string // RpcHost is the address of the NATS gRPC server.
	RpcPort string // RpcPort is the port number of the NATS gRPC server.
}

// Address returns the full address of the NATS server.
func (b *Nats) Address() (address string, err error) {
	if b.RpcHost == "" || b.RpcPort == "" {
		return "", fmt.Errorf("invalid address: host or port is empty")
	}
	return fmt.Sprintf("%s:%s", b.RpcHost, b.RpcPort), nil
}
