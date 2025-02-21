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

## test/url-service/integration: Run integration tests with race detector
.PHONY: test/url-service/integration
test/url-service/integration:
	@echo 'Running integration tests...'
	CGO_ENABLED=1 go test -race -v -count=1 -p=1 ./tests/integration/...
