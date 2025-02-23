package messages

import (
	"context"
	"fmt"
	"shared/mongodb/application/config"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// SetupTestContainer initializes the TestContainer.
func SetupTestContainer(t *testing.T) (container *TestContainer, teardown func()) {
	container = NewTestContainer()

	// Start the nats-service gRPC server.
	var (
		wg         sync.WaitGroup
		natsInfra  = container.NatsServiceInfrastructure.Get()
		busServer  = natsInfra.BusServer.Get()
		busService = natsInfra.BusService.Get()
	)

	// Register the BusService with the BusServer.
	busServer.RegisterService(busService)

	// Start the gRPC server.
	wg.Add(1)
	go func() {
		defer wg.Done()
		busServer.Start()
	}()

	// Start the inbound message service.
	var (
		inboundMessageService = container.InboundMessageService.Get()
		inboundCtx            context.Context
		inboundCancel         context.CancelFunc
	)
	inboundCtx, inboundCancel = context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := inboundMessageService.Start(inboundCtx); err != nil {
			inboundCancel()
			fmt.Printf("Could not start inbound message service: %v\n", err)
			return
		}
	}()

	// Start the outbound message service.
	var (
		outboundMessageService = container.OutboundMessageService.Get()
		outboundCtx            context.Context
		outboundCancel         context.CancelFunc
	)
	outboundCtx, outboundCancel = context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		outboundMessageService.Start(outboundCtx)
	}()

	// Give the background services a moment to start.
	time.Sleep(time.Duration(2) * time.Second)

	// Define the teardown function to gracefully stop background services.
	teardown = func() {
		fmt.Println("Tearing down integration test environment...")
		inboundCancel()
		outboundCancel()
		busServer.GracefulStop()
		wg.Wait()
	}

	// Clean up MongoDB.
	t.Cleanup(func() {
		var (
			mongoCtx    context.Context
			mongoCancel context.CancelFunc
			client      *mongo.Client
			err         error
			db          = config.GetConfig().Mongo.DB
		)
		mongoCtx, mongoCancel = context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
		defer mongoCancel()

		if client, err = container.MongoClient.Get().Connect(); err != nil {
			panic(err)
		}
		if err = client.Database(db).Drop(mongoCtx); err != nil {
			panic(err)
		}
		if err = container.MongoClient.Get().Close(); err != nil {
			panic(err)
		}
	})

	return container, teardown
}
