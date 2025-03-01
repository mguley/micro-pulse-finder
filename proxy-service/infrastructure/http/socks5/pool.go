package socks5

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// CreatorFunc defines a function signature that returns a new *http.Client or an error.
type CreatorFunc func() (client *http.Client, err error)

// ConnectionPool manages a pool of HTTP clients configured to use a SOCKS5 proxy.
type ConnectionPool struct {
	pool          chan *http.Client // pool holds available HTTP clients.
	mu            sync.Mutex        // mu protects concurrent access during refresh and shutdown.
	maxPoolSize   int               // maxPoolSize is the maximum number of connections in the pool.
	refreshTicker *time.Ticker      // refreshTicker triggers periodic refreshes of idle connections.
	stopChan      chan struct{}     // stopChan signals the refresh goroutine to stop.
	creator       CreatorFunc       // creator is a function that returns a new HTTP client.
	shutdownOnce  sync.Once         // shutdownOnce ensures Shutdown is executed only once.
	logger        *slog.Logger
}

// NewConnectionPool creates a new instance of ConnectionPool.
func NewConnectionPool(
	poolSize int,
	refreshInterval time.Duration,
	creator CreatorFunc,
	logger *slog.Logger,
) *ConnectionPool {
	pool := &ConnectionPool{
		pool:        make(chan *http.Client, poolSize),
		maxPoolSize: poolSize,
		stopChan:    make(chan struct{}),
		creator:     creator,
		logger:      logger,
	}

	pool.initialize(refreshInterval)
	return pool
}

// initialize creates initial connections and starts the periodic refresh routine.
func (cp *ConnectionPool) initialize(refreshInterval time.Duration) {
	var (
		httpClient *http.Client
		err        error
	)

	cp.logger.Info("Initializing connection pool", "maxPoolSize", cp.maxPoolSize)

	for i := 0; i < cp.maxPoolSize; i++ {
		if httpClient, err = cp.creator(); err != nil {
			cp.logger.Error("Could not create HTTP client for connection pool", "error", err)
			cp.logger.Warn("Panic due to failure in connection pool initialization")
			panic(fmt.Sprintf("could not create HTTP client for connection pool: %v", err))
		}
		cp.pool <- httpClient
	}

	// Start the periodic refresh routine.
	cp.refreshTicker = time.NewTicker(refreshInterval)
	go cp.startRefresh()
	cp.logger.Info("Connection pool initialized and refresh routine started", "refreshInterval", refreshInterval)
}

// startRefresh periodically refreshes connections in the pool to ensure they remain healthy.
func (cp *ConnectionPool) startRefresh() {
	for {
		select {
		case <-cp.stopChan:
			cp.logger.Info("Stopping connection pool refresh routine")
			return
		case <-cp.refreshTicker.C:
			cp.refreshConnections()
		}
	}
}

// refreshConnections drains idle connections from the pool, closes their idle connections,
// and replaces them with newly created ones. Connections currently borrowed are left intact.
func (cp *ConnectionPool) refreshConnections() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Drain all idle connections from the pool.
	var (
		httpClient *http.Client
		err        error
		idleCount  = len(cp.pool)
	)

	cp.logger.Debug("Refreshing idle connections", "idleCount", idleCount)

	for i := 0; i < idleCount; i++ {
		select {
		case client := <-cp.pool:
			if transport, ok := client.Transport.(*http.Transport); ok {
				transport.CloseIdleConnections()
			}
			if httpClient, err = cp.creator(); err != nil {
				cp.logger.Error("Could not refresh connection", "error", err)
			} else {
				cp.pool <- httpClient
			}
		default:
			cp.logger.Debug("No more idle connections to refresh")
			return
		}
	}
}

// Borrow retrieves an available HTTP client from the pool.
func (cp *ConnectionPool) Borrow() (client *http.Client) {
	client = <-cp.pool
	cp.logger.Debug("HTTP client borrowed from pool")
	return client
}

// Return places an HTTP client back into the pool for reuse.
func (cp *ConnectionPool) Return(client *http.Client) {
	cp.pool <- client
	cp.logger.Debug("HTTP client returned to pool")
}

// Shutdown gracefully stops the connection pool's refresh routine and cleans up resources.
func (cp *ConnectionPool) Shutdown() {
	cp.shutdownOnce.Do(func() {
		cp.mu.Lock()
		defer cp.mu.Unlock()

		// Signal the refresh goroutine to stop and stop the ticker.
		cp.logger.Info("Shutting down connection pool")
		close(cp.stopChan)
		cp.refreshTicker.Stop()

		// Close the pool channel to prevent further usage.
		close(cp.pool)
		// Drain the pool and close idle connections.
		for client := range cp.pool {
			if transport, ok := client.Transport.(*http.Transport); ok {
				transport.CloseIdleConnections()
			}
		}
		cp.logger.Info("Connection pool shutdown complete")
	})
}
