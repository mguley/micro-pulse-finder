package processor

import (
	"log/slog"
	natsServiceInfrastructure "nats-service/infrastructure"
	"os"
	"proxy-service/application/config"
	"proxy-service/application/services"
	"proxy-service/domain/entities"
	"proxy-service/domain/interfaces"
	"proxy-service/infrastructure/http/socks5"
	"proxy-service/infrastructure/http/socks5/agent"
	"shared/dependency"
	"shared/grpc/clients/nats_service"
	"time"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Logger                    dependency.LazyDependency[*slog.Logger]
	Config                    dependency.LazyDependency[*config.Config]
	UserAgent                 dependency.LazyDependency[interfaces.Agent]
	Socks5Client              dependency.LazyDependency[*socks5.Client]
	ConnectionPool            dependency.LazyDependency[*socks5.ConnectionPool]
	NatsGrpcValidator         dependency.LazyDependency[nats_service.Validator]
	NatsGrpcClient            dependency.LazyDependency[*nats_service.NatsClient]
	UrlProcessorService       dependency.LazyDependency[*services.UrlProcessorService]
	NatsServiceInfrastructure dependency.LazyDependency[*natsServiceInfrastructure.Container]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		},
	}
	c.Config = dependency.LazyDependency[*config.Config]{
		InitFunc: config.GetConfig,
	}
	c.UserAgent = dependency.LazyDependency[interfaces.Agent]{
		InitFunc: func() interfaces.Agent {
			return agent.NewChromeAgent(c.Logger.Get())
		},
	}
	c.Socks5Client = dependency.LazyDependency[*socks5.Client]{
		InitFunc: func() *socks5.Client {
			var (
				logger    = c.Logger.Get()
				userAgent = c.UserAgent.Get()
				timeout   = time.Duration(10) * time.Second
			)
			return socks5.NewClient(userAgent, timeout, logger)
		},
	}
	c.ConnectionPool = dependency.LazyDependency[*socks5.ConnectionPool]{
		InitFunc: func() *socks5.ConnectionPool {
			var (
				logger          = c.Logger.Get()
				poolSize        = c.Config.Get().Pool.MaxSize
				refreshInterval = c.Config.Get().Pool.RefreshInterval
				creator         = c.Socks5Client.Get().Create
			)
			return socks5.NewConnectionPool(poolSize, time.Duration(refreshInterval)*time.Second, creator, logger)
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
				logger     = c.Logger.Get()
				env        = c.Config.Get().Env
				validator  = c.NatsGrpcValidator.Get()
				natsClient *nats_service.NatsClient
				address    string
				err        error
			)
			if address, err = entities.GetNats().Address(); err != nil {
				panic(err)
			}
			if natsClient, err = nats_service.NewNatsClient(env, address, validator, logger); err != nil {
				panic(err)
			}
			return natsClient
		},
	}
	c.UrlProcessorService = dependency.LazyDependency[*services.UrlProcessorService]{
		InitFunc: func() *services.UrlProcessorService {
			var (
				logger     = c.Logger.Get()
				pool       = c.ConnectionPool.Get()
				natsClient = c.NatsGrpcClient.Get()
				batchSize  = c.Config.Get().UrlProcessor.BatchSize
				queueGroup = c.Config.Get().UrlProcessor.QueueGroup
			)
			return services.NewUrlProcessorService(pool, natsClient, batchSize, queueGroup, logger)
		},
	}

	// Inject the full nats-service infrastructure container (includes BusServer and BusService).
	c.NatsServiceInfrastructure = dependency.LazyDependency[*natsServiceInfrastructure.Container]{
		InitFunc: natsServiceInfrastructure.NewContainer,
	}

	return c
}
