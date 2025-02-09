package handler

import (
	"context"
	"fmt"
	natsservicev1 "shared/proto/nats-service/gen"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBusService_Publish(t *testing.T) {
	client := SetupTestContainer(t)

	// Define test cases
	tests := []struct {
		name        string
		request     *natsservicev1.PublishRequest
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name: "Large Data Payload",
			request: &natsservicev1.PublishRequest{
				Subject: "test.subject.large",
				Data:    make([]byte, 1024*1024), // 1 MB payload
			},
			expectedErr: false,
		},
		{
			name: "Invalid Subject with Whitespace",
			request: &natsservicev1.PublishRequest{
				Subject: "   ",
				Data:    []byte("payload"),
			},
			expectedErr: true,
			errCode:     codes.InvalidArgument,
		},
		{
			name: "Valid Publish",
			request: &natsservicev1.PublishRequest{
				Subject: "test.subject",
				Data:    []byte("test payload"),
			},
			expectedErr: false,
		},
		{
			name: "Missing Subject",
			request: &natsservicev1.PublishRequest{
				Subject: "",
				Data:    []byte("test payload"),
			},
			expectedErr: true,
			errCode:     codes.InvalidArgument,
		},
		{
			name: "Empty Data Payload",
			request: &natsservicev1.PublishRequest{
				Subject: "test.subject.empty",
				Data:    []byte(""),
			},
			expectedErr: true,
			errCode:     codes.InvalidArgument,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
			defer cancel()

			response, err := client.Publish(ctx, tc.request)

			if tc.expectedErr {
				require.Error(t, err, "Expected error but got nil")
				st, ok := status.FromError(err)
				require.True(t, ok, "Error is not a gRPC status")
				assert.Equal(t, tc.errCode, st.Code(), "Unexpected gRPC error code")
			} else {
				require.NoError(t, err, "Unexpected error occurred")
				assert.True(t, response.GetSuccess(), "Publish response should indicate success")
				assert.NotEmpty(t, response.GetMessage(), "Publish response message should not be empty")
			}
		})
	}
}

func TestBusService_Subscribe(t *testing.T) {
	client := SetupTestContainer(t)

	subject := "test.subject.subscribe"
	message := []byte("test message")

	// Simulate publisher
	go func() {
		var (
			response *natsservicev1.PublishResponse
			err      error
		)
		time.Sleep(time.Duration(1) * time.Second)
		response, err = client.Publish(context.Background(), &natsservicev1.PublishRequest{
			Subject: subject,
			Data:    message,
		})
		require.NoError(t, err, "Failed to publish message")
		require.True(t, response.GetSuccess(), "Publish response should indicate success")
	}()

	// Define test cases
	tests := []struct {
		name        string
		request     *natsservicev1.SubscribeRequest
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name: "Valid Subscribe",
			request: &natsservicev1.SubscribeRequest{
				Subject: subject,
			},
			expectedErr: false,
		},
		{
			name: "Missing Subject",
			request: &natsservicev1.SubscribeRequest{
				Subject: "",
			},
			expectedErr: true,
			errCode:     codes.InvalidArgument,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
			defer cancel()

			stream, err := client.Subscribe(ctx, tc.request)

			if tc.expectedErr {
				_, streamErr := stream.Recv()
				require.Error(t, streamErr, "Expected error during streaming but got nil")
				st, ok := status.FromError(streamErr)
				require.True(t, ok, "Error is not a gRPC status")
				assert.Equal(t, tc.errCode, st.Code(), "Unexpected gRPC error code")
			} else {
				require.NoError(t, err, "Unexpected error occurred during Subscribe")
				require.NotNil(t, stream, "Stream should not be nil for valid Subscribe requests")

				// Read from the stream
				msg, streamErr := stream.Recv()
				require.NoError(t, streamErr, "Unexpected error while receiving from stream")
				assert.Equal(t, subject, msg.GetSubject(), "Message subject mismatch")
				assert.Equal(t, message, msg.GetData(), "Message data mismatch")
			}
		})
	}
}

func TestBusService_Subscribe_Multiple_Messages(t *testing.T) {
	client := SetupTestContainer(t)

	// Define test cases
	tests := []struct {
		name        string
		setup       func()
		request     *natsservicev1.SubscribeRequest
		expectedErr bool
		errCode     codes.Code
		verify      func(t *testing.T, stream natsservicev1.BusService_SubscribeClient)
	}{
		{
			name: "Multiple Messages",
			setup: func() {
				go func() {
					time.Sleep(time.Duration(1) * time.Second)
					for i := 0; i < 10; i++ {
						result, err := client.Publish(context.Background(), &natsservicev1.PublishRequest{
							Subject: "test.multiple.messages",
							Data:    []byte(fmt.Sprintf("message-%d", i)),
						})
						require.NoError(t, err, "Failed to publish message")
						require.True(t, result.GetSuccess(), "Publish response should indicate success")
					}
				}()
			},
			request: &natsservicev1.SubscribeRequest{
				Subject: "test.multiple.messages",
			},
			expectedErr: false,
			verify: func(t *testing.T, stream natsservicev1.BusService_SubscribeClient) {
				for i := 0; i < 10; i++ {
					msg, err := stream.Recv()
					require.NoError(t, err, "Unexpected error while receiving from stream")
					assert.Equal(t, fmt.Sprintf("message-%d", i), string(msg.GetData()))
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
			defer cancel()

			stream, err := client.Subscribe(ctx, tc.request)
			require.NoError(t, err, "Failed to subscribe")
			tc.verify(t, stream)
		})
	}
}
