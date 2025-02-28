package infrastructure

import (
	"log"
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

// Container provides a lazily initialized set of dependencies.
type Container struct {
	Logger          dependency.LazyDependency[*slog.Logger]
	MongoClient     dependency.LazyDependency[*mongodb.Client]
	MongoRepository dependency.LazyDependency[interfaces.UrlRepository]
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

			file, err = os.OpenFile(interfaces.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
			}
			return slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{}))
		},
	}
	c.MongoClient = dependency.LazyDependency[*mongodb.Client]{
		InitFunc: func() *mongodb.Client {
			var (
				logger  = c.Logger.Get()
				address string
				err     error
			)
			if address, err = entities.GetMongo().Address(); err != nil {
				logger.Error("Failed to get mongo address", "error", err)
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
				logger.Error("Failed to connect to MongoDB", "error", err)
				panic(err)
			}
			collection = mongoClient.Database(dbName).Collection(collectionName)
			return url.NewRepository(mongoClient, collection, logger)
		},
	}

	return c
}
