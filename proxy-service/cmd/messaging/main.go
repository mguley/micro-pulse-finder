package main

import (
	"context"
	"os/signal"
	"proxy-service/application"
	"syscall"
	"time"
)

func main() {
	var (
		app             = application.NewContainer()
		logger          = app.Infrastructure.Get().Logger.Get()
		urlProcessor    = app.UrlProcessorService.Get()
		connectionPool  = app.Infrastructure.Get().ConnectionPool.Get()
		natsClient      = app.NatsGrpcClient.Get()
		gracePeriod     = time.Duration(2) * time.Second
		processorCtx    context.Context
		processorCancel context.CancelFunc
	)

	processorCtx, processorCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer processorCancel()

	logger.Info("Starting messaging service")

	// Start the URL processor, it will listen for messages until the context is canceled.
	go func() {
		if err := urlProcessor.Start(processorCtx); err != nil {
			logger.Error("Error running URL processor", "error", err)
		}
	}()

	<-processorCtx.Done()
	logger.Info("Shutdown signal received, commencing graceful shutdown")
	logger.Info("Waiting for in-flight operations to complete", "gracePeriod", gracePeriod)
	time.Sleep(gracePeriod)

	// Clean up resources.
	logger.Info("Shutting down connection pool")
	connectionPool.Shutdown()

	logger.Info("Closing NATS client connection")
	if err := natsClient.Close(); err != nil {
		logger.Error("Error closing NATS client", "error", err)
	}

	logger.Info("Service gracefully shutdown")
}
