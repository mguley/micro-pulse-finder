package proxy

import (
	"proxy-service/infrastructure/proxy"
	"shared/dependency"
	"time"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	PortConnection dependency.LazyDependency[*proxy.Connection]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.PortConnection = dependency.LazyDependency[*proxy.Connection]{
		InitFunc: func() *proxy.Connection {
			return proxy.NewConnection(time.Duration(10) * time.Second)
		},
	}

	return c
}
