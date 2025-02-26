package main

import (
	"context"
	"errors"
	"log"
	"nats-service/application"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting NATS metrics server...")

	var (
		mux      = http.NewServeMux()
		registry = application.NewContainer().Infrastructure.Get().Metrics.Get().Registry
		server   *http.Server
		port     = ":50555"
	)

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
		log.Printf("Metrics server is listening on %s...", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received. Shutting down metrics server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Metrics server shutdown error: %v", err)
	}
}
