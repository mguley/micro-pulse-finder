# Include variables
include .envrc

## help: Print this help message
.PHONY: help
help:
	@echo 'Usage':
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# =============================================================================== #
# TESTING
# =============================================================================== #

## test/proxy-service/integration: Run integration tests (bypass cache)
.PHONY: test/proxy-service/integration
test/proxy-service/integration:
	@echo 'Running integration tests...'
	go test -v -count=1 -p=1 ./tests/integration/...

# =============================================================================== #
# DEVELOPMENT
# =============================================================================== #

## run/url-processor-service: Run url processor service.
.PHONY: run/url-processor-service
run/url-processor-service:
	go run ./cmd/messaging

# =============================================================================== #
# BUILD
# =============================================================================== #

## build/url-processor-service: Build url processor service.
.PHONY: build/url-processor-service
build/url-processor-service:
	@echo 'Building url processor service...'
	@mkdir -p ./bin
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -ldflags="-s -w" -o=./bin/url-processor-service-o ./cmd/messaging
	@echo 'Build for Linux (amd64) complete.'
