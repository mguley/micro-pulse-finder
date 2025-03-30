package runner

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"shared/grpc/clients/nats_service"
	"sync"
	"time"
)

// NatsServiceSubscribeRunner implements the core.Runner interface to test subscribe operations against the NATS service
//
// Fields:
//   - client:              The NATS client for subscribing and publishing messages.
//   - subject:             The subject to subscribe to and publish messages.
//   - queueGroup:          The queue group for subscription (used for load balancing).
//   - payload:             Randomly generated message payload for publishing.
//   - messageSize:         The size of the payload in bytes.
//   - subscriberSemaphore: Semaphore to limit concurrent subscriber goroutines.
//   - publishInterval:     Interval between published messages.
//   - subscribeTimeout:    Timeout for subscription operations.
//   - publisherCtx:        Context controlling the lifecycle of the publisher goroutine.
//   - publisherCancel:     Function to cancel the publisher goroutine.
//   - wg:                  WaitGroup for managing goroutines lifecycle.
//   - logger:              Logger instance for structured logging.
type NatsServiceSubscribeRunner struct {
	client              *nats_service.NatsClient
	subject             string
	queueGroup          string
	payload             []byte
	messageSize         int
	subscriberSemaphore chan struct{}
	publishInterval     time.Duration
	subscribeTimeout    time.Duration
	publisherCtx        context.Context
	publisherCancel     context.CancelFunc
	wg                  sync.WaitGroup
	logger              *slog.Logger
}

// NewNatsServiceSubscribeRunner creates a new instance of NatsServiceSubscribeRunner.
//
// Parameters:
//   - client:           The NATS client for subscribing and publishing messages.
//   - subject:          The subject to subscribe to and publish messages.
//   - queueGroup:       Queue group for message distribution among subscribers.
//   - messageSize:      Size of the message payload in bytes.
//   - maxSubscribers:   Maximum number of concurrent subscribers allowed.
//   - publishInterval:  Interval between each published message.
//   - subscribeTimeout: Maximum duration to wait for subscription messages.
//   - logger:           Logger instance for structured logging.
//
// Returns:
//   - *NatsServiceSubscribeRunner: A pointer to the newly created subscribe runner.
func NewNatsServiceSubscribeRunner(
	client *nats_service.NatsClient,
	subject string,
	queueGroup string,
	messageSize int,
	maxSubscribers int,
	publishInterval time.Duration,
	subscribeTimeout time.Duration,
	logger *slog.Logger,
) *NatsServiceSubscribeRunner {
	return &NatsServiceSubscribeRunner{
		client:              client,
		subject:             subject,
		queueGroup:          queueGroup,
		messageSize:         messageSize,
		publishInterval:     publishInterval,
		subscribeTimeout:    subscribeTimeout,
		subscriberSemaphore: make(chan struct{}, maxSubscribers),
		logger:              logger,
	}
}

// Setup initializes the subscribe runner by generating a random payload and starting a background publisher goroutine.
//
// Parameters:
//   - ctx: The context used for controlling the setup lifecycle.
//
// Returns:
//   - err: An error if payload generation fails; otherwise, nil.
func (r *NatsServiceSubscribeRunner) Setup(ctx context.Context) (err error) {
	r.payload = make([]byte, r.messageSize)
	if _, err = rand.Read(r.payload); err != nil {
		return fmt.Errorf("failed to generate payload: %w", err)
	}

	r.publisherCtx, r.publisherCancel = context.WithCancel(ctx)
	r.wg.Add(1)
	go r.runPublisher()

	r.logger.Info("NatsServiceSubscribeRunner setup complete",
		slog.String("subject", r.subject),
		slog.String("queueGroup", r.queueGroup),
		slog.Int("messageSize", r.messageSize),
		slog.Int("maxSubscribers", cap(r.subscriberSemaphore)),
		slog.String("subscribeTimeout", r.subscribeTimeout.String()),
		slog.String("publishInterval", r.publishInterval.String()))

	return nil
}

// runPublisher is a background goroutine responsible for periodically publishing messages to the configured subject.
//
// This method continuously publishes messages at specified intervals until the publisher context is canceled.
func (r *NatsServiceSubscribeRunner) runPublisher() {
	defer r.wg.Done()
	ticker := time.NewTicker(r.publishInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.publisherCtx.Done():
			return
		case <-ticker.C:
			if err := r.client.Publish(r.publisherCtx, r.subject, r.payload); err != nil {
				r.logger.Error("Error publishing message",
					slog.String("subject", r.subject), slog.String("error", err.Error()))
			}
		}
	}
}

// Run executes the subscription operation, listening for messages published on the specified subject.
// It respects the maximum subscriber limit using a semaphore to control concurrency.
//
// Parameters:
//   - ctx: The context controlling the subscription lifecycle.
//
// Returns:
//   - err: An error if the subscription fails or times out; otherwise, nil upon receiving a message.
func (r *NatsServiceSubscribeRunner) Run(ctx context.Context) (err error) {
	select {
	case r.subscriberSemaphore <- struct{}{}:
		defer func() { <-r.subscriberSemaphore }()
	default:
		// Maximum subscribers reached
		return nil
	}

	var (
		wg      sync.WaitGroup
		msgCh   = make(chan struct{}, 1)
		errCh   = make(chan error, 1)
		handler = func(_ []byte, _ string) {
			select {
			case msgCh <- struct{}{}:
			default:
			}
		}
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if subErr := r.client.Subscribe(ctx, r.subject, r.queueGroup, handler); subErr != nil {
			errCh <- subErr
			return
		}
	}()

	// Continuously listen for messages
	for {
		select {
		case <-ctx.Done():
			r.logger.Debug("Context canceled, existing receiver loop")
			wg.Wait()
			return ctx.Err()
		case err = <-errCh:
			r.logger.Error("Subscription error received", slog.String("error", err.Error()))
			wg.Wait()
			return err
		case <-msgCh:
		}
	}
}

// Teardown stops the background publisher goroutine and closes the NATS client connection,
// ensuring resources are properly released.
//
// Parameters:
//   - ctx: The context controlling the teardown lifecycle.
//
// Returns:
//   - err: An error if closing the NATS client fails; otherwise, nil.
func (r *NatsServiceSubscribeRunner) Teardown(ctx context.Context) (err error) {
	if r.publisherCancel != nil {
		r.publisherCancel()
		r.wg.Wait()
	}

	if r.client != nil {
		if err = r.client.Close(); err != nil {
			return fmt.Errorf("failed to close NATS client: %w", err)
		}
	}

	return err
}

// Name returns the descriptive name of the runner.
//
// Returns:
//   - string: The runner's name ("NATS Service Subscribe Runner").
func (r *NatsServiceSubscribeRunner) Name() (name string) {
	return "NATS Service Subscribe Runner"
}
