package main

import (
	"context"
	"log/slog"
	"nats-service/application"
	"time"
)

func main() {
	var (
		appContainer   = application.NewContainer()
		infra          = application.NewContainer().Infrastructure.Get()
		logger         = infra.Logger.Get()
		busServer      = infra.BusServer.Get()
		busService     = infra.BusService.Get()
		metricsService = appContainer.MetricsService.Get()
	)

	metricsService.Start()
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
		defer cancel()

		if err := metricsService.Stop(stopCtx); err != nil {
			logger.Debug("Error stopping metrics service", slog.String("error", err.Error()))
		}
	}()

	logger.Info("Registering bus service with gRPC server...")
	busServer.RegisterService(busService)

	logger.Info("Starting NATS RPC server")
	busServer.Start()
	busServer.WaitForShutdown()
}
