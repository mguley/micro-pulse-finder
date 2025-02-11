package infrastructure

import (
	"proxy-service/infrastructure/proxy"
	"shared/dependency"
	"time"
)

// Container provides a lazily initialized set of dependencies.
type Container struct {
	PortConnection dependency.LazyDependency[*proxy.Connection]
}

// NewContainer initializes and returns a new Container with dependencies.
func NewContainer() *Container {
	c := &Container{}

	c.PortConnection = dependency.LazyDependency[*proxy.Connection]{
		InitFunc: func() *proxy.Connection {
			return proxy.NewConnection(time.Duration(10) * time.Second)
		},
	}

	return c
}
