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
	"nats-service/infrastructure/metrics"
	"os"
	"shared/dependency"
	"time"

	"github.com/nats-io/nats.go"
)

// Container provides a lazily initialized set of infrastructure dependencies.
//
// Fields:
//   - Logger:     Lazy dependency for the logger instance.
//   - Config:     Lazy dependency for the application configuration.
//   - NatsClient: Lazy dependency for the NATS client.
//   - Operations: Lazy dependency for the NATS operations service.
//   - Validator:  Lazy dependency for the request validator.
//   - BusService: Lazy dependency for the gRPC bus service.
//   - BusServer:  Lazy dependency for the gRPC bus server.
type Container struct {
	Logger          dependency.LazyDependency[*slog.Logger]
	Config          dependency.LazyDependency[*config.Config]
	NatsClient      dependency.LazyDependency[*broker.Client]
	Operations      dependency.LazyDependency[*services.Operations]
	Validator       dependency.LazyDependency[validators.Validator]
	BusService      dependency.LazyDependency[*handler.BusService]
	BusServer       dependency.LazyDependency[*server.BusServer]
	MetricsServer   dependency.LazyDependency[*metrics.Server]
	MetricsProvider dependency.LazyDependency[*metrics.Provider]
}

// NewContainer initializes and returns a new Container with all required dependencies.
//
// Returns:
//   - *Container: A pointer to the newly created infrastructure container.
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
				}
				disconnectErrHandler = func(conn *nats.Conn, err error) {
					logger.Error("Disconnected from NATS", slog.String("error", err.Error()))
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
				logger.Error("Failed to connect to NATS", slog.String("error", err.Error()))
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

	// Metrics section
	c.MetricsServer = dependency.LazyDependency[*metrics.Server]{
		InitFunc: func() *metrics.Server {
			var (
				port            = c.Config.Get().Metrics.ServerPort
				metricsProvider = c.MetricsProvider.Get()
				logger          = c.Logger.Get()
			)
			return metrics.NewServer(port, metricsProvider, logger)
		},
	}
	c.MetricsProvider = dependency.LazyDependency[*metrics.Provider]{
		InitFunc: func() *metrics.Provider {
			var (
				namespace = "nats_service"
				logger    = c.Logger.Get()
			)
			return metrics.NewProvider(namespace, logger)
		},
	}

	return c
}
