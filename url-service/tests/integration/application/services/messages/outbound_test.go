package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"shared/grpc/clients/nats_service/messaging"
	"sync"
	"testing"
	"time"
	"url-service/domain/entities"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// TestOutboundMessageService_ProcessMessage verifies that when a pending URL exists in MongoDB,
// the OutboundMessageService publishes its JSON representation to the outbound NATS subject and
// subsequently updates its status to processed.
func TestOutboundMessageService_ProcessMessage(t *testing.T) {
	container, teardown := SetupTestContainer(t)
	defer teardown()

	// Set up a subscriber on the UrlOutgoing subject to capture responses.
	responseChan := make(chan []byte, 1)
	subCtx, subCancel := context.WithCancel(context.Background())
	defer subCancel()
	go func() {
		var (
			err            error
			subject        = messaging.UrlOutgoing
			queueGroup     = ""
			messageHandler = func(data []byte, subject string) { responseChan <- data }
			natsClient     = container.NatsGrpcClient.Get()
		)

		if err = natsClient.Subscribe(subCtx, subject, queueGroup, messageHandler); err != nil {
			t.Logf("Could not subscribe to the UrlOutgoing subject: %v", err)
			return
		}
	}()

	// Insert a pending URL entity into MongoDB.
	var (
		err        error
		repository = container.MongoRepository.Get()
		testUrl    = "https://example.com/outbound"
		now        = time.Now()
		urlEntity  = &entities.Url{
			Address:   testUrl,
			Status:    entities.StatusPending,
			Source:    "integration_test_outbound",
			CreatedAt: now,
			UpdatedAt: now,
		}
	)
	err = repository.Save(context.Background(), urlEntity)
	require.NoError(t, err, "Failed to save URL entity to MongoDB")

	// Wait for the outbound service to process the message and publish it.
	timeout := time.Duration(15) * time.Second
	select {
	case response := <-responseChan:
		var publishedUrl entities.Url
		err = json.Unmarshal(response, &publishedUrl)
		require.NoError(t, err, "Failed to unmarshal published message")
		require.Equal(t, testUrl, publishedUrl.Address, "Published URL address mismatch")
		close(responseChan)
	case <-time.After(timeout):
		t.Fatal("Timeout waiting for outbound message to be published")
	}

	// Poll MongoDB until the URL status is updated to the processed status.
	var (
		fetchedUrls []*entities.Url
		filter      = bson.M{"address": testUrl}
		delay       = time.After(time.Duration(30) * time.Second)
	)
	for {
		fetchedUrls, err = repository.FetchBatch(context.Background(), filter, 5)
		require.NoError(t, err, "Failed to fetch URLs from MongoDB")
		if len(fetchedUrls) > 0 && fetchedUrls[0].Status == entities.StatusProcessed {
			break
		}

		select {
		case <-delay:
			t.Fatalf("Timeout waiting for URLs to be processed")
		default:
			time.Sleep(time.Duration(2) * time.Second)
		}
	}

	require.NotEmpty(t, fetchedUrls, "Failed to fetch URLs from MongoDB")
	require.Equal(t, entities.StatusProcessed, fetchedUrls[0].Status, "Fetched URL status mismatch")
}

// TestOutboundMessageService_ConcurrentProcessing verifies that the outbound service can handle
// multiple pending URL entities concurrently.
func TestOutboundMessageService_ConcurrentProcessing(t *testing.T) {
	container, teardown := SetupTestContainer(t)
	defer teardown()

	var (
		numMessages = 15
		wg          sync.WaitGroup
		repository  = container.MongoRepository.Get()
	)

	// Insert multiple pending URL entities.
	wg.Add(numMessages)
	for i := 0; i < numMessages; i++ {
		go func(i int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("WARNING: panic recovered: %v\n", r)
				}
			}()

			// Workload
			var (
				err     error
				url     = entities.GetUrl()
				testUrl = fmt.Sprintf("https://example.com/outbound/%d", i)
				now     = time.Now()
			)
			defer url.Release()

			url.Address = testUrl
			url.Status = entities.StatusPending
			url.Source = "concurrent_test_outbound"
			url.CreatedAt = now
			url.UpdatedAt = now
			err = repository.Save(context.Background(), url)
			require.NoError(t, err, "Failed to save URL entity to MongoDB")
		}(i)
	}
	wg.Wait()

	// Wait until all URLs are processed (status updated to processed).
	var (
		filter  = bson.M{"source": "concurrent_test_outbound"}
		fetched []*entities.Url
		timeout = time.After(time.Duration(60) * time.Second)
		ticker  = time.NewTicker(time.Duration(2) * time.Second)
		err     error
	)
	defer ticker.Stop()

Loop:
	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for all outbound messages to be processed")
		case <-ticker.C:
			fetched, err = repository.FetchBatch(context.Background(), filter, numMessages)
			require.NoError(t, err, "Failed to fetch URLs from MongoDB")
			if len(fetched) == numMessages {
				// Ensure every URL has been marked as processed.
				allProcessed := true
				for _, url := range fetched {
					if url.Status != entities.StatusProcessed {
						allProcessed = false
						break
					}
				}
				if allProcessed {
					break Loop
				}
			}
		}
	}

	require.Len(t, fetched, numMessages, "Mismatch in expected number of processed URLs")
}
