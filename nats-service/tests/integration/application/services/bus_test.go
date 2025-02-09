package services

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperations_Subscribe_Publish(t *testing.T) {
	container := SetupTestContainer()
	ops := container.Operations.Get()

	subject := "test.publish.subject"
	data := []byte("test message")

	// Subscribe
	received := make(chan *nats.Msg, 1)
	defer close(received)

	sub, err := ops.Subscribe(subject, "", func(msg *nats.Msg) {
		received <- msg
	})
	require.NoError(t, err, "Failed to subscribe to subject")
	require.NotNil(t, sub, "Failed to create a subscriber")

	// Publish
	err = ops.Publish(subject, data)
	require.NoError(t, err, "Failed to publish message")

	// Verify
	select {
	case msg := <-received:
		assert.Equal(t, data, msg.Data, "Received message does not match published data")
	case <-time.After(time.Duration(2) * time.Second):
		t.Fatal("Did not receive message in time")
	}
}
