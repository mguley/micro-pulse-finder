package commands

import (
	"proxy-service/application/commands"
	"proxy-service/application/config"
	"proxy-service/domain/interfaces"
	"proxy-service/infrastructure/http/socks5"
	"proxy-service/infrastructure/http/socks5/agent"
	"shared/dependency"
	"time"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Config        dependency.LazyDependency[*config.Config]
	UserAgent     dependency.LazyDependency[interfaces.Agent]
	Socks5Client  dependency.LazyDependency[*socks5.Client]
	StatusCommand dependency.LazyDependency[*commands.StatusCommand]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.Config = dependency.LazyDependency[*config.Config]{
		InitFunc: config.GetConfig,
	}
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
	c.StatusCommand = dependency.LazyDependency[*commands.StatusCommand]{
		InitFunc: func() *commands.StatusCommand {
			var (
				timeout = time.Duration(10) * time.Second
				url     = c.Config.Get().Proxy.Url
				client  = c.Socks5Client.Get()
			)
			return commands.NewStatusCommand(timeout, url, client)
		},
	}

	return c
}
