package application

import (
	"shared/dependency"
	"shared/grpc/clients/nats_service"
	"time"
	"url-service/application/config"
	"url-service/application/services/messages"
	"url-service/domain/entities"
	"url-service/infrastructure"
)

// Container provides a lazily initialized set of dependencies.
type Container struct {
	Config                 dependency.LazyDependency[*config.Config]
	Infrastructure         dependency.LazyDependency[*infrastructure.Container]
	NatsGrpcValidator      dependency.LazyDependency[nats_service.Validator]
	NatsGrpcClient         dependency.LazyDependency[*nats_service.NatsClient]
	InboundMessageService  dependency.LazyDependency[*messages.InboundMessageService]
	OutboundMessageService dependency.LazyDependency[*messages.OutboundMessageService]
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
	c.InboundMessageService = dependency.LazyDependency[*messages.InboundMessageService]{
		InitFunc: func() *messages.InboundMessageService {
			var (
				logger        = c.Infrastructure.Get().Logger.Get()
				natsClient    = c.NatsGrpcClient.Get()
				urlRepository = c.Infrastructure.Get().MongoRepository.Get()
				batchSize     = c.Config.Get().InboundMessage.BatchSize
				queueGroup    = c.Config.Get().InboundMessage.QueueGroup
			)
			return messages.NewInboundMessageService(natsClient, urlRepository, batchSize, queueGroup, logger)
		},
	}
	c.OutboundMessageService = dependency.LazyDependency[*messages.OutboundMessageService]{
		InitFunc: func() *messages.OutboundMessageService {
			var (
				logger        = c.Infrastructure.Get().Logger.Get()
				natsClient    = c.NatsGrpcClient.Get()
				urlRepository = c.Infrastructure.Get().MongoRepository.Get()
				interval      = time.Duration(5) * time.Minute
				batchSize     = c.Config.Get().OutboundMessage.BatchSize
			)
			return messages.NewOutboundMessageService(natsClient, urlRepository, interval, batchSize, logger)
		},
	}

	return c
}
