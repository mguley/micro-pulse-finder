package url

import (
	"context"
	"shared/mongodb/application/config"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// SetupTestContainer initializes the TestContainer.
func SetupTestContainer(t *testing.T) *TestContainer {
	c := NewTestContainer()

	t.Cleanup(func() {
		var (
			ctx    context.Context
			cancel context.CancelFunc
			client *mongo.Client
			err    error
			db     = config.GetConfig().Mongo.DB
		)

		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
		defer cancel()

		if client, err = c.MongoClient.Get().Connect(); err != nil {
			panic(err)
		}
		if err = client.Database(db).Drop(ctx); err != nil {
			panic(err)
		}
		if err = c.MongoClient.Get().Close(); err != nil {
			panic(err)
		}
	})
	return c
}
