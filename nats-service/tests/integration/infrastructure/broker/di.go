package broker

import (
	"log"
	"log/slog"
	"nats-service/domain/entities"
	"nats-service/infrastructure/broker"
	"os"
	"shared/dependency"
	"shared/observability/nats-service/metrics"
	"time"

	"github.com/nats-io/nats.go"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Logger     dependency.LazyDependency[*slog.Logger]
	Metrics    dependency.LazyDependency[*metrics.Metrics]
	NatsClient dependency.LazyDependency[*broker.Client]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		},
	}
	c.Metrics = dependency.LazyDependency[*metrics.Metrics]{
		InitFunc: metrics.NewMetrics,
	}
	c.NatsClient = dependency.LazyDependency[*broker.Client]{
		InitFunc: func() *broker.Client {
			var (
				logger           = c.Logger.Get()
				address          string
				err              error
				reconnectWait    = time.Duration(5) * time.Second
				maxReconnect     = 5
				timeout          = time.Duration(5) * time.Second
				reconnectHandler = func(conn *nats.Conn) {
					log.Println("Reconnected to NATS at", conn.ConnectedUrl())
				}
				disconnectErrHandler = func(conn *nats.Conn, cErr error) {
					log.Println("Disconnected from NATS due to: ", cErr)
				}
			)
			if address, err = entities.GetBroker().Address(); err != nil {
				panic(err)
			}

			options := &nats.Options{
				Url:               address,
				ReconnectWait:     reconnectWait,
				MaxReconnect:      maxReconnect,
				Timeout:           timeout,
				ReconnectedCB:     reconnectHandler,
				DisconnectedErrCB: disconnectErrHandler,
				AllowReconnect:    true,
			}
			return broker.NewClient(options, c.Metrics.Get(), logger)
		},
	}

	return c
}
