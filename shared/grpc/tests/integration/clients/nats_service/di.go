package nats_service

import (
	"log/slog"
	"os"
	"shared/dependency"
	"shared/grpc/clients/nats_service"
	"shared/grpc/tests/integration/clients/nats_service/server"
)

// TestContainer holds dependencies for integration tests.
type TestContainer struct {
	Logger               dependency.LazyDependency[*slog.Logger]
	MockBusServiceServer dependency.LazyDependency[*server.MockBusService]
	TestServerContainer  dependency.LazyDependency[*server.TestServerContainer]
	NatsValidator        dependency.LazyDependency[nats_service.Validator]
	NatsClient           dependency.LazyDependency[*nats_service.NatsClient]
}

// NewTestContainer initializes the test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		},
	}
	c.MockBusServiceServer = dependency.LazyDependency[*server.MockBusService]{
		InitFunc: server.NewMockBusService,
	}
	c.TestServerContainer = dependency.LazyDependency[*server.TestServerContainer]{
		InitFunc: func() *server.TestServerContainer {
			var (
				testServer *server.TestServerContainer
				err        error
			)
			if testServer, err = server.NewTestServerContainer(c.MockBusServiceServer.Get()); err != nil {
				panic(err)
			}
			return testServer
		},
	}
	c.NatsValidator = dependency.LazyDependency[nats_service.Validator]{
		InitFunc: func() nats_service.Validator {
			return nats_service.NewBusClientValidator()
		},
	}
	c.NatsClient = dependency.LazyDependency[*nats_service.NatsClient]{
		InitFunc: func() *nats_service.NatsClient {
			var (
				logger     = c.Logger.Get()
				address    = c.TestServerContainer.Get().Address
				validator  = c.NatsValidator.Get()
				natsClient *nats_service.NatsClient
				env        = "dev"
				err        error
			)
			if natsClient, err = nats_service.NewNatsClient(env, address, validator, logger); err != nil {
				panic(err)
			}
			return natsClient
		},
	}

	return c
}
