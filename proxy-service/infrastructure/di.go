package infrastructure

import (
	"proxy-service/domain/interfaces"
	"proxy-service/infrastructure/http/socks5"
	"proxy-service/infrastructure/http/socks5/agent"
	"proxy-service/infrastructure/proxy"
	"shared/dependency"
	"time"
)

// Container provides a lazily initialized set of dependencies.
type Container struct {
	PortConnection dependency.LazyDependency[*proxy.Connection]
	UserAgent      dependency.LazyDependency[interfaces.Agent]
	Socks5Client   dependency.LazyDependency[*socks5.Client]
}

// NewContainer initializes and returns a new Container with dependencies.
func NewContainer() *Container {
	c := &Container{}

	c.UserAgent = dependency.LazyDependency[interfaces.Agent]{
		InitFunc: func() interfaces.Agent {
			return agent.NewChromeAgent()
		},
	}
	c.Socks5Client = dependency.LazyDependency[*socks5.Client]{
		InitFunc: func() *socks5.Client {
			var (
				userAgent = c.UserAgent.Get()
				timeout   = time.Duration(10) * time.Second
			)
			return socks5.NewClient(userAgent, timeout)
		},
	}
	c.PortConnection = dependency.LazyDependency[*proxy.Connection]{
		InitFunc: func() *proxy.Connection {
			return proxy.NewConnection(time.Duration(10) * time.Second)
		},
	}

	return c
}
