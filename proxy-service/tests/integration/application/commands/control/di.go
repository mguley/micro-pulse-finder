package control

import (
	"log/slog"
	"os"
	"proxy-service/application/commands/control"
	"proxy-service/application/config"
	"proxy-service/infrastructure/proxy"
	"shared/dependency"
	"time"
)

// TestContainer holds dependencies for the integration tests.
type TestContainer struct {
	Logger              dependency.LazyDependency[*slog.Logger]
	Config              dependency.LazyDependency[*config.Config]
	PortConnection      dependency.LazyDependency[*proxy.Connection]
	AuthenticateCommand dependency.LazyDependency[*control.AuthenticateCommand]
	SignalCommand       dependency.LazyDependency[*control.SignalCommand]
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
	c.PortConnection = dependency.LazyDependency[*proxy.Connection]{
		InitFunc: func() *proxy.Connection {
			var (
				timeout = time.Duration(10) * time.Second
				logger  = c.Logger.Get()
			)
			return proxy.NewConnection(timeout, logger)
		},
	}
	c.AuthenticateCommand = dependency.LazyDependency[*control.AuthenticateCommand]{
		InitFunc: func() *control.AuthenticateCommand {
			var (
				logger   = c.Logger.Get()
				adapter  = c.PortConnection.Get()
				password = c.Config.Get().Proxy.ControlPassword
			)
			return control.NewAuthenticateCommand(adapter, password, logger)
		},
	}
	c.SignalCommand = dependency.LazyDependency[*control.SignalCommand]{
		InitFunc: func() *control.SignalCommand {
			var (
				logger  = c.Logger.Get()
				adapter = c.PortConnection.Get()
				signal  = "NEWNYM"
			)
			return control.NewSignalCommand(adapter, signal, logger)
		},
	}

	return c
}
