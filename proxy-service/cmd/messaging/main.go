package main

import (
	"context"
	"fmt"
	"os/signal"
	"proxy-service/application"
	"syscall"
	"time"
)

func main() {
	var (
		app             = application.NewContainer()
		urlProcessor    = app.UrlProcessorService.Get()
		connectionPool  = app.Infrastructure.Get().ConnectionPool.Get()
		natsClient      = app.NatsGrpcClient.Get()
		gracePeriod     = time.Duration(2) * time.Second
		processorCtx    context.Context
		processorCancel context.CancelFunc
	)

	processorCtx, processorCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer processorCancel()

	// Start the URL processor, it will listen for messages until the context is canceled.
	go func() {
		if err := urlProcessor.Start(processorCtx); err != nil {
			fmt.Printf("Error running URL processor: %v", err)
		}
	}()

	<-processorCtx.Done()
	fmt.Println("Shutdown signal received, commencing graceful shutdown...")
	fmt.Printf("Waiting %v for in-flight operations to complete...\n", gracePeriod)
	time.Sleep(gracePeriod)

	// Clean up resources.
	fmt.Println("Shutting down connection pool...")
	connectionPool.Shutdown()

	fmt.Println("Closing NATS client connection...")
	if err := natsClient.Close(); err != nil {
		fmt.Printf("Error closing NATS client: %v", err)
	}
	fmt.Println("Service gracefully shutdown")
}
