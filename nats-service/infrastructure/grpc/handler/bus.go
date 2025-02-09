package handler

import (
	"nats-service/application/services"
	"nats-service/infrastructure/grpc/validators"
	natsservicev1 "shared/proto/nats-service/gen"
)

// BusService is the gRPC service for handling NATS operations.
// It provides methods for publishing messages (unary RPC) and subscribing to messages (server-streaming RPC).
type BusService struct {
	natsservicev1.UnimplementedBusServiceServer
	operations *services.Operations
	validator  validators.Validator
}

// NewBusService creates a new instance of BusService.
func NewBusService(operations *services.Operations, validator validators.Validator) *BusService {
	return &BusService{operations: operations, validator: validator}
}
