package metrics

import (
	"log/slog"
	metricsCollectors "nats-service/infrastructure/metrics/collectors"
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Provider manages multiple Prometheus metrics collectors.
//
// Fields:
//   - Registry:   Prometheus registry for collecting and exposing metrics.
//   - Collectors: Slice containing registered metrics collectors.
//   - logger:     Logger for recording collector statuses and errors.
type Provider struct {
	Registry   *prometheus.Registry
	Collectors []metricsCollectors.Collector
	logger     *slog.Logger
}

// NewProvider creates and initializes a new Provider instance.
//
// Parameters:
//   - namespace: A string namespace for metrics to avoid naming collisions.
//   - logger:    Logger instance for structured logging.
//
// Returns:
//   - *Provider: Initialized Provider instance with registered collectors.
func NewProvider(namespace string, logger *slog.Logger) *Provider {
	registry := prometheus.NewRegistry()

	// Add default prometheus collectors
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	allCollectors := []metricsCollectors.Collector{
		metricsCollectors.NewRuntimeMetrics(namespace, logger),
		metricsCollectors.NewHeapMetrics(namespace, logger),
		// Add more collectors here
	}

	for _, collector := range allCollectors {
		if err := collector.InitMetrics(registry); err != nil {
			logger.Error("Collector init failed", slog.String("error", err.Error()))
		}
	}

	return &Provider{
		Registry:   registry,
		Collectors: allCollectors,
		logger:     logger,
	}
}

// StartCollectors initiates metric collection for all registered collectors.
//
// Parameters:
//   - interval: Duration defining how frequently metrics are collected.
func (p *Provider) StartCollectors(interval time.Duration) {
	for _, collector := range p.Collectors {
		collector.Start(interval)
	}
	p.logger.Info("Metrics collectors started")
}

// Stop gracefully terminates all metric collectors.
func (p *Provider) Stop(timeout time.Duration) {
	for _, collector := range p.Collectors {
		collector.StopWithTimeout(timeout)
	}
	p.logger.Info("Metrics collectors stopped")
}

// GetCollectorByType returns a collector matching the given type.
//
// Parameters:
//   - t: The reflect.Type of the desired collector.
//
// Returns:
//   - metricsCollectors.Collector: The matching collector, or nil if none is found.
func (p *Provider) GetCollectorByType(t reflect.Type) (collector metricsCollectors.Collector) {
	for _, collector = range p.Collectors {
		if reflect.TypeOf(collector) == t {
			return collector
		}
	}
	return nil
}
