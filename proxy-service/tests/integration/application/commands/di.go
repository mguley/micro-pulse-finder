package commands

import (
	"log/slog"
	"os"
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
	Logger         dependency.LazyDependency[*slog.Logger]
	Config         dependency.LazyDependency[*config.Config]
	UserAgent      dependency.LazyDependency[interfaces.Agent]
	Socks5Client   dependency.LazyDependency[*socks5.Client]
	ConnectionPool dependency.LazyDependency[*socks5.ConnectionPool]
	StatusCommand  dependency.LazyDependency[*commands.StatusCommand]
}

// NewTestContainer initializes a new test container.
func NewTestContainer() *TestContainer {
	c := &TestContainer{}

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		},
	}
	c.Config = dependency.LazyDependency[*config.Config]{
		InitFunc: config.GetConfig,
	}
	c.UserAgent = dependency.LazyDependency[interfaces.Agent]{
		InitFunc: func() interfaces.Agent {
			return agent.NewChromeAgent(c.Logger.Get())
		},
	}
	c.Socks5Client = dependency.LazyDependency[*socks5.Client]{
		InitFunc: func() *socks5.Client {
			var (
				logger    = c.Logger.Get()
				userAgent = c.UserAgent.Get()
				timeout   = time.Duration(10) * time.Second
			)
			return socks5.NewClient(userAgent, timeout, logger)
		},
	}
	c.ConnectionPool = dependency.LazyDependency[*socks5.ConnectionPool]{
		InitFunc: func() *socks5.ConnectionPool {
			var (
				logger          = c.Logger.Get()
				poolSize        = c.Config.Get().Pool.MaxSize
				refreshInterval = c.Config.Get().Pool.RefreshInterval
				creator         = c.Socks5Client.Get().Create
			)
			return socks5.NewConnectionPool(poolSize, time.Duration(refreshInterval)*time.Second, creator, logger)
		},
	}
	c.StatusCommand = dependency.LazyDependency[*commands.StatusCommand]{
		InitFunc: func() *commands.StatusCommand {
			var (
				logger  = c.Logger.Get()
				timeout = time.Duration(10) * time.Second
				url     = c.Config.Get().Proxy.Url
				pool    = c.ConnectionPool.Get()
			)
			return commands.NewStatusCommand(timeout, url, pool, logger)
		},
	}

	return c
}
