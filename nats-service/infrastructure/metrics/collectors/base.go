package collectors

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector defines the interface for Prometheus metrics collectors.
//
// Methods:
//   - Register: Registers collector metrics with a Prometheus registry.
//   - Start:    Starts metrics collection at a specified interval.
//   - Stop:     Stops metrics collection gracefully.
type Collector interface {
	Register(registry *prometheus.Registry) (err error)
	Start(interval time.Duration)
	Stop()
}

// BaseCollector provides shared lifecycle management for metric collectors.
//
// Fields:
//   - ctx:    Context used to signal goroutine cancellation.
//   - cancel: Function to cancel the context and stop metric collection.
//   - wg:     WaitGroup for ensuring graceful termination of collector goroutines.
type BaseCollector struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Init initializes BaseCollector with a cancellable context.
func (b *BaseCollector) Init() {
	b.ctx, b.cancel = context.WithCancel(context.Background())
}

// Done provides the context cancellation channel.
//
// Returns:
//   - <-chan struct{}: A channel that is closed when the context is canceled.
func (b *BaseCollector) Done() <-chan struct{} {
	return b.ctx.Done()
}

// Stop gracefully stops metric collection and waits for goroutines to exit.
func (b *BaseCollector) Stop() {
	b.cancel()
	b.wg.Wait()
}
