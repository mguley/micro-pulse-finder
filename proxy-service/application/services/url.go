package services

import (
	"context"
	"fmt"
	"io"
	"log"
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
}

// NewUrlProcessorService creates a new instance of UrlProcessorService.
func NewUrlProcessorService(
	pool *socks5.ConnectionPool,
	natsClient *nats_service.NatsClient,
	batchSize int,
	queueGroup string,
) *UrlProcessorService {
	return &UrlProcessorService{
		pool:       pool,
		natsClient: natsClient,
		batchSize:  batchSize,
		queueGroup: queueGroup,
		semaphore:  make(chan struct{}, batchSize),
	}
}

// Start subscribes to the ProxyUrlRequest subject and processes incoming URL messages.
func (s *UrlProcessorService) Start(ctx context.Context) (err error) {
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
				fmt.Printf("[PANIC] Recovered in goroutine for subject %s: %v\n", subject, r)
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

		// Validate that URL is well-formed.
		if parsedURL, err = url.ParseRequestURI(rawURL); err != nil {
			log.Printf("Invalid URL received: %s, error: %v", rawURL, err)
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
			log.Printf("Could not create HTTP request for URL %s, error: %v", parsedURL.String(), err)
			return
		}
		if response, err = client.Do(request); err != nil {
			log.Printf("Could not make HTTP request to %s, error: %v", parsedURL.String(), err)
			return
		}
		defer func() {
			if closeErr := response.Body.Close(); closeErr != nil {
				log.Printf("Could not close response body from %s: %v", parsedURL.String(), closeErr)
			}
		}()

		// Process the response.
		if body, err = io.ReadAll(response.Body); err != nil {
			log.Printf("Could not read response body from %s: %v", parsedURL.String(), err)
			return
		}
		if err = s.natsClient.Publish(request.Context(), messaging.ProxyUrlResponse, body); err != nil {
			log.Printf("Could not publish message for URL %s, error: %v", parsedURL.String(), err)
			return
		}
	}(data, subject)
}
