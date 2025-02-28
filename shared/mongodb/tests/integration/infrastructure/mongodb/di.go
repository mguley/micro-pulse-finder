package mongodb

import (
	"log/slog"
	"os"
	"shared/dependency"
	"shared/mongodb/domain/entities"
	"shared/mongodb/infrastructure/mongodb"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Logger      dependency.LazyDependency[*slog.Logger]
	MongoClient dependency.LazyDependency[*mongodb.Client]
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
				logger  = c.Logger.Get()
				address string
				err     error
			)
			if address, err = entities.GetMongo().Address(); err != nil {
				panic(err)
			}
			return mongodb.NewClient(address, logger)
		},
	}

	return c
}
