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
		app            = application.NewContainer()
		inboundService = app.InboundMessageService.Get()
		natsClient     = app.NatsGrpcClient.Get()
		gracePeriod    = time.Duration(2) * time.Second
		inboundCtx     context.Context
		inboundCancel  context.CancelFunc
	)

	inboundCtx, inboundCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer inboundCancel()

	// Start the inbound service.
	go func() {
		if err := inboundService.Start(inboundCtx); err != nil {
			fmt.Printf("Error starting inbound service: %v", err)
		}
	}()

	<-inboundCtx.Done()
	fmt.Println("Shutdown signal received, stopping inbound service...")
	fmt.Printf("Waiting %v for in-flight operations to complete...\n", gracePeriod)
	time.Sleep(gracePeriod)

	fmt.Println("\nClosing NATS connection...")
	if err := natsClient.Close(); err != nil {
		fmt.Printf("Error closing NATS connection: %v", err)
	}
	fmt.Println("Inbound service gracefully shutdown.")
}
