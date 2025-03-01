package infrastructure

import (
	"log"
	"log/slog"
	"os"
	"proxy-service/application/config"
	"proxy-service/domain/interfaces"
	"proxy-service/infrastructure/http/socks5"
	"proxy-service/infrastructure/http/socks5/agent"
	"proxy-service/infrastructure/proxy"
	"shared/dependency"
	"time"
)

// Container provides a lazily initialized set of dependencies.
type Container struct {
	Logger         dependency.LazyDependency[*slog.Logger]
	Config         dependency.LazyDependency[*config.Config]
	PortConnection dependency.LazyDependency[*proxy.Connection]
	UserAgent      dependency.LazyDependency[interfaces.Agent]
	Socks5Client   dependency.LazyDependency[*socks5.Client]
	ConnectionPool dependency.LazyDependency[*socks5.ConnectionPool]
}

// NewContainer initializes and returns a new Container with dependencies.
func NewContainer() *Container {
	c := &Container{}

	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			var (
				file *os.File
				err  error
			)
			if err = os.MkdirAll(interfaces.LogDir, 0o755); err != nil {
				log.Fatalf("Failed to create log directory: %v", err)
			}
			file, err = os.OpenFile(interfaces.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
			}
			return slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{}))
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
	c.PortConnection = dependency.LazyDependency[*proxy.Connection]{
		InitFunc: func() *proxy.Connection {
			var (
				timeout = time.Duration(10) * time.Second
				logger  = c.Logger.Get()
			)
			return proxy.NewConnection(timeout, logger)
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

	return c
}
