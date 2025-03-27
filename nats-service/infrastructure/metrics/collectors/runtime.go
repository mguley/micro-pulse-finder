package collectors

import (
	"log/slog"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// RuntimeMetrics collects specific runtime memory statistics.
//
// Purpose: Periodically captures selected Go runtime statistics, complementing
// the default Prometheus GoCollector.
//
// Fields:
//   - BaseCollector: Embeds shared lifecycle management functionality.
//   - metrics:       Map of Prometheus Gauges for each collected metric.
//   - namespace:     Namespace prefix for metric names.
type RuntimeMetrics struct {
	BaseCollector
	metrics   map[string]prometheus.Gauge
	namespace string
}

// NewRuntimeMetrics creates an initialized RuntimeMetrics collector.
//
// Parameters:
//   - namespace: Metric namespace to prevent naming collisions.
//   - logger:    Structured logger instance for collector lifecycle logging.
//
// Returns:
//   - *RuntimeMetrics: Pointer to fully initialized RuntimeMetrics.
func NewRuntimeMetrics(namespace string, logger *slog.Logger) *RuntimeMetrics {
	runtimeMetrics := &RuntimeMetrics{
		metrics:   make(map[string]prometheus.Gauge),
		namespace: namespace,
	}
	runtimeMetrics.InitBase("RuntimeMetrics", logger)
	return runtimeMetrics
}

// InitMetrics initializes and registers runtime metrics.
//
// Parameters:
//   - registry: Prometheus registry for metric registration.
//
// Returns:
//   - err: Error during registration, or nil if successful.
func (r *RuntimeMetrics) InitMetrics(registry *prometheus.Registry) (err error) {
	metricDefs := map[string]string{
		"heap_objects": "Number of allocated heap objects",
	}

	for name, help := range metricDefs {
		item := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: r.namespace,
			Name:      name,
			Help:      help,
		})
		if err = registry.Register(item); err != nil {
			r.logger.Error("Runtime metric registration failed",
				slog.String("metric", name),
				slog.String("error", err.Error()))
			return err
		}
		r.metrics[name] = item
	}
	return nil
}

// Start initiates periodic collection of runtime metrics.
//
// Parameters:
//   - interval: Interval for updating runtime metrics.
func (r *RuntimeMetrics) Start(interval time.Duration) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-r.Done():
				return
			case <-ticker.C:
				r.collectMetrics()
			}
		}
	}()
}

// collectMetrics updates runtime metrics with current data from Go runtime.
func (r *RuntimeMetrics) collectMetrics() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	r.metrics["heap_objects"].Set(float64(mem.HeapObjects))
}
