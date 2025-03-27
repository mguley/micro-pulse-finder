package services

import (
	"context"
	"errors"
	"log/slog"
	"nats-service/infrastructure/metrics"
	metricsCollectors "nats-service/infrastructure/metrics/collectors"
	"net/http"
	"sync"
	"time"
)

// MetricsService manages the lifecycle of application metrics collection and exposure.
//
// Fields:
//   - provider: Metrics provider responsible for managing metrics collectors.
//   - server:   HTTP metrics server instance for exposing collected metrics.
//   - logger:   Structured logger instance for lifecycle logging.
//   - interval: Duration interval at which metrics are collected.
//   - wg:       WaitGroup ensuring graceful shutdown.
type MetricsService struct {
	provider *metrics.Provider
	server   *metrics.Server
	logger   *slog.Logger
	interval time.Duration
	wg       sync.WaitGroup
}

// NewMetricsService creates and initializes a new MetricsService.
//
// Parameters:
//   - provider: Initialized metrics.Provider for managing collectors.
//   - server:   Initialized metrics.Server exposing metrics via HTTP.
//   - logger:   Structured logger for logging metrics service lifecycle events.
//   - interval: Duration between metrics collection cycles.
//
// Returns:
//   - *MetricsService: Fully initialized MetricsService instance.
func NewMetricsService(
	provider *metrics.Provider,
	server *metrics.Server,
	logger *slog.Logger,
	interval time.Duration,
) *MetricsService {
	return &MetricsService{
		provider: provider,
		server:   server,
		logger:   logger,
		interval: interval,
	}
}

// Start initiates metrics collection and starts the metrics HTTP server.
//
// Implementation Detail: The server is started in a separate goroutine to prevent blocking the calling goroutine.
func (m *MetricsService) Start() {
	m.provider.StartCollectors(m.interval)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.logger.Error("Metrics server stopped unexpectedly", slog.String("error", err.Error()))
		}
	}()

	m.logger.Info("Metrics service started")
}

// Stop gracefully stops all metrics collection and shuts down the metrics server.
//
// Parameters:
//   - ctx: Context to control graceful shutdown timeout.
//
// Returns:
//   - error: An error if stopping the metrics server encounters issues; otherwise, nil.
func (m *MetricsService) Stop(ctx context.Context) (err error) {
	m.logger.Info("Stopping metrics service")
	m.provider.Stop(time.Duration(10) * time.Second)

	if err = m.server.Stop(ctx); err != nil {
		m.logger.Error("Error stopping metrics HTTP server", slog.String("error", err.Error()))
		return err
	}

	m.wg.Wait()
	m.logger.Info("Metrics service stopped successfully")
	return nil
}

// GetCollectors returns a list of active metrics collectors.
//
// Returns:
//   - []metricsCollectors.Collector: Slice containing registered metric collectors.
func (m *MetricsService) GetCollectors() []metricsCollectors.Collector {
	return m.provider.Collectors
}
