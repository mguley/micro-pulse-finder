package collectors

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// RuntimeMetrics collects Go runtime statistics.
//
// Fields:
//   - BaseCollector:  Embeds BaseCollector for lifecycle management.
//   - goroutines:     Gauge tracking the number of active goroutines.
//   - gcPauses:       Histogram of garbage collection pause durations.
//   - malloc:         Counter for total memory allocations.
//   - frees:          Counter for total memory frees.
//   - gcCount:        Counter for total garbage collections performed.
//   - heapObjects:    Gauge tracking the number of allocated heap objects.
//   - prevMalloc:     Stored value for previous memory allocations (used for delta computation).
//   - prevFrees:      Stored value for previous memory frees (used for delta computation).
//   - prevGCCount:    Stored value for previous garbage collection count (used for delta computation).
//   - namespace:      Metric namespace to avoid naming collisions.
type RuntimeMetrics struct {
	BaseCollector
	goroutines  prometheus.Gauge
	gcPauses    prometheus.Histogram
	malloc      prometheus.Counter
	frees       prometheus.Counter
	gcCount     prometheus.Counter
	heapObjects prometheus.Gauge
	prevMalloc  uint64
	prevFrees   uint64
	prevGCCount uint32
	namespace   string
}

// NewRuntimeMetrics initializes a new RuntimeMetrics collector.
//
// Parameters:
//   - namespace: Metric namespace to uniquely identify runtime metrics.
//
// Returns:
//   - *RuntimeMetrics: Fully initialized runtime metrics collector.
func NewRuntimeMetrics(namespace string) *RuntimeMetrics {
	runtimeMetrics := &RuntimeMetrics{
		goroutines: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "goroutines",
			Help:      "Number of running goroutines",
		}),
		gcPauses: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "gc_pause_seconds",
			Help:      "Histogram of GC pause durations",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 15),
		}),
		malloc: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "memory_malloc_total",
			Help:      "Total number of memory allocations",
		}),
		frees: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "memory_frees_total",
			Help:      "Total number of memory frees",
		}),
		gcCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "gc_count_total",
			Help:      "Total number of garbage collections",
		}),
		heapObjects: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "heap_objects",
			Help:      "HeapObjects is the number of allocated heap objects.",
		}),
		namespace: namespace,
	}

	runtimeMetrics.Init()
	return runtimeMetrics
}

// Register registers runtime metrics to a Prometheus registry.
//
// Parameters:
//   - registry: Prometheus registry for metrics registration.
//
// Returns:
//   - err: Error if registration fails; otherwise nil.
func (r *RuntimeMetrics) Register(registry *prometheus.Registry) (err error) {
	metrics := []prometheus.Collector{
		r.goroutines,
		r.gcPauses,
		r.malloc,
		r.frees,
		r.gcCount,
		r.heapObjects,
	}

	for _, metric := range metrics {
		if err = registry.Register(metric); err != nil {
			break
		}
	}
	return err
}

// Start initiates periodic collection of runtime metrics.
//
// Parameters:
//   - interval: Frequency at which runtime metrics are updated.
func (r *RuntimeMetrics) Start(interval time.Duration) {
	r.wg.Add(1)

	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-r.ctx.Done():
				return
			case <-ticker.C:
				r.collectMetrics()
			}
		}
	}()
}

// collectMetrics retrieves and updates runtime metrics with current data.
//
// Implementation Detail: After reading the memory statistics, it updates the metrics and stores the current values
// for subsequent delta calculations.
func (r *RuntimeMetrics) collectMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	r.goroutines.Set(float64(runtime.NumGoroutine()))
	r.heapObjects.Set(float64(memStats.HeapObjects))

	// Update Counters with deltas
	r.malloc.Add(float64(memStats.Mallocs - r.prevMalloc))
	r.frees.Add(float64(memStats.Frees - r.prevFrees))
	r.gcCount.Add(float64(memStats.NumGC - r.prevGCCount))

	r.prevMalloc, r.prevFrees, r.prevGCCount = memStats.Mallocs, memStats.Frees, memStats.NumGC

	// Observe recent GC pause
	if memStats.NumGC > 0 {
		lastPause := memStats.PauseNs[(memStats.NumGC-1)%uint32(len(memStats.PauseNs))]
		r.gcPauses.Observe(float64(lastPause) / 1e9)
	}
}
