package main

import (
	"context"
	"log/slog"
	"nats-service/application"
)

func main() {
	var (
		app            = application.NewContainer()
		infra          = application.NewContainer().Infrastructure.Get()
		logger         = infra.Logger.Get()
		busServer      = infra.BusServer.Get()
		busService     = infra.BusService.Get()
		metricsService = app.MetricsService.Get()
	)

	metricsService.Start()
	defer func() {
		if err := metricsService.Stop(context.Background()); err != nil {
			logger.Debug("Error stopping metrics service", slog.String("error", err.Error()))
		}
	}()

	logger.Info("Registering bus service with gRPC server...")
	busServer.RegisterService(busService)

	logger.Info("Starting NATS RPC server")
	busServer.Start()
	busServer.WaitForShutdown()
}
