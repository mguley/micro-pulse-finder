package processor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SetupTestContainer initializes the TestContainer.
func SetupTestContainer() (container *TestContainer, teardown func()) {
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

	// Start the gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		busServer.Start()
	}()

	// Start the URL processor service.
	var (
		urlProcessor = container.UrlProcessorService.Get()
		ctx          context.Context
		cancel       context.CancelFunc
	)
	ctx, cancel = context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := urlProcessor.Start(ctx); err != nil {
			cancel()
			fmt.Printf("Could not subscribe to the ProxyUrlRequest subject: %v", err)
			return
		}
	}()

	// Give the background services a moment to start.
	time.Sleep(time.Duration(2) * time.Second)

	// Define the teardown function to gracefully stop background services.
	teardown = func() {
		fmt.Println("Tearing down integration test environment...")
		cancel()
		busServer.GracefulStop()
		wg.Wait()
	}

	return container, teardown
}
