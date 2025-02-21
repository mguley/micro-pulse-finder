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
.PHONY: tidy
tidy:
	@echo 'Tidying root module dependencies...'
	(cd ./ && go mod tidy)
	@echo 'Verifying root module dependencies...'
	(cd ./ && go mod verify)

	@echo 'Tidying nats-service module dependencies...'
	(cd ./nats-service && go mod tidy)
	@echo 'Verifying nats-service module dependencies...'
	(cd ./nats-service && go mod verify)

	@echo 'Tidying proxy-service module dependencies...'
	(cd ./proxy-service && go mod tidy)
	@echo 'Verifying proxy-service module dependencies...'
	(cd ./proxy-service && go mod verify)

	@echo 'Tidying url-service module dependencies...'
	(cd ./url-service && go mod tidy)
	@echo 'Verifying url-service module dependencies...'
	(cd ./url-service && go mod verify)

	@echo 'Tidying shared module dependencies...'
	(cd ./shared && go mod tidy)
	@echo 'Verifying shared module dependencies...'
	(cd ./shared && go mod verify)

	@echo 'Vendoring workspace dependencies...'
	go work vendor

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
