package validators

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

// BusValidator implements validation rules for gRPC requests.
type BusValidator struct{}

// NewBusValidator creates a new instance of BusValidator.
func NewBusValidator() *BusValidator {
	return &BusValidator{}
}

// ValidatePublishRequest validates the PublishRequest fields.
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

// ValidateSubscribeRequest validates the SubscribeRequest fields.
func (v *BusValidator) ValidateSubscribeRequest(request *natsservicev1.SubscribeRequest) (err error) {
	var errors []error

	if strings.TrimSpace(request.GetSubject()) == "" {
		errors = append(errors, status.Error(codes.InvalidArgument, "subject is required"))
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
