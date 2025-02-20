package infrastructure

import (
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
	MongoClient     dependency.LazyDependency[*mongodb.Client]
	MongoRepository dependency.LazyDependency[interfaces.UrlRepository]
}

// NewContainer initializes and returns a new Container with dependencies.
func NewContainer() *Container {
	c := &Container{}

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
			return url.NewRepository(mongoClient, collection)
		},
	}

	return c
}
