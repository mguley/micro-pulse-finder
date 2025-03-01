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

	"go.mongodb.org/mongo-driver/bson"
)

// OutboundMessageService periodically scans MongoDB for pending URL entities and pushes them to a NATS subject.
type OutboundMessageService struct {
	natsClient    *nats_service.NatsClient
	urlRepository interfaces.UrlRepository
	batchSize     int
	semaphore     chan struct{}
	interval      time.Duration
	logger        *slog.Logger
}

// NewOutboundMessageService creates a new instance of OutboundMessageService.
func NewOutboundMessageService(
	natsClient *nats_service.NatsClient,
	urlRepository interfaces.UrlRepository,
	interval time.Duration,
	batchSize int,
	logger *slog.Logger,
) *OutboundMessageService {
	return &OutboundMessageService{
		natsClient:    natsClient,
		urlRepository: urlRepository,
		batchSize:     batchSize,
		semaphore:     make(chan struct{}, batchSize),
		interval:      interval,
		logger:        logger,
	}
}

// Start begins the periodic scanning and publishing process.
func (s *OutboundMessageService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context canceled, outbound service stopped.")
			return
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

// scan retrieves for pending URL entities (up to the batchSize) and processes them.
func (s *OutboundMessageService) scan(ctx context.Context) {
	var (
		filter = bson.M{"status": entities.StatusPending}
		list   []*entities.Url
		err    error
	)

	if list, err = s.urlRepository.FetchBatch(ctx, filter, s.batchSize); err != nil {
		s.logger.Error("Failed to fetch pending URLs", "error", err)
		return
	}

	if len(list) == 0 {
		s.logger.Info("No pending URLs found")
		return
	}

	// Launch a goroutine for each URL while respecting the semaphore limit.
	for _, url := range list {
		s.semaphore <- struct{}{}
		go s.processMessage(ctx, url)
	}
}

// processMessage serializes URL entity, publishes it to a NATS subject, and updates its status.
func (s *OutboundMessageService) processMessage(ctx context.Context, url *entities.Url) {
	defer func() { <-s.semaphore }()
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic recovered in processMessage", "urlID", url.Id.Hex(), "panic", r)
		}
	}()

	// Workload
	var (
		data       []byte
		marshalErr error
		pubErr     error
		updateErr  error
	)

	if data, marshalErr = json.Marshal(url); marshalErr != nil {
		s.logger.Error("Failed to marshal URL", "urlID", url.Id.Hex(), "error", marshalErr)
		return
	}
	if pubErr = s.natsClient.Publish(ctx, messaging.UrlOutgoing, data); pubErr != nil {
		s.logger.Error("Failed to publish URL", "urlID", url.Id.Hex(), "error", pubErr)
		return
	}
	s.logger.Info("Published URL", "urlID", url.Id.Hex(), "subject", messaging.UrlOutgoing)

	// Update the URL's status to processed to avoid republishing.
	now := time.Now()
	updateFields := bson.M{
		"status":     entities.StatusProcessed,
		"processed":  now,
		"updated_at": now,
	}
	if updateErr = s.urlRepository.UpdateFields(ctx, url.Id.Hex(), updateFields); updateErr != nil {
		s.logger.Error("Failed to update URL", "urlID", url.Id.Hex(), "error", updateErr)
		return
	}

	s.logger.Info("Updated URL", "urlID", url.Id.Hex(), "updateFields", updateFields)
}
