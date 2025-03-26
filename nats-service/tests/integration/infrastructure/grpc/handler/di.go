package handler

import (
	"log"
	"log/slog"
	"nats-service/application/services"
	"nats-service/domain/entities"
	"nats-service/infrastructure/broker"
	"nats-service/infrastructure/grpc/handler"
	"nats-service/infrastructure/grpc/validators"
	"os"
	"shared/dependency"
	"time"

	"github.com/nats-io/nats.go"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Logger     dependency.LazyDependency[*slog.Logger]
	NatsClient dependency.LazyDependency[*broker.Client]
	Operations dependency.LazyDependency[*services.Operations]
	Validator  dependency.LazyDependency[validators.Validator]
	BusService dependency.LazyDependency[*handler.BusService]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		},
	}
	c.NatsClient = dependency.LazyDependency[*broker.Client]{
		InitFunc: func() *broker.Client {
			var (
				logger           = c.Logger.Get()
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
			return broker.NewClient(options, logger)
		},
	}
	c.Operations = dependency.LazyDependency[*services.Operations]{
		InitFunc: func() *services.Operations {
			var (
				logger = c.Logger.Get()
				conn   *nats.Conn
				err    error
			)
			if conn, err = c.NatsClient.Get().Connect(); err != nil {
				panic(err)
			}
			return services.NewOperations(conn, logger)
		},
	}
	c.Validator = dependency.LazyDependency[validators.Validator]{
		InitFunc: func() validators.Validator {
			return validators.NewBusValidator()
		},
	}
	c.BusService = dependency.LazyDependency[*handler.BusService]{
		InitFunc: func() *handler.BusService {
			return handler.NewBusService(c.Operations.Get(), c.Validator.Get(), c.Logger.Get())
		},
	}

	return c
}
