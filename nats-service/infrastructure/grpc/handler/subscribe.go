package handler

import (
	"fmt"
	natsservicev1 "shared/proto/nats-service/gen"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
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
		s.logger.Error(
			"Subscribe request failed due to validation",
			"subject", request.GetSubject(), "error", result)
		return result
	}

	var (
		sub        *nats.Subscription
		messagesCh = make(chan *nats.Msg, 64)
		ctx        = server.Context() // Use server context to detect when client cancels or server is shutting down
		subject    = request.GetSubject()
		queueGroup = request.GetQueueGroup()
		handler    = func(msg *nats.Msg) { messagesCh <- msg }
	)

	if sub, err = s.operations.Subscribe(subject, queueGroup, handler); err != nil {
		s.logger.Error("Subscribe operation failed", "subject", subject, "error", err)
		return fmt.Errorf("could not subscribe to subject: %s: %w", request.Subject, err)
	}

	// Ensure we unsubscribe and close the channel when finished.
	defer func() {
		err = sub.Unsubscribe()
		close(messagesCh)
	}()

	// Listen for messages or context cancellation.
	for {
		select {
		case <-ctx.Done():
			// The client closed the stream or the server is shutting down.
			s.logger.Info("Context canceled, shutting down subscription.")
			return ctx.Err()

		case msg := <-messagesCh:
			if msg == nil {
				// Channel closed, end the stream.
				s.logger.Info("Message channel closed, end the stream.")
				return nil
			}

			// Construct a SubscribeResponse and send it to the client.
			response := &natsservicev1.SubscribeResponse{
				Data:    msg.Data,
				Subject: msg.Subject,
			}
			if err = server.Send(response); err != nil {
				s.logger.Error("Send response failed", "subject", response.Subject, "error", err)
				return fmt.Errorf("could not send message to stream: %w", err)
			}
		}
	}
}
