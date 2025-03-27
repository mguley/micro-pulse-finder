package collectors

import (
	"log/slog"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HeapMetrics collects detailed Go heap and stack memory usage metrics.
//
// Purpose: Periodically captures heap, stack, and garbage collection statistics
// from the Go runtime and exposes them as Prometheus metrics.
//
// Fields:
//   - BaseCollector: Embeds lifecycle management functionalities.
//   - metrics:       Map holding Prometheus Gauges for each metric.
//   - namespace:     Prefix namespace for all metric names.
type HeapMetrics struct {
	BaseCollector
	metrics   map[string]prometheus.Gauge
	namespace string
}

// NewHeapMetrics constructs a fully initialized HeapMetrics collector.
//
// Parameters:
//   - namespace: Prefix namespace applied to all heap metric names.
//   - logger:    Structured logger for logging collector lifecycle events.
//
// Returns:
//   - *HeapMetrics: Pointer to the initialized HeapMetrics instance.
func NewHeapMetrics(namespace string, logger *slog.Logger) *HeapMetrics {
	heapMetrics := &HeapMetrics{
		metrics:   make(map[string]prometheus.Gauge),
		namespace: namespace,
	}
	heapMetrics.InitBase("HeapMetrics", logger)
	return heapMetrics
}

// InitMetrics initializes and registers heap metrics with the provided Prometheus registry.
//
// Parameters:
//   - registry: Prometheus registry to register heap metrics.
//
// Returns:
//   - err: Error encountered during registration; nil on success.
func (h *HeapMetrics) InitMetrics(registry *prometheus.Registry) (err error) {
	metricDefs := map[string]string{
		"heap_total_alloc_bytes": "Cumulative bytes allocated for heap objects",
		"heap_live_objects":      "Number of live objects (Mallocs - Frees)",
		"heap_alloc_bytes":       "Bytes of allocated heap objects",
		"heap_sys_bytes":         "Heap memory obtained from the OS",
		"heap_idle_bytes":        "Idle heap memory",
		"heap_inuse_bytes":       "Heap memory actively in use",
		"heap_released_bytes":    "Memory returned to OS",

		"stack_inuse_bytes": "Stack memory in use",
		"stack_sys_bytes":   "Stack memory obtained from OS",

		"num_gc":        "Number of completed GC cycles",
		"num_forced_gc": "Number of forced GC cycles",
	}

	for name, help := range metricDefs {
		item := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: h.namespace,
			Name:      name,
			Help:      help,
		})
		if err = registry.Register(item); err != nil {
			h.logger.Error("Heap metric registration failed",
				slog.String("metric", name),
				slog.String("error", err.Error()))
			return err
		}
		h.metrics[name] = item
	}
	return nil
}

// Start initiates periodic collection and update of heap metrics.
//
// Parameters:
//   - interval: Interval at which metrics are collected and updated.
func (h *HeapMetrics) Start(interval time.Duration) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-h.Done():
				return
			case <-ticker.C:
				h.collectMetrics()
			}
		}
	}()
}

// collectMetrics retrieves current memory statistics from Go runtime and updates metrics.
func (h *HeapMetrics) collectMetrics() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	h.metrics["heap_total_alloc_bytes"].Set(float64(mem.TotalAlloc))
	h.metrics["heap_live_objects"].Set(float64(mem.Mallocs - mem.Frees))
	h.metrics["heap_alloc_bytes"].Set(float64(mem.HeapAlloc))
	h.metrics["heap_sys_bytes"].Set(float64(mem.HeapSys))
	h.metrics["heap_idle_bytes"].Set(float64(mem.HeapIdle))
	h.metrics["heap_inuse_bytes"].Set(float64(mem.HeapInuse))
	h.metrics["heap_released_bytes"].Set(float64(mem.HeapReleased))

	h.metrics["stack_inuse_bytes"].Set(float64(mem.StackInuse))
	h.metrics["stack_sys_bytes"].Set(float64(mem.StackSys))

	h.metrics["num_gc"].Set(float64(mem.NumGC))
	h.metrics["num_forced_gc"].Set(float64(mem.NumForcedGC))
}
