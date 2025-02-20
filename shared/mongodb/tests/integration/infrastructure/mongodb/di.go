package mongodb

import (
	"shared/dependency"
	"shared/mongodb/domain/entities"
	"shared/mongodb/infrastructure/mongodb"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	MongoClient dependency.LazyDependency[*mongodb.Client]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

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

	return c
}
