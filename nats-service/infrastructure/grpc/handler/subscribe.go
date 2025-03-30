package handler

import (
	"log/slog"
	natsservicev1 "shared/proto/nats-service/gen"
	"sync"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// channelBufferSize defines the buffer size for the messages channel.
	channelBufferSize = 64
)

// responsePool is a sync.Pool used to reuse SubscribeResponse objects
// to minimize allocations during high message throughput.
var responsePool = sync.Pool{
	New: func() interface{} {
		return &natsservicev1.SubscribeResponse{}
	},
}

// reset clears all fields in the SubscribeResponse to prepare it for reuse.
func reset(response *natsservicev1.SubscribeResponse) {
	response.Data = nil
	response.Subject = ""
}

// Subscribe is a server-streaming RPC method that subscribes to a NATS subject
// and streams incoming messages to the client.
//
// It listens for messages on the specified subject and streams them as SubscribeResponse messages.
//
// Parameters:
//   - request: Pointer to the SubscribeRequest containing the subject and an optional queue group.
//   - server:  The gRPC server streaming interface used to send SubscribeResponse messages.
//
// Returns:
//   - err: An error if the subscription or streaming fails, or nil if the operation is successful.
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
		messagesCh = make(chan *nats.Msg, channelBufferSize)
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
			return status.Error(codes.Canceled, ctx.Err().Error())
		case message, ok := <-messagesCh:
			if !ok {
				s.logger.Info("Message channel closed, shutting down subscription")
				return nil
			}

			// Retrieve a response object from the pool and populate it.
			response := responsePool.Get().(*natsservicev1.SubscribeResponse)
			response.Data = message.Data
			response.Subject = message.Subject

			if err = server.Send(response); err != nil {
				reset(response)
				responsePool.Put(response)
				s.logger.Error("Failed to send response",
					slog.String("topic", subject), slog.String("error", err.Error()))
				return status.Error(codes.Internal, err.Error())
			}

			// Reset and return the response object to the pool.
			reset(response)
			responsePool.Put(response)
		}
	}
}
