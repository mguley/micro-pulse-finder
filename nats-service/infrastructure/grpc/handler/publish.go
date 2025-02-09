package handler

import (
	"context"
	"fmt"
	natsservicev1 "shared/proto/nats-service/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Publish is a unary RPC that publishes a message to a given NATS subject.
func (s *BusService) Publish(
	ctx context.Context,
	request *natsservicev1.PublishRequest,
) (response *natsservicev1.PublishResponse, err error) {
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "context canceled")

	default:
		if result := s.validator.ValidatePublishRequest(request); result != nil {
			return nil, result
		}
		if result := s.operations.Publish(request.GetSubject(), request.GetData()); result != nil {
			return &natsservicev1.PublishResponse{
				Success: false,
				Message: fmt.Sprintf("Could not publish: %v", result),
			}, result
		}

		return &natsservicev1.PublishResponse{
			Success: true,
			Message: "Message published successfully",
		}, nil
	}
}
