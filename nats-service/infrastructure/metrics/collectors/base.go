package collectors

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector defines the interface for Prometheus metrics collectors.
//
// Methods:
//   - InitMetrics: Initializes and registers metrics with a Prometheus registry.
//   - Start: Starts periodic metrics collection at the specified interval.
//   - StopWithTimeout: Gracefully stops metrics collection with a given timeout.
type Collector interface {
	InitMetrics(registry *prometheus.Registry) (err error)
	Start(interval time.Duration)
	StopWithTimeout(timeout time.Duration)
}

// BaseCollector provides shared lifecycle management functionalities for metric collectors.
//
// Fields:
//   - name:   Human-readable collector name for logging purposes.
//   - ctx:    Context used for signaling cancellation to goroutines.
//   - cancel: Function to cancel the context, triggering collector shutdown.
//   - wg:     WaitGroup to ensure all goroutines finish gracefully.
//   - logger: Structured logger for lifecycle and error logging.
type BaseCollector struct {
	name   string
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	logger *slog.Logger
}

// InitBase initializes BaseCollector with a cancellable context, structured logger, and collector name.
//
// Parameters:
//   - name:   Name of the specific collector.
//   - logger: Structured logger instance for collector lifecycle logging.
func (b *BaseCollector) InitBase(name string, logger *slog.Logger) {
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.logger = logger
	b.name = name
}

// Done provides the cancellation channel from the context.
//
// Returns:
//   - <-chan struct{}: A channel closed when the context is canceled.
func (b *BaseCollector) Done() <-chan struct{} {
	return b.ctx.Done()
}

// StopWithTimeout gracefully stops metric collection, waiting until either all goroutines
// have terminated or the specified timeout elapses.
//
// Parameters:
//   - timeout: Duration to wait before forcibly exiting the goroutines.
func (b *BaseCollector) StopWithTimeout(timeout time.Duration) {
	b.cancel()
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		b.logger.Info("Collector stopped gracefully", slog.String("collector", b.name))
	case <-time.After(timeout):
		b.logger.Warn("Collector stop timed out", slog.String("collector", b.name))
	}
}
