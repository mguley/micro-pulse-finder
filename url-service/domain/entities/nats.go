package entities

import (
	"fmt"
	"sync"
	"url-service/application/config"
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
func (n *Nats) Address() (address string, err error) {
	if n.RpcHost == "" || n.RpcPort == "" {
		return "", fmt.Errorf("invalid address: host or port is empty")
	}
	return fmt.Sprintf("%s:%s", n.RpcHost, n.RpcPort), nil
}
