package nats_service

import (
	"fmt"
	natsservicev1 "shared/proto/nats-service/gen"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validator defines the interface for validating gRPC requests.
type Validator interface {
	ValidatePublishRequest(request *natsservicev1.PublishRequest) (err error)
	ValidateSubscribeRequest(request *natsservicev1.SubscribeRequest) (err error)
}

// BusClientValidator implements validation rules for gRPC client requests.
type BusClientValidator struct{}

// NewBusClientValidator creates a new instance of BusClientValidator.
func NewBusClientValidator() *BusClientValidator { return &BusClientValidator{} }

// ValidatePublishRequest ensures that the PublishRequest has valid fields.
func (v *BusClientValidator) ValidatePublishRequest(request *natsservicev1.PublishRequest) (err error) {
	var errors []error

	if strings.TrimSpace(request.GetSubject()) == "" {
		errors = append(errors, status.Error(codes.InvalidArgument, "subject required"))
	}
	if len(request.GetData()) == 0 {
		errors = append(errors, status.Error(codes.InvalidArgument, "data required"))
	}

	return combineErrors(errors)
}

// ValidateSubscribeRequest ensures that the SubscribeRequest has valid fields.
func (v *BusClientValidator) ValidateSubscribeRequest(request *natsservicev1.SubscribeRequest) (err error) {
	var errors []error

	if strings.TrimSpace(request.GetSubject()) == "" {
		errors = append(errors, status.Error(codes.InvalidArgument, "subject required"))
	}

	return combineErrors(errors)
}

// combineErrors merges multiple validation errors into a single gRPC error.
func combineErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	var sb strings.Builder
	for i, err := range errs {
		sb.WriteString(fmt.Sprintf("Error %d: %v\n", i+1, err.Error()))
	}
	return status.Error(codes.InvalidArgument, sb.String())
}
