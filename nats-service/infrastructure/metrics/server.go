package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server provides an HTTP server to expose Prometheus metrics and profiling endpoints.
//
// Fields:
//   - server:          HTTP server instance serving metrics endpoints.
//   - metricsProvider: Provider managing registered Prometheus collectors.
//   - logger:          Structured logger for server lifecycle logging.
type Server struct {
	server          *http.Server
	metricsProvider *Provider
	logger          *slog.Logger
}

// NewServer initializes and configures a new metrics HTTP server.
//
// Parameters:
//   - port:            String specifying the port number (e.g., ":9090").
//   - metricsProvider: Initialized Provider with registered Prometheus collectors.
//   - logger:          Logger instance for structured logging.
//
// Returns:
//   - *Server: A configured HTTP metrics server ready to start.
func NewServer(port string, metricsProvider *Provider, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/nats-service/metrics", promhttp.HandlerFor(
		metricsProvider.Registry,
		promhttp.HandlerOpts{EnableOpenMetrics: true}),
	)

	// pprof endpoints for runtime profiling
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			return
		}
	})

	// Initialize HTTP server
	httpServer := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(5) * time.Second,
		WriteTimeout:      time.Duration(5) * time.Second,
		IdleTimeout:       time.Duration(10) * time.Second,
	}

	return &Server{
		server:          httpServer,
		metricsProvider: metricsProvider,
		logger:          logger,
	}
}

// Start launches the HTTP metrics server.
//
// Returns:
//   - err: An error if the server fails to start; nil otherwise.
func (s *Server) Start() (err error) {
	s.logger.Info("Starting metrics server", slog.String("address", s.server.Addr))
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the HTTP metrics server.
//
// Parameters:
//   - ctx: Context for graceful shutdown timeout.
//
// Returns:
//   - err: An error if shutdown is unsuccessful; nil otherwise.
func (s *Server) Stop(ctx context.Context) (err error) {
	s.logger.Info("Stopping metrics server", slog.String("address", s.server.Addr))
	return s.server.Shutdown(ctx)
}
