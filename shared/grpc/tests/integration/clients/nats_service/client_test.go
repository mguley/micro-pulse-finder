package nats_service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNatsClient_Publish_Subscribe(t *testing.T) {
	env := SetupTestEnvironment(t)

	subject := "test.subject"
	data := []byte("Hello Nats")

	// Publish a message
	err := env.Client.Publish(context.Background(), subject, data)
	require.NoError(t, err, "expected successful publish")

	// Channel to capture received messages
	received := make(chan []byte, 1)
	defer close(received)

	// Subscribe to the subject
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()

	subErr := make(chan error, 1)
	defer close(subErr)

	go func() {
		err = env.Client.Subscribe(ctx, subject, "", func(data []byte, topic string) {
			assert.Equal(t, subject, topic, "subject mismatch")
			received <- data
		})
		subErr <- err
	}()

	// Wait for the message to be received or timeout
	select {
	case msg := <-received:
		assert.Equal(t, data, msg, "received message does not match published data")
	case err = <-subErr:
		t.Fatalf("Subscription failed: %v", err)
	case <-ctx.Done():
		t.Fatal("Did not receive the published message in time")
	}
}

func TestNatsClient_Publish_EmptySubject(t *testing.T) {
	env := SetupTestEnvironment(t)

	subject := ""
	data := []byte("Message without a subject")

	err := env.Client.Publish(context.Background(), subject, data)
	assert.Error(t, err, "expected error when publishing with an empty subject")
	assert.Contains(t, err.Error(), "subject required", "error message mismatch")
}

func TestNatsClient_Publish_EmptyData(t *testing.T) {
	env := SetupTestEnvironment(t)

	subject := "test.empty.data"

	err := env.Client.Publish(context.Background(), subject, nil)
	assert.Error(t, err, "expected error when publishing empty data")
	assert.Contains(t, err.Error(), "data required", "error message mismatch")
}

func TestNatsClient_Subscribe_WithQueueGroup(t *testing.T) {
	env := SetupTestEnvironment(t)

	subject := "test.queue.group"
	data := []byte("Message for queue group")
	queueGroup := "testGroup"

	err := env.Client.Publish(context.Background(), subject, data)
	require.NoError(t, err, "expected successful publish")

	// Channel to capture received messages
	received := make(chan []byte, 1)
	defer close(received)

	// Subscribe with a queue group
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	subErr := make(chan error, 1)
	defer close(subErr)

	go func() {
		err = env.Client.Subscribe(ctx, subject, queueGroup, func(data []byte, topic string) {
			assert.Equal(t, subject, topic, "subject mismatch")
			received <- data
		})
		subErr <- err
	}()

	// Wait for the message
	select {
	case msg := <-received:
		assert.Equal(t, data, msg, "received message does not match published data")
	case err = <-subErr:
		t.Fatalf("Subscription failed: %v", err)
	case <-ctx.Done():
		t.Fatal("Did not receive the published message in time")
	}
}

func TestNatsClient_Subscribe_MultipleSubscribers(t *testing.T) {
	env := SetupTestEnvironment(t)

	subject := "test.multiple.subscribers"
	data := []byte("Message for multiple subscribers")

	// Publish the message
	err := env.Client.Publish(context.Background(), subject, data)
	require.NoError(t, err, "expected successful publish")

	// Channels to capture received messages
	received1 := make(chan []byte, 1)
	received2 := make(chan []byte, 1)
	defer close(received1)
	defer close(received2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()

	// Start first subscriber
	go func() {
		err = env.Client.Subscribe(ctx, subject, "", func(data []byte, topic string) {
			assert.Equal(t, subject, topic, "subject mismatch")
			received1 <- data
		})
	}()

	// Start second subscriber
	go func() {
		err = env.Client.Subscribe(ctx, subject, "", func(data []byte, topic string) {
			assert.Equal(t, subject, topic, "subject mismatch")
			received2 <- data
		})
	}()

	// Check both subscribers received the message
	select {
	case msg1 := <-received1:
		t.Logf("Received message 1: %s", string(msg1))
		assert.Equal(t, data, msg1, "subscriber 1 message mismatch")
	case <-ctx.Done():
		t.Fatal("Subscriber 1 did not receive the message in time")
	}

	select {
	case msg2 := <-received2:
		t.Logf("Received message 2: %s", string(msg2))
		assert.Equal(t, data, msg2, "subscriber 2 message mismatch")
	case <-ctx.Done():
		t.Fatal("Subscriber 2 did not receive the message in time")
	}
}
