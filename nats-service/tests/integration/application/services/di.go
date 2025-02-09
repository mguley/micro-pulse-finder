package services

import (
	"log"
	"nats-service/application/services"
	"nats-service/domain/entities"
	"nats-service/infrastructure/broker"
	"shared/dependency"
	"time"

	"github.com/nats-io/nats.go"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	NatsClient dependency.LazyDependency[*broker.Client]
	Operations dependency.LazyDependency[*services.Operations]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.NatsClient = dependency.LazyDependency[*broker.Client]{
		InitFunc: func() *broker.Client {
			var (
				address          string
				err              error
				reconnectWait    = time.Duration(5) * time.Second
				maxReconnect     = 5
				timeout          = time.Duration(5) * time.Second
				reconnectHandler = func(conn *nats.Conn) {
					log.Println("Reconnected to NATS at", conn.ConnectedUrl())
				}
				disconnectErrHandler = func(conn *nats.Conn, err error) {
					log.Println("Disconnected from NATS due to", err)
				}
			)
			if address, err = entities.GetBroker().Address(); err != nil {
				panic(err)
			}

			options := &nats.Options{
				Url:               address,
				ReconnectWait:     reconnectWait,
				MaxReconnect:      maxReconnect,
				Timeout:           timeout,
				ReconnectedCB:     reconnectHandler,
				DisconnectedErrCB: disconnectErrHandler,
				AllowReconnect:    true,
			}
			return broker.NewClient(options)
		},
	}
	c.Operations = dependency.LazyDependency[*services.Operations]{
		InitFunc: func() *services.Operations {
			var (
				conn *nats.Conn
				err  error
			)
			if conn, err = c.NatsClient.Get().Connect(); err != nil {
				panic(err)
			}
			return services.NewOperations(conn)
		},
	}

	return c
}
