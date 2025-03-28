package handler

import (
	"context"
	"fmt"
	"log/slog"
	natsservicev1 "shared/proto/nats-service/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Publish is a unary RPC method that publishes a message to a specified NATS subject.
//
// Parameters:
//   - ctx:     The context for the RPC request.
//   - request: Pointer to the PublishRequest containing the subject and data.
//
// Returns:
//   - response: Response containing the status of the publish operation.
//   - err:      An error if the operation fails, or nil if successful.
func (s *BusService) Publish(
	ctx context.Context,
	request *natsservicev1.PublishRequest,
) (response *natsservicev1.PublishResponse, err error) {
	if result := s.validator.ValidatePublishRequest(request); result != nil {
		s.logger.Error("Publish request failed due to validation",
			slog.String("subject", request.Subject), slog.String("error", result.Error()))
		return nil, result
	}

	if err = s.operations.Publish(ctx, request.GetSubject(), request.GetData()); err != nil {
		s.logger.Error("Failed to publish",
			slog.String("subject", request.GetSubject()),
			slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not publish: %v", err))
	}

	return &natsservicev1.PublishResponse{
		Success: true,
		Message: "Message published successfully",
	}, nil
}
