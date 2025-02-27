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
	s.logger.Debug(
		"Publish request received",
		"subject", request.GetSubject(), "size", len(request.GetData()))

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "context canceled")

	default:
		if result := s.validator.ValidatePublishRequest(request); result != nil {
			s.logger.Error(
				"Publish request failed due to validation error",
				"subject", request.GetSubject(), "error", result)
			return nil, result
		}
		if result := s.operations.Publish(request.GetSubject(), request.GetData()); result != nil {
			s.logger.Error("Publish failed", "subject", request.GetSubject(), "error", result)
			return &natsservicev1.PublishResponse{
				Success: false,
				Message: fmt.Sprintf("Could not publish: %v", result),
			}, result
		}

		s.logger.Info("Publish succeeded", "subject", request.GetSubject())
		return &natsservicev1.PublishResponse{
			Success: true,
			Message: "Message published successfully",
		}, nil
	}
}
