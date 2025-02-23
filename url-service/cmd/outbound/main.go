package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"
	"url-service/application"
)

func main() {
	var (
		app             = application.NewContainer()
		outboundService = app.OutboundMessageService.Get()
		natsClient      = app.NatsGrpcClient.Get()
		gracePeriod     = time.Duration(2) * time.Second
		outboundCtx     context.Context
		outboundCancel  context.CancelFunc
	)

	outboundCtx, outboundCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer outboundCancel()

	// Start the outbound service.
	go outboundService.Start(outboundCtx)

	<-outboundCtx.Done()
	fmt.Println("Shutdown signal received, stopping outbound service...")
	fmt.Printf("Waiting %v for in-flight operations to complete...\n", gracePeriod)
	time.Sleep(gracePeriod)

	fmt.Println("Closing NATS connection...")
	if err := natsClient.Close(); err != nil {
		fmt.Printf("Error closing NATS connection: %v", err)
	}
	fmt.Println("Outbound service gracefully shutdown")
}
