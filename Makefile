# Include variables
include .envrc

## help: Print this help message
.PHONY: help
help:
	@echo 'Usage':
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# =============================================================================== #
# QUALITY
# =============================================================================== #

## install/linter: Install GolangCI-Lint
.PHONY: install/linter
install/linter:
	@echo "Installing GolangCI-Lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

## lint: Run linter on all Go files in each module directory
.PHONY: lint
lint: install/linter
	@echo "Running GolangCI-Lint on all Go files in each module directory..."
	@find ./nats-service ./proxy-service ./url-service ./shared -name '*.go' -exec dirname {} \; | sort -u | xargs $(shell go env GOPATH)/bin/golangci-lint run

## tidy: format all .go files and tidy module dependencies
SERVICES = nats-service proxy-service url-service shared

.PHONY: tidy
tidy:
	@for service in $(SERVICES); do \
		echo "Tidying and verifying $$service..."; \
		(cd $$service && go mod tidy && go mod verify); \
	done
	@echo "Vendoring workspace dependencies..."
	go work vendor

## vet: Run go vet on all Go packages
.PHONY: vet
vet:
	@echo "Running go vet on all microservices..."
	@for service in nats-service proxy-service shared url-service; do \
		echo "Running go vet in $$service..."; \
		(cd $$service && go vet ./...); \
	done

# =============================================================================== #
# RPC
# =============================================================================== #

## generate/rpc: Generate Go code.
.PHONY: generate/rpc
generate/rpc:
	protoc --go_out=. --go-grpc_out=. shared/proto/nats-service/*.proto

# =============================================================================== #
# TESTING
# =============================================================================== #

## test/shared/grpc/integration: Run grpc integration tests with race detector
.PHONY: test/shared/grpc/integration
test/shared/grpc/integration:
	@echo 'Running RPC integration tests...'
	CGO_ENABLED=1 go test -race -v -count=1 -p=1 ./shared/grpc/tests/integration/...

## test/shared/mongodb/integration: Run mongodb integration tests with race detector
.PHONY: test/shared/mongodb/integration
test/shared/mongodb/integration:
	@echo 'Running MongoDB integration tests...'
	CGO_ENABLED=1 go test -race -v -count=1 -p=1 ./shared/mongodb/tests/integration/...

## test/shared/integration: Run shared integration tests
.PHONY: test/shared/integration
test/shared/integration:
	@$(MAKE) test/shared/grpc/integration
	@$(MAKE) test/shared/mongodb/integration
