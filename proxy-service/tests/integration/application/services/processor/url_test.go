package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"shared/grpc/clients/nats_service/messaging"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestUrlProcessorService_ValidURL verifies that when a valid URL is published on the
// ProxyUrlRequest subject, the URL processor makes an HTTP GET request to the URL
// and then publishes a valid response (in JSON) on the ProxyUrlResponse subject.
func TestUrlProcessorService_ValidURL(t *testing.T) {
	container, teardown := SetupTestContainer()
	defer teardown()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a subscriber on the ProxyUrlResponse subject to capture responses.
	responseChan := make(chan []byte, 1)
	subCtx, subCancel := context.WithCancel(ctx)
	defer subCancel()
	go func() {
		var (
			err            error
			subject        = messaging.ProxyUrlResponse
			queueGroup     = container.Config.Get().UrlProcessor.QueueGroup
			messageHandler = func(data []byte, subject string) { responseChan <- data }
			natsClient     = container.NatsGrpcClient.Get()
		)

		if err = natsClient.Subscribe(subCtx, subject, queueGroup, messageHandler); err != nil {
			t.Logf("Could not subscribe to the ProxyUrlResponse subject: %v", err)
			return
		}
	}()

	// Allow a brief moment for the subscriber to be established.
	time.Sleep(time.Duration(2) * time.Second)

	// Publish a valid URL to the ProxyUrlRequest subject.
	validURL := "https://httpbin.org/ip"
	err := container.NatsGrpcClient.Get().Publish(ctx, messaging.ProxyUrlRequest, []byte(validURL))
	require.NoError(t, err, "Failed to publish valid URL message.")

	// Wait for the response from the URL processor.
	select {
	case response := <-responseChan:
		require.NotEmpty(t, response, "Expected non-empty response from the URL processor")

		// Parse the response JSON to check for expected content.
		type ipResponse struct {
			Origin string `json:"origin"`
		}
		var ipData ipResponse
		err = json.Unmarshal(response, &ipData)
		require.NoError(t, err, "Failed to parse JSON response")
		require.NotEmpty(t, ipData.Origin, "Expected non-empty origin from the URL processor")
		t.Logf("Received response from the URL processor: %s", ipData.Origin)
		close(responseChan)
	case <-time.After(time.Duration(15) * time.Second):
		t.Fatal("Timeout waiting for response from the URL processor")
	}
}

// TestUrlProcessorService_ConcurrentURLs verifies that the URL processor service
// can handle multiple concurrent URL messages. It publishes several valid URLs
// concurrently to the ProxyUrlRequest subject and then collects responses from the
// ProxyUrlResponse subject.
func TestUrlProcessorService_ConcurrentURLs(t *testing.T) {
	container, teardown := SetupTestContainer()
	defer teardown()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a subscriber on the ProxyUrlResponse subject to capture responses.
	responseChan := make(chan []byte, 10)
	subCtx, subCancel := context.WithCancel(ctx)
	defer subCancel()
	go func() {
		var (
			err            error
			subject        = messaging.ProxyUrlResponse
			queueGroup     = container.Config.Get().UrlProcessor.QueueGroup
			messageHandler = func(data []byte, subject string) { responseChan <- data }
			natsClient     = container.NatsGrpcClient.Get()
		)

		if err = natsClient.Subscribe(subCtx, subject, queueGroup, messageHandler); err != nil {
			t.Logf("Could not subscribe to the ProxyUrlResponse subject: %v", err)
			return
		}
	}()

	// Allow the subscriber to establish.
	time.Sleep(time.Duration(2) * time.Second)

	// Define a slice of valid URLs to be processed concurrently.
	urls := []string{
		"https://httpbin.org/ip",
		"https://httpbin.org/get",
		"https://httpbin.org/uuid",
		"https://httpbin.org/user-agent",
		"https://httpbin.org/headers",
	}

	// Publish all URLs concurrently.
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			err := container.NatsGrpcClient.Get().Publish(ctx, messaging.ProxyUrlRequest, []byte(url))
			require.NoError(t, err, fmt.Sprintf("Failed to publish valid URL message: %s", url))
		}(url)
	}
	wg.Wait()

	// Collect responses until the number of responses matches the number of URLs.
	responses := make([][]byte, 0, len(urls))
	timeout := time.After(time.Duration(30) * time.Second)
	for len(responses) < len(urls) {
		select {
		case response := <-responseChan:
			responses = append(responses, response)
		case <-timeout:
			t.Fatalf("Timeout waiting for responses: got %d, expected %d", len(responses), len(urls))
		}
	}

	// Validate that each response is a valid JSON and contains data.
	for i, response := range responses {
		var data map[string]any
		err := json.Unmarshal(response, &data)
		require.NoError(t, err, "Response %d is not a valid JSON", i)
		require.NotEmpty(t, data, "Response %d is empty", i)
		t.Logf("Response %d: %v", i, data)
	}
}
