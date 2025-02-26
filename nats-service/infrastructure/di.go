package infrastructure

import (
	"log"
	"nats-service/application/config"
	"nats-service/application/services"
	"nats-service/domain/entities"
	"nats-service/infrastructure/broker"
	"nats-service/infrastructure/grpc/handler"
	"nats-service/infrastructure/grpc/server"
	"nats-service/infrastructure/grpc/validators"
	"shared/dependency"
	"shared/observability/nats-service/metrics"
	"time"

	"github.com/nats-io/nats.go"
)

// Container provides a lazily initialized set of dependencies.
type Container struct {
	Config     dependency.LazyDependency[*config.Config]
	Metrics    dependency.LazyDependency[*metrics.Metrics]
	NatsClient dependency.LazyDependency[*broker.Client]
	Operations dependency.LazyDependency[*services.Operations]
	Validator  dependency.LazyDependency[validators.Validator]
	BusService dependency.LazyDependency[*handler.BusService]
	BusServer  dependency.LazyDependency[*server.BusServer]
}

// NewContainer initializes and returns a new Container with dependencies.
func NewContainer() *Container {
	c := &Container{}

	c.Config = dependency.LazyDependency[*config.Config]{
		InitFunc: config.GetConfig,
	}
	c.Metrics = dependency.LazyDependency[*metrics.Metrics]{
		InitFunc: metrics.NewMetrics,
	}
	c.NatsClient = dependency.LazyDependency[*broker.Client]{
		InitFunc: func() *broker.Client {
			var (
				address          string
				err              error
				reconnectWait    = time.Duration(5) * time.Second
				maxReconnect     = 5
				timeout          = time.Duration(5) * time.Second
				reconnectHandler = func(conn *nats.Conn) {
					c.Metrics.Get().Connection.NATSConnectionStatus.Set(1)
					log.Println("Reconnected to NATS at", conn.ConnectedUrl())
				}
				disconnectErrHandler = func(conn *nats.Conn, err error) {
					c.Metrics.Get().Connection.NATSConnectionStatus.Set(0)
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
			return broker.NewClient(options, c.Metrics.Get())
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
			return services.NewOperations(conn, c.Metrics.Get())
		},
	}
	c.Validator = dependency.LazyDependency[validators.Validator]{
		InitFunc: func() validators.Validator {
			return validators.NewBusValidator()
		},
	}
	c.BusService = dependency.LazyDependency[*handler.BusService]{
		InitFunc: func() *handler.BusService {
			return handler.NewBusService(c.Operations.Get(), c.Validator.Get())
		},
	}
	c.BusServer = dependency.LazyDependency[*server.BusServer]{
		InitFunc: func() *server.BusServer {
			var (
				env       = c.Config.Get().Env
				port      = c.Config.Get().RPC.Port
				certFile  = c.Config.Get().TLS.Certificate
				keyFile   = c.Config.Get().TLS.Key
				err       error
				busServer *server.BusServer
			)
			if busServer, err = server.NewBusServer(env, port, certFile, keyFile); err != nil {
				panic(err)
			}
			return busServer
		},
	}

	return c
}
