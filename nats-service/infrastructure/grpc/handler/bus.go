package handler

import (
	"log/slog"
	"nats-service/application/services"
	"nats-service/infrastructure/grpc/validators"
	natsservicev1 "shared/proto/nats-service/gen"
)

// BusService is the gRPC service implementation for handling NATS operations.
//
// It provides methods for publishing and subscribing to NATS messages.
//
// Fields:
//   - operations: Reference to service operations for interacting with NATS.
//   - validator:  Validator for incoming gRPC requests.
//   - logger:     Logger for structured logging of service events.
type BusService struct {
	natsservicev1.UnimplementedBusServiceServer
	operations *services.Operations
	validator  validators.Validator
	logger     *slog.Logger
}

// NewBusService creates a new instance of BusService.
//
// Parameters:
//   - operations: Pointer to the Operations service for NATS interactions.
//   - validator:  Validator for validating incoming requests.
//   - logger:     Logger instance for logging.
//
// Returns:
//   - *BusService: A pointer to the newly created BusService.
func NewBusService(operations *services.Operations, validator validators.Validator, logger *slog.Logger) *BusService {
	return &BusService{operations: operations, validator: validator, logger: logger}
}
