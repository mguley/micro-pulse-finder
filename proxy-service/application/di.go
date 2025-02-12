package application

import (
	"proxy-service/application/commands/control"
	"proxy-service/application/config"
	"proxy-service/infrastructure"
	"shared/dependency"
)

// Container provides a lazily initialized set of dependencies.
type Container struct {
	Config              dependency.LazyDependency[*config.Config]
	Infrastructure      dependency.LazyDependency[*infrastructure.Container]
	AuthenticateCommand dependency.LazyDependency[*control.AuthenticateCommand]
}

// NewContainer initializes and returns a new Container with dependencies.
func NewContainer() *Container {
	c := &Container{}

	c.Config = dependency.LazyDependency[*config.Config]{
		InitFunc: config.GetConfig,
	}
	c.Infrastructure = dependency.LazyDependency[*infrastructure.Container]{
		InitFunc: infrastructure.NewContainer,
	}

	// Proxy commands
	c.AuthenticateCommand = dependency.LazyDependency[*control.AuthenticateCommand]{
		InitFunc: func() *control.AuthenticateCommand {
			var (
				adapter  = c.Infrastructure.Get().PortConnection.Get()
				password = c.Config.Get().Proxy.ControlPassword
			)
			return control.NewAuthenticateCommand(adapter, password)
		},
	}

	return c
}
