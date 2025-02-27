package main

import (
	"context"
	"errors"
	"nats-service/application"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		app      = application.NewContainer()
		mux      = http.NewServeMux()
		registry = app.Infrastructure.Get().Metrics.Get().Registry
		logger   = app.Infrastructure.Get().Logger.Get()
		server   *http.Server
		port     = ":50555"
	)

	logger.Info("Starting NATS metrics server...")
	mux.Handle("/nats-service/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	server = &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(5) * time.Second,
		WriteTimeout:      time.Duration(10) * time.Second,
		IdleTimeout:       time.Duration(15) * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("Metrics server is listening ...", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Metrics server error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutdown signal received. Shutting down metrics server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Metrics server shutdown error", "error", err)
	}
}
