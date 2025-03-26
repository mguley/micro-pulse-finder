package application

import (
	"nats-service/application/services"
	"nats-service/infrastructure"
	"shared/dependency"

	"github.com/nats-io/nats.go"
)

// Container provides a lazily initialized set of application dependencies.
//
// Fields:
//   - Operations:     Lazy dependency for the NATS operations service.
//   - Infrastructure: Lazy dependency for the infrastructure container.
type Container struct {
	Operations     dependency.LazyDependency[*services.Operations]
	Infrastructure dependency.LazyDependency[*infrastructure.Container]
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
				logger  = c.Infrastructure.Get().Logger.Get()
				metrics = c.Infrastructure.Get().Metrics.Get()
				conn    *nats.Conn
				err     error
			)
			if conn, err = c.Infrastructure.Get().NatsClient.Get().Connect(); err != nil {
				panic(err)
			}
			return services.NewOperations(conn, metrics, logger)
		},
	}

	return c
}
