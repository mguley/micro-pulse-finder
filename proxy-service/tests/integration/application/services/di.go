package services

import (
	"log/slog"
	"os"
	"proxy-service/application/services"
	"proxy-service/domain/interfaces"
	"shared/dependency"
	"time"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Logger        dependency.LazyDependency[*slog.Logger]
	RetryStrategy dependency.LazyDependency[interfaces.RetryStrategy]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		},
	}
	c.RetryStrategy = dependency.LazyDependency[interfaces.RetryStrategy]{
		InitFunc: func() interfaces.RetryStrategy {
			var (
				logger     = c.Logger.Get()
				baseDelay  = time.Duration(5) * time.Second
				maxDelay   = time.Duration(45) * time.Second
				attempts   = 5
				multiplier = 2.0
			)
			return services.NewExponentialBackoffStrategy(baseDelay, maxDelay, attempts, multiplier, logger)
		},
	}

	return c
}
