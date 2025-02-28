package url

import (
	"log/slog"
	"os"
	"shared/dependency"
	"shared/mongodb/application/config"
	"shared/mongodb/domain/entities"
	"shared/mongodb/infrastructure/mongodb"
	"url-service/domain/interfaces"
	"url-service/infrastructure/url"

	"go.mongodb.org/mongo-driver/mongo"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Logger          dependency.LazyDependency[*slog.Logger]
	MongoClient     dependency.LazyDependency[*mongodb.Client]
	MongoRepository dependency.LazyDependency[interfaces.UrlRepository]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		},
	}
	c.MongoClient = dependency.LazyDependency[*mongodb.Client]{
		InitFunc: func() *mongodb.Client {
			var (
				address string
				err     error
			)
			if address, err = entities.GetMongo().Address(); err != nil {
				panic(err)
			}
			return mongodb.NewClient(address)
		},
	}
	c.MongoRepository = dependency.LazyDependency[interfaces.UrlRepository]{
		InitFunc: func() interfaces.UrlRepository {
			var (
				logger         = c.Logger.Get()
				mongoClient    *mongo.Client
				collection     *mongo.Collection
				collectionName = config.GetConfig().Mongo.Collection
				dbName         = config.GetConfig().Mongo.DB
				err            error
			)
			if mongoClient, err = c.MongoClient.Get().Connect(); err != nil {
				panic(err)
			}
			collection = mongoClient.Database(dbName).Collection(collectionName)
			return url.NewRepository(mongoClient, collection, logger)
		},
	}

	return c
}
