package main

import (
	"context"
	"log"
	"os/signal"
	"proxy-service/application"
	"syscall"
	"time"
)

func main() {
	app := application.NewContainer()
	urlProcessor := app.UrlProcessorService.Get()

	// Create a context that will be canceled on SIGINT or SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the URL processor, it will listen for messages until the context is canceled.
	done := make(chan struct{})
	go func() {
		if err := urlProcessor.Start(ctx); err != nil {
			log.Printf("Error running URL processor: %v", err)
		}
		close(done)
	}()

	// Wait until either a termination signal is received or the URL processor finishes.
	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received, commencing graceful shutdown...")
	case <-done:
		log.Println("URL processor has finished processing messages.")
	}

	// Allow in-flight requests a short period to finish.
	gracePeriod := time.Duration(2) * time.Second
	log.Printf("Waiting %v for in-flight operations to complete...", gracePeriod)
	time.Sleep(gracePeriod)

	// Clean up resources.
	log.Println("Shutting down connection pool...")
	app.Infrastructure.Get().ConnectionPool.Get().Shutdown()

	log.Println("Closing NATS client connection...")
	if err := app.NatsGrpcClient.Get().Close(); err != nil {
		log.Printf("Error closing NATS client: %v", err)
	}

	log.Println("Service gracefully shutdown")
}
