package infrastructure

import (
	"log"
	"log/slog"
	"nats-service/application/config"
	"nats-service/application/services"
	"nats-service/domain/entities"
	"nats-service/domain/interfaces"
	"nats-service/infrastructure/broker"
	"nats-service/infrastructure/grpc/handler"
	"nats-service/infrastructure/grpc/server"
	"nats-service/infrastructure/grpc/validators"
	"os"
	"shared/dependency"
	"shared/observability/nats-service/metrics"
	"time"

	"github.com/nats-io/nats.go"
)

// Container provides a lazily initialized set of dependencies.
type Container struct {
	Logger     dependency.LazyDependency[*slog.Logger]
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

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			var (
				file *os.File
				err  error
			)
			if err = os.MkdirAll(interfaces.LogDir, 0o755); err != nil {
				log.Fatalf("Failed to create log directory: %v", err)
			}

			file, err = os.OpenFile(interfaces.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
			}
			return slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{}))
		},
	}
	c.Config = dependency.LazyDependency[*config.Config]{
		InitFunc: config.GetConfig,
	}
	c.Metrics = dependency.LazyDependency[*metrics.Metrics]{
		InitFunc: metrics.NewMetrics,
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
					logger.Info("Reconnected to NATS", slog.String("url", conn.ConnectedUrl()))
					c.Metrics.Get().Connection.NATSConnectionStatus.Set(1)
				}
				disconnectErrHandler = func(conn *nats.Conn, err error) {
					logger.Error("Disconnected from NATS", slog.String("error", err.Error()))
					c.Metrics.Get().Connection.NATSConnectionStatus.Set(0)
				}
			)
			if address, err = entities.GetBroker().Address(); err != nil {
				logger.Error("Failed to get broker address", slog.String("error", err.Error()))
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
			return broker.NewClient(options, c.Metrics.Get(), logger)
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
				logger.Error("Failed to connect to NATS", slog.String("error", err.Error()))
				panic(err)
			}
			return services.NewOperations(conn, c.Metrics.Get(), logger)
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
	c.BusServer = dependency.LazyDependency[*server.BusServer]{
		InitFunc: func() *server.BusServer {
			var (
				logger    = c.Logger.Get()
				env       = c.Config.Get().Env
				port      = c.Config.Get().RPC.Port
				certFile  = c.Config.Get().TLS.Certificate
				keyFile   = c.Config.Get().TLS.Key
				err       error
				busServer *server.BusServer
			)
			if busServer, err = server.NewBusServer(env, port, certFile, keyFile, logger); err != nil {
				logger.Error("Failed to create BusServer", slog.String("error", err.Error()))
				panic(err)
			}
			return busServer
		},
	}

	return c
}
