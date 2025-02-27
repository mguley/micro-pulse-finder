package main

import "nats-service/application"

func main() {
	var (
		app        = application.NewContainer().Infrastructure.Get()
		logger     = app.Logger.Get()
		busServer  = app.BusServer.Get()
		busService = app.BusService.Get()
	)

	logger.Debug("Registering bus service with gRPC server...")
	busServer.RegisterService(busService)

	logger.Info("Starting gRPC server...")
	busServer.Start()
	busServer.WaitForShutdown()
}
