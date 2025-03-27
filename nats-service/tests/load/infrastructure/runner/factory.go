package runner

import (
	"fmt"
	"log/slog"
	"nats-service/tests/load/config"
	"shared/grpc/clients/nats_service"

	"github.com/mguley/go-loadtest/pkg/core"
)

// NatsServiceRunnerFactory creates NATS service runners based on the load test configuration.
//
// Fields:
//   - client: Pointer to the NatsRpcClient for communicating with the NATS service.
//   - config: Pointer to the load test configuration.
//   - logger: Logger instance for event logging.
type NatsServiceRunnerFactory struct {
	client *nats_service.NatsClient
	config *config.LoadTestConfig
	logger *slog.Logger
}

// NewNatsServiceRunnerFactory creates a new instance of NatsServiceRunnerFactory.
//
// Parameters:
//   - client: Pointer to the NatsRpcClient for gRPC communication.
//   - config: Pointer to the load test configuration.
//   - logger: Logger instance for logging events.
//
// Returns:
//   - *NatsServiceRunnerFactory: A pointer to the newly created NatsServiceRunnerFactory.
func NewNatsServiceRunnerFactory(
	client *nats_service.NatsClient,
	config *config.LoadTestConfig,
	logger *slog.Logger,
) *NatsServiceRunnerFactory {
	return &NatsServiceRunnerFactory{
		client: client,
		config: config,
		logger: logger,
	}
}

// CreateRunner creates a new runner based on the specified load test type.
//
// Parameters:
//   - testType: The type of load test to perform (e.g., PublishTest or SubscribeTest).
//
// Returns:
//   - core.Runner: An instance of a runner that implements the core.Runner interface.
//   - error: An error if the load test type is unknown.
func (f *NatsServiceRunnerFactory) CreateRunner(testType config.LoadTestType) (runner core.Runner, err error) {
	switch testType {
	case config.PublishTest:
		return NewNatsServicePublishRunner(f.client, f.config.MessageSize, f.config.Subject, f.logger), nil
	case config.SubscribeTest:
		return runner, nil
	default:
		return nil, fmt.Errorf("unknown load test type: %s", testType)
	}
}

// Name returns a descriptive name for this factory.
//
// Returns:
//   - string: A name identifying the factory.
func (f *NatsServiceRunnerFactory) Name() string {
	return "NATS Service Runner Factory"
}
