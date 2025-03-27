package load

import (
	"fmt"
	"log/slog"
	"nats-service/tests/load/config"
	"nats-service/tests/load/infrastructure/runner"
	"os"
	"shared/dependency"
	"shared/grpc/clients/nats_service"

	"github.com/mguley/go-loadtest/pkg"
	"github.com/mguley/go-loadtest/pkg/collector"
	"github.com/mguley/go-loadtest/pkg/core"
	"github.com/mguley/go-loadtest/pkg/reporter"
)

// Container aggregates all the dependencies required to run the load test.
//
// Fields:
//   - Config:                   The load test configuration.
//   - Logger:                   A structured logger for logging events.
//   - Orchestrator:             The load test orchestrator.
//   - NatsRpcClient:            The gRPC client for NATS service communication.
//   - NatsServiceRunnerFactory: Factory for creating NATS service runners.
//   - TestType:                 The type of load test to run (publish or subscribe).
//   - SystemCollector:          Collector for system metrics.
//   - CompositeCollector:       Composite collector to aggregate multiple collectors.
//   - ConsoleReporter:          Reporter that outputs test results to the console.
type Container struct {
	Config                   dependency.LazyDependency[*config.LoadTestConfig]
	Logger                   dependency.LazyDependency[*slog.Logger]
	Orchestrator             dependency.LazyDependency[*pkg.Orchestrator]
	NatsRpcClient            dependency.LazyDependency[*nats_service.NatsClient]
	NatsRpcValidator         dependency.LazyDependency[nats_service.Validator]
	NatsServiceRunnerFactory dependency.LazyDependency[*runner.NatsServiceRunnerFactory]
	TestType                 config.LoadTestType
	SystemCollector          dependency.LazyDependency[*collector.SystemCollector]
	CompositeCollector       dependency.LazyDependency[*collector.CompositeCollector]
	ConsoleReporter          dependency.LazyDependency[*reporter.ConsoleReporter]
}

// NewContainer creates and initializes a new Container with all required dependencies
// for running the load test.
//
// Returns:
//   - *Container: A pointer to the fully initialized dependency container.
func NewContainer() *Container {
	c := new(Container)

	c.Config = dependency.LazyDependency[*config.LoadTestConfig]{
		InitFunc: config.GetConfig,
	}
	c.Logger = dependency.LazyDependency[*slog.Logger]{
		InitFunc: func() *slog.Logger {
			var (
				cfg      = c.Config.Get()
				logLevel = func() slog.Level {
					switch cfg.LogLevel {
					case "debug":
						return slog.LevelDebug
					case "info":
						return slog.LevelInfo
					case "warn":
						return slog.LevelWarn
					case "error":
						return slog.LevelError
					default:
						return slog.LevelInfo
					}
				}()
			)
			return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
		},
	}
	c.Orchestrator = dependency.LazyDependency[*pkg.Orchestrator]{
		InitFunc: func() *pkg.Orchestrator {
			var (
				logger     = c.Logger.Get()
				cfg        = c.Config.Get()
				testConfig = &core.TestConfig{
					TestDuration:   cfg.Duration,
					Concurrency:    cfg.Concurrency,
					WarmupDuration: cfg.WarmupDuration,
					ReportInterval: cfg.ReportInterval,
					LogLevel:       cfg.LogLevel,
					Tags:           cfg.Tags,
				}
			)
			return pkg.NewOrchestrator(testConfig, logger)
		},
	}
	c.NatsRpcValidator = dependency.LazyDependency[nats_service.Validator]{
		InitFunc: func() nats_service.Validator {
			return nats_service.NewBusClientValidator()
		},
	}
	c.NatsRpcClient = dependency.LazyDependency[*nats_service.NatsClient]{
		InitFunc: func() *nats_service.NatsClient {
			var (
				env       = "dev"
				cfg       = c.Config.Get()
				logger    = c.Logger.Get()
				rpcClient *nats_service.NatsClient
				address   = fmt.Sprintf("%s:%s", cfg.RpcHost, cfg.RpcPort)
				validator = c.NatsRpcValidator.Get()
				err       error
			)
			if rpcClient, err = nats_service.NewNatsClient(env, address, validator, logger); err != nil {
				panic(err)
			}
			return rpcClient
		},
	}
	c.NatsServiceRunnerFactory = dependency.LazyDependency[*runner.NatsServiceRunnerFactory]{
		InitFunc: func() *runner.NatsServiceRunnerFactory {
			var (
				natsRpcClient = c.NatsRpcClient.Get()
				cfg           = c.Config.Get()
				logger        = c.Logger.Get()
			)
			return runner.NewNatsServiceRunnerFactory(natsRpcClient, cfg, logger)
		},
	}
	c.TestType = func() config.LoadTestType {
		var cfg = c.Config.Get()
		switch cfg.TestType {
		case "publish":
			return config.PublishTest
		case "subscribe":
			return config.SubscribeTest
		default:
			panic(fmt.Sprintf("unknown load test type: %s", cfg.TestType))
		}
	}()
	c.SystemCollector = dependency.LazyDependency[*collector.SystemCollector]{
		InitFunc: func() *collector.SystemCollector {
			return collector.NewSystemCollector(c.Logger.Get())
		},
	}
	c.CompositeCollector = dependency.LazyDependency[*collector.CompositeCollector]{
		InitFunc: func() *collector.CompositeCollector {
			return collector.NewCompositeCollector(c.SystemCollector.Get())
		},
	}
	c.ConsoleReporter = dependency.LazyDependency[*reporter.ConsoleReporter]{
		InitFunc: func() *reporter.ConsoleReporter {
			return reporter.NewConsoleReporter(os.Stdout, c.Config.Get().ReportInterval)
		},
	}

	return c
}
