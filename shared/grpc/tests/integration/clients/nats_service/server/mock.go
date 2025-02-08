package server

import (
	"context"
	"fmt"
	natsservicev1 "shared/proto/nats-service/gen"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockBusService is a mock implementation of BusServiceServer for testing.
type MockBusService struct {
	natsservicev1.UnimplementedBusServiceServer
	messages sync.Map // Concurrent map for storing messages
}

// NewMockBusService creates a new instance of MockBusService.
func NewMockBusService() *MockBusService { return &MockBusService{} }

// Publish simulates message publishing.
func (m *MockBusService) Publish(
	ctx context.Context,
	request *natsservicev1.PublishRequest,
) (response *natsservicev1.PublishResponse, err error) {
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "context canceled")
	default:
		if vErr := m.validatePublishRequest(request); vErr != nil {
			return nil, vErr
		}

		// Store the message
		m.messages.Store(request.GetSubject(), request.GetData())

		return &natsservicev1.PublishResponse{
			Success: true,
			Message: "Message published successfully",
		}, nil
	}
}

// Subscribe simulates subscribing to a subject and streaming messages.
func (m *MockBusService) Subscribe(
	request *natsservicev1.SubscribeRequest,
	stream natsservicev1.BusService_SubscribeServer,
) (err error) {
	if vErr := m.validateSubscribeRequest(request); vErr != nil {
		return vErr
	}

	// Retrieve message
	data, exists := m.getMessage(request.GetSubject())
	if !exists {
		return status.Error(codes.NotFound, "message not found")
	}

	// Simulate streaming a message with a delay to mimic real NATS behavior
	time.Sleep(time.Duration(500) * time.Millisecond)

	response := &natsservicev1.SubscribeResponse{
		Data:    data,
		Subject: request.GetSubject(),
	}

	if err = stream.Send(response); err != nil {
		return fmt.Errorf("could not send message to stream: %w", err)
	}

	return nil
}

// getMessage retrieves a stored message from sync.Map.
func (m *MockBusService) getMessage(subject string) (message []byte, exists bool) {
	value, ok := m.messages.Load(subject)
	if !ok {
		return nil, false
	}
	return value.([]byte), true
}

// validatePublishRequest ensures that the PublishRequest has valid fields.
func (m *MockBusService) validatePublishRequest(request *natsservicev1.PublishRequest) (err error) {
	if strings.TrimSpace(request.GetSubject()) == "" {
		return status.Error(codes.InvalidArgument, "subject required")
	}
	if len(request.GetData()) == 0 {
		return status.Error(codes.InvalidArgument, "data required")
	}
	return nil
}

// validateSubscribeRequest ensures that the SubscribeRequest has valid fields.
func (m *MockBusService) validateSubscribeRequest(request *natsservicev1.SubscribeRequest) (err error) {
	if strings.TrimSpace(request.GetSubject()) == "" {
		return status.Error(codes.InvalidArgument, "subject required")
	}
	return nil
}
