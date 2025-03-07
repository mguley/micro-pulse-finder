package messages

import (
	"log/slog"
	natsServiceInfrastructure "nats-service/infrastructure"
	"os"
	"shared/dependency"
	"shared/grpc/clients/nats_service"
	sharedConfig "shared/mongodb/application/config"
	sharedDomain "shared/mongodb/domain/entities"
	"shared/mongodb/infrastructure/mongodb"
	"time"
	urlServiceConfig "url-service/application/config"
	"url-service/application/services/messages"
	urlServiceDomain "url-service/domain/entities"
	"url-service/domain/interfaces"
	"url-service/infrastructure/url"

	"go.mongodb.org/mongo-driver/mongo"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Logger                    dependency.LazyDependency[*slog.Logger]
	Config                    dependency.LazyDependency[*urlServiceConfig.Config]
	MongoClient               dependency.LazyDependency[*mongodb.Client]
	MongoRepository           dependency.LazyDependency[interfaces.UrlRepository]
	NatsGrpcValidator         dependency.LazyDependency[nats_service.Validator]
	NatsGrpcClient            dependency.LazyDependency[*nats_service.NatsClient]
	InboundMessageService     dependency.LazyDependency[*messages.InboundMessageService]
	OutboundMessageService    dependency.LazyDependency[*messages.OutboundMessageService]
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
	c.Config = dependency.LazyDependency[*urlServiceConfig.Config]{
		InitFunc: urlServiceConfig.GetConfig,
	}
	c.MongoClient = dependency.LazyDependency[*mongodb.Client]{
		InitFunc: func() *mongodb.Client {
			var (
				logger  = c.Logger.Get()
				address string
				err     error
			)
			if address, err = sharedDomain.GetMongo().Address(); err != nil {
				panic(err)
			}
			return mongodb.NewClient(address, logger)
		},
	}
	c.MongoRepository = dependency.LazyDependency[interfaces.UrlRepository]{
		InitFunc: func() interfaces.UrlRepository {
			var (
				logger         = c.Logger.Get()
				mongoClient    *mongo.Client
				collection     *mongo.Collection
				collectionName = sharedConfig.GetConfig().Mongo.Collection
				dbName         = sharedConfig.GetConfig().Mongo.DB
				err            error
			)
			if mongoClient, err = c.MongoClient.Get().Connect(); err != nil {
				panic(err)
			}
			collection = mongoClient.Database(dbName).Collection(collectionName)
			return url.NewRepository(mongoClient, collection, logger)
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
			if address, err = urlServiceDomain.GetNats().Address(); err != nil {
				panic(err)
			}
			if natsClient, err = nats_service.NewNatsClient(env, address, validator, logger); err != nil {
				panic(err)
			}
			return natsClient
		},
	}
	c.InboundMessageService = dependency.LazyDependency[*messages.InboundMessageService]{
		InitFunc: func() *messages.InboundMessageService {
			var (
				logger        = c.Logger.Get()
				natsClient    = c.NatsGrpcClient.Get()
				urlRepository = c.MongoRepository.Get()
				batchSize     = c.Config.Get().InboundMessage.BatchSize
				queueGroup    = c.Config.Get().InboundMessage.QueueGroup
			)
			return messages.NewInboundMessageService(natsClient, urlRepository, batchSize, queueGroup, logger)
		},
	}
	c.OutboundMessageService = dependency.LazyDependency[*messages.OutboundMessageService]{
		InitFunc: func() *messages.OutboundMessageService {
			var (
				logger        = c.Logger.Get()
				natsClient    = c.NatsGrpcClient.Get()
				urlRepository = c.MongoRepository.Get()
				interval      = time.Duration(5) * time.Second
				batchSize     = c.Config.Get().OutboundMessage.BatchSize
			)
			return messages.NewOutboundMessageService(natsClient, urlRepository, interval, batchSize, logger)
		},
	}

	// Inject the full nats-service infrastructure container (includes BusServer and BusService).
	c.NatsServiceInfrastructure = dependency.LazyDependency[*natsServiceInfrastructure.Container]{
		InitFunc: natsServiceInfrastructure.NewContainer,
	}

	return c
}
