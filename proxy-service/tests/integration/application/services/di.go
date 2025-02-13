package services

import (
	"proxy-service/application/services"
	"proxy-service/domain/interfaces"
	"shared/dependency"
	"time"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	RetryStrategy dependency.LazyDependency[interfaces.RetryStrategy]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.RetryStrategy = dependency.LazyDependency[interfaces.RetryStrategy]{
		InitFunc: func() interfaces.RetryStrategy {
			var (
				baseDelay  = time.Duration(5) * time.Second
				maxDelay   = time.Duration(45) * time.Second
				attempts   = 5
				multiplier = 2.0
			)
			return services.NewExponentialBackoffStrategy(baseDelay, maxDelay, attempts, multiplier)
		},
	}

	return c
}
