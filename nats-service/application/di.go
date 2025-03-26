package application

import (
	"nats-service/application/services"
	"nats-service/infrastructure"
	"shared/dependency"
	"time"

	"github.com/nats-io/nats.go"
)

// Container provides a lazily initialized set of application dependencies.
//
// Fields:
//   - Operations:     Lazy dependency for the NATS operations service.
//   - Infrastructure: Lazy dependency for the infrastructure container.
//   - MetricsService: Lazy dependency for the Prometheus metrics service.
type Container struct {
	Operations     dependency.LazyDependency[*services.Operations]
	Infrastructure dependency.LazyDependency[*infrastructure.Container]
	MetricsService dependency.LazyDependency[*services.MetricsService]
}

// NewContainer initializes and returns a new Container with the required dependencies.
//
// Returns:
//   - *Container: A pointer to the newly created dependency container.
func NewContainer() *Container {
	c := &Container{}

	c.Infrastructure = dependency.LazyDependency[*infrastructure.Container]{
		InitFunc: infrastructure.NewContainer,
	}
	c.Operations = dependency.LazyDependency[*services.Operations]{
		InitFunc: func() *services.Operations {
			var (
				logger = c.Infrastructure.Get().Logger.Get()
				conn   *nats.Conn
				err    error
			)
			if conn, err = c.Infrastructure.Get().NatsClient.Get().Connect(); err != nil {
				panic(err)
			}
			return services.NewOperations(conn, logger)
		},
	}
	c.MetricsService = dependency.LazyDependency[*services.MetricsService]{
		InitFunc: func() *services.MetricsService {
			var (
				provider      = c.Infrastructure.Get().MetricsProvider.Get()
				metricsServer = c.Infrastructure.Get().MetricsServer.Get()
				logger        = c.Infrastructure.Get().Logger.Get()
				interval      = time.Duration(5) * time.Second
			)
			return services.NewMetricsService(provider, metricsServer, logger, interval)
		},
	}

	return c
}
