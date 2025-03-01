package services

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"proxy-service/infrastructure/http/socks5"
	"shared/grpc/clients/nats_service"
	"shared/grpc/clients/nats_service/messaging"
	"time"
)

// UrlProcessorService coordinates processing of URL messages received from a NATS subject.
type UrlProcessorService struct {
	pool       *socks5.ConnectionPool   // pool is the connection pool used to borrow/return HTTP clients.
	natsClient *nats_service.NatsClient // natsClient is used for NATS subscriptions and publishing.
	batchSize  int                      // batchSize determines the max. number of concurrent URL processing goroutines.
	semaphore  chan struct{}            // semaphore is used to limit the number of concurrently processing goroutines.
	queueGroup string                   // queueGroup is the NATS queue group for load balancing.
	logger     *slog.Logger             // logger for structured logging.
}

// NewUrlProcessorService creates a new instance of UrlProcessorService.
func NewUrlProcessorService(
	pool *socks5.ConnectionPool,
	natsClient *nats_service.NatsClient,
	batchSize int,
	queueGroup string,
	logger *slog.Logger,
) *UrlProcessorService {
	return &UrlProcessorService{
		pool:       pool,
		natsClient: natsClient,
		batchSize:  batchSize,
		queueGroup: queueGroup,
		semaphore:  make(chan struct{}, batchSize),
		logger:     logger,
	}
}

// Start subscribes to the ProxyUrlRequest subject and processes incoming URL messages.
func (s *UrlProcessorService) Start(ctx context.Context) (err error) {
	s.logger.Info("Starting URL processor service", "queueGroup", s.queueGroup)
	return s.natsClient.Subscribe(ctx, messaging.ProxyUrlRequest, s.queueGroup, s.messageHandler)
}

// messageHandler is the callback function that processes each incoming message.
// It validates the URL, makes an HTTP GET request using a borrowed client from the connection pool,
// and publishes the response body to the ProxyUrlResponse subject.
func (s *UrlProcessorService) messageHandler(data []byte, subject string) {
	// Acquire a semaphore slot.
	s.semaphore <- struct{}{}

	// Process a message.
	go func(data []byte, subject string) {
		defer func() { <-s.semaphore }()
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Recovered from panic in URL message processing", "subject", subject, "panic", r)
			}
		}()

		// Workload
		var (
			rawURL     = string(data)
			client     *http.Client
			parsedURL  *url.URL
			request    *http.Request
			response   *http.Response
			requestCtx context.Context
			cancel     context.CancelFunc
			body       []byte
			err        error
		)

		s.logger.Info("Processing URL", "url", rawURL, "subject", subject)

		// Validate that URL is well-formed.
		if parsedURL, err = url.ParseRequestURI(rawURL); err != nil {
			s.logger.Error("Invalid URL received", "url", rawURL, "error", err)
			return
		}

		// Borrow HTTP client from the pool.
		client = s.pool.Borrow()
		defer s.pool.Return(client)

		requestCtx, cancel = context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
		defer cancel()

		// Create and execute HTTP request.
		request, err = http.NewRequestWithContext(requestCtx, http.MethodGet, parsedURL.String(), http.NoBody)
		if err != nil {
			s.logger.Error("Could not create HTTP request", "url", parsedURL.String(), "error", err)
			return
		}

		if response, err = client.Do(request); err != nil {
			s.logger.Error("Could not make HTTP request", "url", parsedURL.String(), "error", err)
			return
		}
		defer func() {
			if closeErr := response.Body.Close(); closeErr != nil {
				s.logger.Error("Could not close response body", "url", parsedURL.String(), "error", closeErr)
			}
		}()

		// Process the response.
		if body, err = io.ReadAll(response.Body); err != nil {
			s.logger.Error("Could not read response body", "url", parsedURL.String(), "error", err)
			return
		}
		if err = s.natsClient.Publish(request.Context(), messaging.ProxyUrlResponse, body); err != nil {
			s.logger.Error("Could not publish URL response", "url", parsedURL.String(), "error", err)
			return
		}

		s.logger.Info("Successfully processed URL", "url", parsedURL.String())
	}(data, subject)
}
