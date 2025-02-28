package messages

import (
	"context"
	"encoding/json"
	"log/slog"
	"shared/grpc/clients/nats_service"
	"shared/grpc/clients/nats_service/messaging"
	"time"
	"url-service/domain/entities"
	"url-service/domain/interfaces"
)

// InboundMessageService coordinates processing of URL messages received from a NATS subject.
type InboundMessageService struct {
	natsClient    *nats_service.NatsClient // natsClient is used for NATS subscriptions and publishing.
	urlRepository interfaces.UrlRepository // urlRepository is used for interacting with the persistence layer.
	batchSize     int                      // batchSize determines the max. number of URL processing goroutines.
	semaphore     chan struct{}            // semaphore is used to limit the number of processing goroutines.
	queueGroup    string                   // queueGroup is the NATS queue group for load balancing.
	logger        *slog.Logger
}

// NewInboundMessageService creates a new instance of InboundMessageService.
func NewInboundMessageService(
	natsClient *nats_service.NatsClient,
	urlRepository interfaces.UrlRepository,
	batchSize int,
	queueGroup string,
	logger *slog.Logger,
) *InboundMessageService {
	return &InboundMessageService{
		natsClient:    natsClient,
		urlRepository: urlRepository,
		batchSize:     batchSize,
		semaphore:     make(chan struct{}, batchSize),
		queueGroup:    queueGroup,
		logger:        logger,
	}
}

// Start subscribes to the UrlIncoming subject and processes incoming URL messages.
func (s *InboundMessageService) Start(ctx context.Context) (err error) {
	return s.natsClient.Subscribe(ctx, messaging.UrlIncoming, s.queueGroup, s.messageHandler)
}

// messageHandler is the callback function that processes each incoming message.
func (s *InboundMessageService) messageHandler(data []byte, subject string) {
	s.semaphore <- struct{}{}

	// Process a message.
	go func(data []byte, subject string) {
		defer func() { <-s.semaphore }()
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Panic recovered in message handler", "subject", subject, "panic", r)
			}
		}()

		// Workload
		s.logger.Info("Received message", "subject", subject, "data", string(data))
		var (
			unmarshalErr error
			err          error
			url          = entities.GetUrl()
			now          = time.Now()
			saveCtx      context.Context
			cancel       context.CancelFunc
		)
		defer url.Release()

		saveCtx, cancel = context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
		defer cancel()

		if unmarshalErr = json.Unmarshal(data, url); unmarshalErr != nil {
			s.logger.Error("JSON unmarshal failed", "subject", subject, "error", unmarshalErr)
			return
		}

		url.Status = entities.StatusPending
		url.CreatedAt = now
		if err = s.urlRepository.Save(saveCtx, url); err != nil {
			s.logger.Error("Failed to save URL", "subject", subject, "error", err)
			return
		}
		s.logger.Info("Successfully saved URL", "url", url)
	}(data, subject)
}
