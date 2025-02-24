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

# =============================================================================== #
# DEVELOPMENT
# =============================================================================== #

## run/inbound-message-service: Run inbound-message-service.
.PHONY: run/inbound-message-service
run/inbound-message-service:
	go run ./cmd/inbound

## run/outbound-message-service: Run outbound-message-service.
.PHONY: run/outbound-message-service
run/outbound-message-service:
	go run ./cmd/outbound

# =============================================================================== #
# BUILD
# =============================================================================== #

## build/inbound-message-service: Build inbound message service.
.PHONY: build/inbound-message-service
build/inbound-message-service:
	@echo 'Building inbound message service...'
	@mkdir -p ./bin
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -ldflags="-s -w" -o=./bin/inbound-message-service-o ./cmd/inbound
	@echo 'Build for Linux (amd64) complete.'

## build/outbound-message-service: Build outbound message service.
.PHONY: build/outbound-message-service
build/outbound-message-service:
	@echo 'Building outbound message service...'
	@mkdir -p ./bin
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -ldflags="-s -w" -o=./bin/outbound-message-service-o ./cmd/outbound
	@echo 'Build for Linux (amd64) complete.'

# =============================================================================== #
# DEPLOYMENT
# =============================================================================== #

## remote/copy-files/inbound-service: Transfer binary and service files.
.PHONY: remote/copy-files/inbound-service
remote/copy-files/inbound-service:
	@echo 'Deploying new binary and service files...'
	rsync -P ./bin/inbound-message-service-o url-service@${PRODUCTION_HOST_IP}:/tmp/inbound-message-service-o
	rsync -P ./remote/service/inbound-message.service url-service@${PRODUCTION_HOST_IP}:/tmp/inbound-message.service
	ssh -t url-service@${PRODUCTION_HOST_IP} 'set -e; \
	  sudo mkdir -p /opt/inbound-message-service && \
	  sudo mv /tmp/inbound-message-service-o /opt/inbound-message-service/ && \
	  sudo chown -R url-service:url-service /opt/inbound-message-service && \
	  sudo chmod +x /opt/inbound-message-service/inbound-message-service-o && \
	  sudo mv /tmp/inbound-message.service /etc/systemd/system/'

## remote/copy-files/outbound-service: Transfer binary and service files.
.PHONY: remote/copy-files/outbound-service
remote/copy-files/outbound-service:
	@echo 'Deploying new binary and service files...'
	rsync -P ./bin/outbound-message-service-o url-service@${PRODUCTION_HOST_IP}:/tmp/outbound-message-service-o
	rsync -P ./remote/service/outbound-message.service url-service@${PRODUCTION_HOST_IP}:/tmp/outbound-message.service
	ssh -t url-service@${PRODUCTION_HOST_IP} 'set -e; \
	  sudo mkdir -p /opt/outbound-message-service && \
	  sudo mv /tmp/outbound-message-service-o /opt/outbound-message-service/ && \
	  sudo chown -R url-service:url-service /opt/outbound-message-service && \
	  sudo chmod +x /opt/outbound-message-service/outbound-message-service-o && \
	  sudo mv /tmp/outbound-message.service /etc/systemd/system/'

## remote/restart-services: Restart services.
.PHONY: remote/restart-services
remote/restart-services:
	@echo 'Restarting services...'
	ssh -t url-service@${PRODUCTION_HOST_IP} 'set -e; \
	  sudo systemctl daemon-reload && \
	  sudo systemctl enable inbound-message && \
	  sudo systemctl enable outbound-message && \
	  sudo systemctl restart inbound-message && \
	  sudo systemctl restart outbound-message'

## remote/deploy: Deploy to remote host.
.PHONY: remote/deploy
remote/deploy:
	@$(MAKE) build/inbound-message-service
	@$(MAKE) build/outbound-message-service
	@$(MAKE) remote/copy-files/inbound-service
	@$(MAKE) remote/copy-files/outbound-service
	@$(MAKE) remote/restart-services
	@echo 'Deployment complete.'