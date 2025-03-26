package validators

import (
	"fmt"
	natsservicev1 "shared/proto/nats-service/gen"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validator defines the interface for validating gRPC requests.
//
// Methods:
//   - ValidatePublishRequest:   Validates a PublishRequest.
//   - ValidateSubscribeRequest: Validates a SubscribeRequest.
type Validator interface {
	ValidatePublishRequest(request *natsservicev1.PublishRequest) (err error)
	ValidateSubscribeRequest(request *natsservicev1.SubscribeRequest) (err error)
}

// BusValidator implements validation rules for gRPC requests.
type BusValidator struct{}

// NewBusValidator creates a new instance of BusValidator.
//
// Returns:
//   - *BusValidator: A pointer to the newly created BusValidator.
func NewBusValidator() *BusValidator {
	return &BusValidator{}
}

// ValidatePublishRequest validates the fields of a PublishRequest.
//
// Parameters:
//   - request: Pointer to the PublishRequest to validate.
//
// Returns:
//   - error: A gRPC error if validation fails, or nil if the request is valid.
func (v *BusValidator) ValidatePublishRequest(request *natsservicev1.PublishRequest) (err error) {
	var errors []error

	if strings.TrimSpace(request.GetSubject()) == "" {
		errors = append(errors, status.Error(codes.InvalidArgument, "subject is required"))
	}
	if len(request.GetData()) == 0 {
		errors = append(errors, status.Error(codes.InvalidArgument, "data is required"))
	}

	return combineErrors(errors)
}

// ValidateSubscribeRequest validates the fields of a SubscribeRequest.
//
// Parameters:
//   - request: Pointer to the SubscribeRequest to validate.
//
// Returns:
//   - error: A gRPC error if validation fails, or nil if the request is valid.
func (v *BusValidator) ValidateSubscribeRequest(request *natsservicev1.SubscribeRequest) (err error) {
	var errors []error

	if strings.TrimSpace(request.GetSubject()) == "" {
		errors = append(errors, status.Error(codes.InvalidArgument, "subject is required"))
	}

	return combineErrors(errors)
}

// combineErrors merges multiple validation errors into a single gRPC error.
//
// Parameters:
//   - errs: A slice of error instances to combine.
//
// Returns:
//   - error: A single gRPC error representing all validation errors, or nil if no errors.
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
