package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"
	"url-service/application"
)

func main() {
	var (
		app            = application.NewContainer()
		logger         = app.Infrastructure.Get().Logger.Get()
		inboundService = app.InboundMessageService.Get()
		natsClient     = app.NatsGrpcClient.Get()
		gracePeriod    = time.Duration(2) * time.Second
		inboundCtx     context.Context
		inboundCancel  context.CancelFunc
	)

	inboundCtx, inboundCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer inboundCancel()

	logger.Info("Starting inbound service")
	go func() {
		if err := inboundService.Start(inboundCtx); err != nil {
			logger.Error("Error starting inbound service", "error", err)
		}
	}()

	<-inboundCtx.Done()
	logger.Info("Shutdown signal received, stopping inbound service", "gracePeriod", gracePeriod)
	time.Sleep(gracePeriod)

	logger.Info("Closing NATS connection...")
	if err := natsClient.Close(); err != nil {
		logger.Error("Error closing NATS connection", "error", err)
	}
	logger.Info("Inbound service gracefully shutdown.")
}
