package proxy

import (
	"log/slog"
	"os"
	"proxy-service/infrastructure/proxy"
	"shared/dependency"
	"time"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Logger         dependency.LazyDependency[*slog.Logger]
	PortConnection dependency.LazyDependency[*proxy.Connection]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		},
	}
	c.PortConnection = dependency.LazyDependency[*proxy.Connection]{
		InitFunc: func() *proxy.Connection {
			var (
				timeout = time.Duration(10) * time.Second
				logger  = c.Logger.Get()
			)
			return proxy.NewConnection(timeout, logger)
		},
	}

	return c
}
