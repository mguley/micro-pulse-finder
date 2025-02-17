package application

import (
	"proxy-service/application/commands"
	"proxy-service/application/commands/control"
	"proxy-service/application/config"
	"proxy-service/application/services"
	"proxy-service/domain/entities"
	"proxy-service/domain/interfaces"
	"proxy-service/infrastructure"
	"shared/dependency"
	"shared/grpc/clients/nats_service"
	"time"
)

// Container provides a lazily initialized set of dependencies.
type Container struct {
	Config              dependency.LazyDependency[*config.Config]
	Infrastructure      dependency.LazyDependency[*infrastructure.Container]
	AuthenticateCommand dependency.LazyDependency[*control.AuthenticateCommand]
	SignalCommand       dependency.LazyDependency[*control.SignalCommand]
	StatusCommand       dependency.LazyDependency[*commands.StatusCommand]
	RetryStrategy       dependency.LazyDependency[interfaces.RetryStrategy]
	NatsGrpcValidator   dependency.LazyDependency[nats_service.Validator]
	NatsGrpcClient      dependency.LazyDependency[*nats_service.NatsClient]
	UrlProcessorService dependency.LazyDependency[*services.UrlProcessorService]
}

// NewContainer initializes and returns a new Container with dependencies.
func NewContainer() *Container {
	c := &Container{}

	c.Config = dependency.LazyDependency[*config.Config]{
		InitFunc: config.GetConfig,
	}
	c.Infrastructure = dependency.LazyDependency[*infrastructure.Container]{
		InitFunc: infrastructure.NewContainer,
	}
	c.RetryStrategy = dependency.LazyDependency[interfaces.RetryStrategy]{
		InitFunc: func() interfaces.RetryStrategy {
			var (
				baseDelay  = time.Duration(5) * time.Second
				maxDelay   = time.Duration(45) * time.Second
				attempts   = 5
				multiplier = 2.0
			)
			return services.NewExponentialBackoffStrategy(baseDelay, maxDelay, attempts, multiplier)
		},
	}
	c.NatsGrpcValidator = dependency.LazyDependency[nats_service.Validator]{
		InitFunc: func() nats_service.Validator {
			return nats_service.NewBusClientValidator()
		},
	}
	c.NatsGrpcClient = dependency.LazyDependency[*nats_service.NatsClient]{
		InitFunc: func() *nats_service.NatsClient {
			var (
				env        = c.Config.Get().Env
				validator  = c.NatsGrpcValidator.Get()
				natsClient *nats_service.NatsClient
				address    string
				err        error
			)
			if address, err = entities.GetNats().Address(); err != nil {
				panic(err)
			}
			if natsClient, err = nats_service.NewNatsClient(env, address, validator); err != nil {
				panic(err)
			}
			return natsClient
		},
	}
	c.UrlProcessorService = dependency.LazyDependency[*services.UrlProcessorService]{
		InitFunc: func() *services.UrlProcessorService {
			var (
				pool       = c.Infrastructure.Get().ConnectionPool.Get()
				natsClient = c.NatsGrpcClient.Get()
				batchSize  = c.Config.Get().UrlProcessor.BatchSize
				queueGroup = c.Config.Get().UrlProcessor.QueueGroup
			)
			return services.NewUrlProcessorService(pool, natsClient, batchSize, queueGroup)
		},
	}

	// Proxy commands
	c.AuthenticateCommand = dependency.LazyDependency[*control.AuthenticateCommand]{
		InitFunc: func() *control.AuthenticateCommand {
			var (
				adapter  = c.Infrastructure.Get().PortConnection.Get()
				password = c.Config.Get().Proxy.ControlPassword
			)
			return control.NewAuthenticateCommand(adapter, password)
		},
	}
	c.SignalCommand = dependency.LazyDependency[*control.SignalCommand]{
		InitFunc: func() *control.SignalCommand {
			var (
				adapter = c.Infrastructure.Get().PortConnection.Get()
				signal  = "NEWNYM"
			)
			return control.NewSignalCommand(adapter, signal)
		},
	}
	c.StatusCommand = dependency.LazyDependency[*commands.StatusCommand]{
		InitFunc: func() *commands.StatusCommand {
			var (
				timeout = time.Duration(10) * time.Second
				url     = c.Config.Get().Proxy.Url
				client  = c.Infrastructure.Get().Socks5Client.Get()
			)
			return commands.NewStatusCommand(timeout, url, client)
		},
	}

	return c
}
