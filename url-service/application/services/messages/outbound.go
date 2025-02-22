package messages

import (
	"context"
	"encoding/json"
	"fmt"
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
}

// NewOutboundMessageService creates a new instance of OutboundMessageService.
func NewOutboundMessageService(
	natsClient *nats_service.NatsClient,
	urlRepository interfaces.UrlRepository,
	interval time.Duration,
	batchSize int,
) *OutboundMessageService {
	return &OutboundMessageService{
		natsClient:    natsClient,
		urlRepository: urlRepository,
		batchSize:     batchSize,
		semaphore:     make(chan struct{}, batchSize),
		interval:      interval,
	}
}

// Start begins the periodic scanning and publishing process.
func (s *OutboundMessageService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Outbound service stopped")
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
		fmt.Printf("[ERROR] Failed to fetch pending URLs: %v\n", err)
		return
	}

	if len(list) == 0 {
		fmt.Println("[INFO] No pending URLs found")
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
			fmt.Printf("[PANIC] Recovered in goroutine for URL ID %s: %v\n", url.Id.Hex(), r)
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
		fmt.Printf("[ERROR] Could not marshal URL with ID %s: %v\n", url.Id.Hex(), marshalErr)
		return
	}
	if pubErr = s.natsClient.Publish(ctx, messaging.UrlOutgoing, data); pubErr != nil {
		fmt.Printf("[ERROR] Could not publish URL with ID %s: %v\n", url.Id.Hex(), pubErr)
		return
	}
	fmt.Printf("[INFO] Published URL with ID %s to subject %s\n", url.Id.Hex(), messaging.UrlOutgoing)

	// Update the URL's status to processed to avoid republishing.
	now := time.Now()
	updateFields := bson.M{
		"status":     entities.StatusProcessed,
		"processed":  now,
		"updated_at": now,
	}
	if updateErr = s.urlRepository.UpdateFields(ctx, url.Id.Hex(), updateFields); updateErr != nil {
		fmt.Printf("[ERROR] Could not update URL with ID %s: %v\n", url.Id.Hex(), updateErr)
		return
	}
	fmt.Printf("[INFO] Updated URL with ID %s to processed status\n", url.Id.Hex())
}
