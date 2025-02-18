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

# =============================================================================== #
# DEPLOYMENT
# =============================================================================== #

## remote/copy-files: Transfer binary and service files.
.PHONY: remote/copy-files
remote/copy-files:
	@echo 'Deploying new binary and service files...'
	rsync -P ./bin/url-processor-service-o proxy-service@${PRODUCTION_HOST_IP}:/tmp/url-processor-service-o
	rsync -P ./remote/service/url-processor.service proxy-service@${PRODUCTION_HOST_IP}:/tmp/url-processor.service
	ssh -t proxy-service@${PRODUCTION_HOST_IP} 'set -e; \
	  sudo mkdir -p /opt/url-processor-service && \
	  sudo mv /tmp/url-processor-service-o /opt/url-processor-service/ && \
	  sudo chown -R proxy-service:proxy-service /opt/url-processor-service && \
	  sudo chmod +x /opt/url-processor-service/url-processor-service-o && \
	  sudo mv /tmp/url-processor.service /etc/systemd/system/'

## remote/restart-service: Restart url processor service.
.PHONY: remote/restart-service
remote/restart-service:
	@echo 'Restarting services...'
	ssh -t proxy-service@${PRODUCTION_HOST_IP} 'set -e; \
	  sudo systemctl daemon-reload && \
	  sudo systemctl enable url-processor && \
	  sudo systemctl restart url-processor'

## remote/deploy: Deploy to remote host.
.PHONY: remote/deploy
remote/deploy:
	@$(MAKE) build/url-processor-service
	@$(MAKE) remote/copy-files
	@$(MAKE) remote/restart-service
	@echo 'Deployment complete.'
