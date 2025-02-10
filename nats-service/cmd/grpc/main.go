package main

import "nats-service/application"

func main() {
	app := application.NewContainer().Infrastructure.Get()
	busServer := app.BusServer.Get()
	busService := app.BusService.Get()

	// Register
	busServer.RegisterService(busService)

	// Start
	busServer.Start()
	busServer.WaitForShutdown()
}
