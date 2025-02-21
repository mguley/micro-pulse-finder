package nats_service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNatsClient_Publish_Subscribe verifies that a published message is correctly received via subscription.
func TestNatsClient_Publish_Subscribe(t *testing.T) {
	var (
		env     = SetupTestEnvironment(t)
		subject = "test.subject"
		data    = []byte("Hello Nats")
		err     error
	)

	// Publish a message to the subject.
	err = env.Client.Publish(context.Background(), subject, data)
	require.NoError(t, err, "Expected successful publish")

	var (
		received = make(chan []byte, 1)
		subErr   = make(chan error, 1)
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()

	go func() {
		err = env.Client.Subscribe(ctx, subject, "", func(data []byte, topic string) {
			assert.Equal(t, subject, topic, "Subject mismatch")
			received <- data
		})
		subErr <- err
	}()

	// Wait for either a received message, an error, or a timeout.
	select {
	case msg := <-received:
		assert.Equal(t, data, msg, "Received message does not match published data")
		close(received)
	case err = <-subErr:
		close(subErr)
		t.Fatalf("Subscription failed: %v", err)
	case <-ctx.Done():
		t.Fatal("Did not receive the published message in time")
	}
}

// TestNatsClient_Publish_EmptySubject ensures that publishing a message with an empty subject returns an error.
func TestNatsClient_Publish_EmptySubject(t *testing.T) {
	var (
		env     = SetupTestEnvironment(t)
		subject = ""
		data    = []byte("Message without a subject")
		err     error
	)

	err = env.Client.Publish(context.Background(), subject, data)
	assert.Error(t, err, "Expected error when publishing with an empty subject")
	assert.Contains(t, err.Error(), "subject required", "Error message mismatch")
}

// TestNatsClient_Publish_EmptyData ensures that publishing a message with no data returns an error.
func TestNatsClient_Publish_EmptyData(t *testing.T) {
	var (
		env     = SetupTestEnvironment(t)
		subject = "test.empty.data"
		err     error
	)

	err = env.Client.Publish(context.Background(), subject, nil)
	assert.Error(t, err, "Expected error when publishing empty data")
	assert.Contains(t, err.Error(), "data required", "Error message mismatch")
}

// TestNatsClient_Subscribe_WithQueueGroup tests that subscribing with a queue group works as expected.
func TestNatsClient_Subscribe_WithQueueGroup(t *testing.T) {
	var (
		env        = SetupTestEnvironment(t)
		subject    = "test.queue.group"
		data       = []byte("Message for queue group")
		queueGroup = "testGroup"
		err        error
	)

	err = env.Client.Publish(context.Background(), subject, data)
	require.NoError(t, err, "Expected successful publish")

	var (
		received = make(chan []byte, 1)
		subErr   = make(chan error, 1)
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		err = env.Client.Subscribe(ctx, subject, queueGroup, func(data []byte, topic string) {
			assert.Equal(t, subject, topic, "subject mismatch")
			received <- data
		})
		subErr <- err
	}()

	// Wait for the message or an error.
	select {
	case msg := <-received:
		assert.Equal(t, data, msg, "Received message does not match published data")
		close(received)
	case err = <-subErr:
		close(subErr)
		t.Fatalf("Subscription failed: %v", err)
	case <-ctx.Done():
		t.Fatal("Did not receive the published message in time")
	}
}

// TestNatsClient_Subscribe_MultipleSubscribers verifies that multiple subscribers receive the published message.
func TestNatsClient_Subscribe_MultipleSubscribers(t *testing.T) {
	var (
		err     error
		subject = "test.multiple.subscribers"
		data    = []byte("Message for multiple subscribers")
		env     = SetupTestEnvironment(t)
	)

	// Publish the message
	err = env.Client.Publish(context.Background(), subject, data)
	require.NoError(t, err, "Expected successful publish")

	type result struct {
		index int
		msg   []byte
	}

	var (
		counter = 5
		results = make(chan result, counter)
		errs    = make(chan error, counter)
		wg      sync.WaitGroup
		ctx     context.Context
		cancel  context.CancelFunc
	)

	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()

	wg.Add(counter)
	for i := 0; i < counter; i++ {
		go func(index int) {
			defer wg.Done()

			var (
				subErr  error
				handler = func(data []byte, topic string) {
					assert.Equal(t, subject, topic, "subject mismatch")
					results <- result{index: index, msg: data}
				}
			)
			if subErr = env.Client.Subscribe(ctx, subject, "", handler); subErr != nil {
				errs <- subErr
				return
			}
		}(i)
	}
	wg.Wait()
	close(results)
	close(errs)

	for err = range errs {
		require.NoError(t, err, "Expected no error")
	}

	for res := range results {
		assert.Equal(t, data, res.msg, "Received message does not match published data")
		t.Logf("Received message at index %d: %s", res.index, string(res.msg))
	}
}
