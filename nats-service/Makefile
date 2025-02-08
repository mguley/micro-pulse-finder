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

## test/nats-service/integration: Run integration tests (bypass cache)
.PHONY: test/nats-service/integration
test/nats-service/integration:
	@echo 'Running integration tests...'
	go test -v -count=1 -p=1 ./tests/integration/...
