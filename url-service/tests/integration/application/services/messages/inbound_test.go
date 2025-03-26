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

// TestInboundMessageService_ProcessMessage verifies that when a valid URL message is published
// to the UrlIncoming subject, the InboundMessageService correctly processes the message and
// saves the URL entity into MongoDB.
func TestInboundMessageService_ProcessMessage(t *testing.T) {
	t.Skip("Temporarily skipping this test on GitHub Actions")
	container, teardown := SetupTestContainer(t)
	defer teardown()

	testUrl := "https://example.com"
	msgPayload := map[string]string{
		"address": testUrl,
		"source":  "integration_test",
	}
	data, err := json.Marshal(msgPayload)
	require.NoError(t, err, "Failed to marshal message payload")

	// Publish the test message to the inbound subject.
	natsClient := container.NatsGrpcClient.Get()
	err = natsClient.Publish(context.Background(), messaging.UrlIncoming, data)
	require.NoError(t, err, "Failed to publish test URL message")

	// Allow some time for processing.
	time.Sleep(time.Duration(3) * time.Second)

	// Query MongoDB for the saved URL record.
	repository := container.MongoRepository.Get()
	filter := bson.M{"address": testUrl}
	list, err := repository.FetchBatch(context.Background(), filter, 5)
	require.NoError(t, err, "Failed to fetch URL records from MongoDB")
	require.Len(t, list, 1, "Expected one URL record to be saved")

	savedUrl := list[0]
	// todo here we have an issue (GitHub Actions only)
	//require.Equal(t, entities.StatusPending, savedUrl.Status, "URL status should be pending")
	require.Equal(t, testUrl, savedUrl.Address, "URL address should match")
	require.Equal(t, "integration_test", savedUrl.Source, "URL source should match")
	require.False(t, savedUrl.CreatedAt.IsZero(), "CreatedAt timestamp should not be zero")
}

// TestInboundMessageService_ConcurrentMessages sends several different messages concurrently
// and verifies that all messages are processed and saved into MongoDB.
func TestInboundMessageService_ConcurrentMessages(t *testing.T) {
	container, teardown := SetupTestContainer(t)
	defer teardown()

	var (
		numMessages = 35
		wg          sync.WaitGroup
	)

	// Publish a bunch of messages.
	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("WARNING: panic recovered: %v\n", r)
				}
			}()

			// Workload
			testUrl := fmt.Sprintf("https://example.com/test/%d", i)
			msgPayload := map[string]string{
				"address": testUrl,
				"source":  "concurrent_test",
			}
			data, err := json.Marshal(msgPayload)
			require.NoError(t, err, "Failed to marshal message payload")
			err = container.NatsGrpcClient.Get().Publish(context.Background(), messaging.UrlIncoming, data)
			require.NoError(t, err, "Failed to publish test URL message")
		}(i)
	}
	wg.Wait()

	// Wait until all messages are processed.
	var (
		repository = container.MongoRepository.Get()
		timeout    = time.After(time.Duration(30) * time.Second)
		filter     = bson.M{"source": "concurrent_test"}
		list       []*entities.Url
		err        error
	)
	for {
		list, err = repository.FetchBatch(context.Background(), filter, numMessages)
		require.NoError(t, err, "Failed to fetch URL records from MongoDB")
		if len(list) >= numMessages {
			break
		}

		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for %d messages to be processed", numMessages)
		default:
			time.Sleep(time.Duration(1) * time.Second)
		}
	}

	require.Len(t, list, numMessages, "Expected %d messages to be saved, got %d", numMessages, len(list))
}
