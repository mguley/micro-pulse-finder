package handler

import (
	"log/slog"
	natsservicev1 "shared/proto/nats-service/gen"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Subscribe is a server-streaming RPC method that subscribes to a NATS subject and streams incoming messages.
//
// It listens for messages on the specified subject and streams them to the client.
//
// Parameters:
//   - request: Pointer to the SubscribeRequest containing the subject and optional queue group.
//   - server:  The gRPC server streaming interface for sending SubscribeResponse messages.
//
// Returns:
//   - err: An error if the subscription or streaming fails, or nil if successful.
func (s *BusService) Subscribe(
	request *natsservicev1.SubscribeRequest,
	server grpc.ServerStreamingServer[natsservicev1.SubscribeResponse],
) (err error) {
	if result := s.validator.ValidateSubscribeRequest(request); result != nil {
		s.logger.Error("Subscribe request failed due to validation",
			slog.String("subject", request.Subject), slog.String("error", result.Error()))
		return result
	}

	var (
		sub        *nats.Subscription
		messagesCh = make(chan *nats.Msg, 64)
		ctx        = server.Context()
		subject    = request.GetSubject()
		queueGroup = request.GetQueueGroup()
		handler    = func(msg *nats.Msg) { messagesCh <- msg }
	)

	if sub, err = s.operations.Subscribe(ctx, subject, queueGroup, handler); err != nil {
		s.logger.Error("Failed to subscribe",
			slog.String("topic", subject), slog.String("error", err.Error()))
		return status.Error(codes.Internal, err.Error())
	}

	defer func() {
		if unsubErr := sub.Unsubscribe(); unsubErr != nil {
			s.logger.Error("Failed to unsubscribe", slog.String("error", unsubErr.Error()))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context canceled, shutting down subscription")
			return status.Error(codes.Canceled, ctx.Err().Error())
		case message, ok := <-messagesCh:
			if !ok {
				s.logger.Info("Message channel closed, shutting down subscription")
				return nil
			}

			response := &natsservicev1.SubscribeResponse{
				Data:    message.Data,
				Subject: message.Subject,
			}
			if err = server.Send(response); err != nil {
				s.logger.Error("Failed to send response",
					slog.String("topic", subject), slog.String("error", err.Error()))
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}
