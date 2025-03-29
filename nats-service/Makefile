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

## test/nats-service/integration: Run integration tests with race detector (bypass cache)
.PHONY: test/nats-service/integration
test/nats-service/integration:
	@echo 'Running integration tests...'
	CGO_ENABLED=1 go test -race -v -count=1 -p=1 ./tests/integration/...

# =============================================================================== #
# DEVELOPMENT
# =============================================================================== #

## run/nats-grpc-server: Run NATS gRPC server.
.PHONY: run/nats-grpc-server
run/nats-grpc-server:
	go run ./cmd/grpc

# =============================================================================== #
# BUILD
# =============================================================================== #

## build/service: Build nats service
.PHONY: build/service
build/service:
	@echo 'Building nats service...'
	@mkdir -p ./bin
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -ldflags="-s -w" -o=./bin/nats-service-o ./cmd/grpc
	@echo 'Build for Linux (amd64) complete.'

# =============================================================================== #
# DEPLOYMENT
# =============================================================================== #

## remote/copy-files: Transfer binary and service files.
.PHONY: remote/copy-files
remote/copy-files:
	@echo 'Deploying new binary and service files...'
	rsync -P ./bin/nats-service-o nats-service@${PRODUCTION_HOST_IP}:/tmp/nats-service-o
	rsync -P ./remote/service/nats.service nats-service@${PRODUCTION_HOST_IP}:/tmp/nats.service
	rsync -P ./remote/service/nats-grpc-server.service nats-service@${PRODUCTION_HOST_IP}:/tmp/nats-grpc-server.service
	ssh -t nats-service@${PRODUCTION_HOST_IP} 'set -e; \
	  sudo mkdir -p /opt/nats-grpc-server && \
	  sudo mv /tmp/nats-service-o /opt/nats-grpc-server/ && \
	  sudo chown -R nats-service:nats-service /opt/nats-grpc-server && \
	  sudo chmod +x /opt/nats-grpc-server/nats-service-o && \
	  sudo mv /tmp/nats.service /etc/systemd/system/ && \
	  sudo mv /tmp/nats-grpc-server.service /etc/systemd/system/'

## remote/restart-service: Restart nats-service and NATS services.
.PHONY: remote/restart-service
remote/restart-service:
	@echo 'Restarting services...'
	ssh -t nats-service@${PRODUCTION_HOST_IP} 'set -e; \
	  sudo systemctl daemon-reload && \
	  sudo systemctl enable nats-grpc-server && \
	  sudo systemctl restart nats-grpc-server && \
	  sudo systemctl enable nats && \
	  sudo systemctl restart nats'

## remote/deploy: Deploy to remote host.
.PHONY: remote/deploy
remote/deploy:
	@$(MAKE) build/service
	@$(MAKE) remote/copy-files
	@$(MAKE) remote/restart-service
	@echo 'Deployment complete.'

# =============================================================================== #
# LOAD TEST RUNNERS
# =============================================================================== #

## run/load-tests: Run load tests with default configuration.
.PHONY: run/load-tests
run/load-tests:
	go run ./cmd/load

## run/load-publish: Run NATS publish load tests.
.PHONY: run/load-publish
run/load-publish:
	LOAD_TEST_TYPE=publish go run ./cmd/load

## run/load-subscribe: Run NATS subscribe load tests.
.PHONY: run/load-subscribe
run/load-subscribe:
	LOAD_TEST_TYPE=subscribe go run ./cmd/load

## run/load-high-concurrency: Run with high concurrency (300 clients).
.PHONY: run/load-high-concurrency
run/load-high-concurrency:
	LOAD_TEST_CONCURRENCY=300 go run ./cmd/load

## run/load-long-duration: Run extended duration test (5 minutes).
.PHONY: run/load-long-duration
run/load-long-duration:
	LOAD_TEST_DURATION=5m go run ./cmd/load
