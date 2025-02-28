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
		app             = application.NewContainer()
		logger          = app.Infrastructure.Get().Logger.Get()
		outboundService = app.OutboundMessageService.Get()
		natsClient      = app.NatsGrpcClient.Get()
		gracePeriod     = time.Duration(2) * time.Second
		outboundCtx     context.Context
		outboundCancel  context.CancelFunc
	)

	outboundCtx, outboundCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer outboundCancel()

	logger.Info("Starting outbound service")
	go outboundService.Start(outboundCtx)

	<-outboundCtx.Done()
	logger.Info("Shutdown signal received, stopping outbound service", "gracePeriod", gracePeriod)
	time.Sleep(gracePeriod)

	logger.Info("Closing NATS connection...")
	if err := natsClient.Close(); err != nil {
		logger.Error("Error closing NATS connection", "error", err)
	}
	logger.Info("Outbound service gracefully shutdown.")
}
