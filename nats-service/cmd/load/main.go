package main

import (
	"log/slog"
	"nats-service/tests/load"
	"os"

	"github.com/mguley/go-loadtest/pkg/core"
)

func main() {
	var (
		app           = load.NewContainer()
		config        = app.Config.Get()
		orchestrator  = app.Orchestrator.Get()
		runnerFactory = app.NatsServiceRunnerFactory.Get()
		logger        = app.Logger.Get()
		testType      = app.TestType
		testRunner    core.Runner
		err           error
	)

	if testRunner, err = runnerFactory.CreateRunner(testType); err != nil {
		logger.Error("Failed to create test runner", slog.String("error", err.Error()))
		os.Exit(1)
	}

	orchestrator.AddRunner(testRunner)

	// Add metrics collectors.
	orchestrator.AddCollector(app.CompositeCollector.Get())

	// Add reporters.
	orchestrator.AddReporter(app.ConsoleReporter.Get())

	// Log test start and parameters.
	logger.Info("Starting NATS service load test",
		slog.String("test_type", string(testType)),
		slog.String("subject", config.Subject),
		slog.Int("concurrency", config.Concurrency),
		slog.Any("duration", config.Duration.String()))

	// Run the load test.
	if err = orchestrator.Run(); err != nil {
		logger.Error("Load test failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Load test completed successfully")
}
