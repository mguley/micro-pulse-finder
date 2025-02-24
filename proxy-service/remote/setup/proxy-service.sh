#!/bin/bash

set -euo pipefail

# ======================================================================== #
# VARIABLES
# ======================================================================== #

USERNAME="proxy-service"

NATS_HOST=127.0.0.1
NATS_PORT=4222
NATS_RPC_HOST=127.0.0.1
NATS_RPC_PORT=61355

TLS_CERTIFICATE=/opt/grpc-certs/fullchain.pem
TLS_KEY=/opt/grpc-certs/privkey.pem

PROXY_RPC_SERVER_PORT=61155

PROXY_HOST=127.0.0.1
PROXY_PORT=9050
PROXY_CONTROL_PASSWORD=password
PROXY_CONTROL_PORT=9051
PROXY_URL=https://httpbin.org/ip

POOL_MAX_SIZE=5
POOL_REFRESH_INTERVAL=15

URL_PROCESSOR_BATCH_SIZE=5
URL_PROCESSOR_QUEUE_GROUP=

ENV=dev

# ======================================================================== #
# FUNCTIONS
# ======================================================================== #

# ------------------------------------------------------------------------------
# create_user
#
# Creates a new system user with sudo privileges. If root's SSH keys are
# available, they will be copied to the new user's home directory.
# ------------------------------------------------------------------------------
create_user() {
    if id "${USERNAME}" &>/dev/null; then
        echo "User ${USERNAME} already exists. Skipping creation."
    else
        echo "Creating user ${USERNAME}..."
        useradd --create-home --shell "/bin/bash" --groups sudo "${USERNAME}"
        passwd --delete "${USERNAME}"               # Remove password (force password reset on first login)
        chage --lastday 0 "${USERNAME}"             # Expire password immediately

        echo "Copying SSH keys to new user..."
        if [ -d "/root/.ssh" ]; then
            rsync --archive --chown="${USERNAME}:${USERNAME}" /root/.ssh /home/"${USERNAME}"
        else
            echo "No SSH keys found in /root/.ssh. Skipping SSH key copy."
        fi
    fi
}

# ------------------------------------------------------------------------------
# set_environment_variables
#
# Sets essential environment variables in /etc/service/proxy-service for global access.
# ------------------------------------------------------------------------------
set_environment_variables() {
    local env_file="/etc/service/proxy-service"
    echo "Writing proxy-service environment variables to ${env_file}..."
    mkdir -p "$(dirname "${env_file}")"

    cat <<EOF > "${env_file}"
# ====================================================
# proxy-service Environment Variables
# ====================================================

# NATS server configuration
NATS_HOST=${NATS_HOST}
NATS_PORT=${NATS_PORT}

# NATS RPC configuration
NATS_RPC_HOST=${NATS_RPC_HOST}
NATS_RPC_PORT=${NATS_RPC_PORT}

# TLS configuration for secure connections
TLS_CERTIFICATE=${TLS_CERTIFICATE}
TLS_KEY=${TLS_KEY}

# RPC configuration for the proxy service
PROXY_RPC_SERVER_PORT=${PROXY_RPC_SERVER_PORT}

# Proxy server configuration
PROXY_HOST=${PROXY_HOST}
PROXY_PORT=${PROXY_PORT}
PROXY_CONTROL_PASSWORD=${PROXY_CONTROL_PASSWORD}
PROXY_CONTROL_PORT=${PROXY_CONTROL_PORT}
PROXY_URL=${PROXY_URL}

# Connection Pool configuration
POOL_MAX_SIZE=${POOL_MAX_SIZE}
POOL_REFRESH_INTERVAL=${POOL_REFRESH_INTERVAL}

# URL Processor configuration
URL_PROCESSOR_BATCH_SIZE=${URL_PROCESSOR_BATCH_SIZE}
URL_PROCESSOR_QUEUE_GROUP=${URL_PROCESSOR_QUEUE_GROUP}

# Application environment
ENV=${ENV}
EOF
}

# ------------------------------------------------------------------------------
# setup_tor
#
# Installs and configures the Tor proxy for the application.
# ------------------------------------------------------------------------------
setup_tor() {
    echo "Installing Tor and necessary dependencies..."
    apt-get update -q
    apt-get install -y tor curl netcat-openbsd

    echo "Configuring Tor..."
    cat <<EOF >/etc/tor/torrc
SocksPort 0.0.0.0:9050 IsolateSOCKSAuth
ControlPort 0.0.0.0:9051
HashedControlPassword 16:EC1800A189DA53D6600B08E22D26B20C2A34E24962AA23FC6E5AA8B8F4
EOF

    echo "Starting and enabling Tor service..."
    systemctl restart tor
    systemctl enable tor

    echo "Waiting for Tor to initialize..."
    sleep 15
}

# ------------------------------------------------------------------------------
# verify_tor
#
# Validates Tor proxy installation.
# ------------------------------------------------------------------------------
verify_tor() {
    echo "Verifying Tor proxy installation..."

    echo "Checking exit IP using Tor..."
    exit_ip=$(curl --socks5-hostname localhost:9050 https://httpbin.org/ip --max-time 10 2>/dev/null | jq -r '.origin')
    if [ -z "$exit_ip" ]; then
        echo "Tor proxy validation failed! Unable to fetch exit IP."
        exit 1
    fi

    echo "Exit IP via Tor: $exit_ip"

    echo "Checking Tor control port..."
    echo -e 'authenticate "password"\ngetinfo status/circuit-established' | nc -w 5 localhost 9051 > /tmp/tor_control_status

    if grep -q "250-status/circuit-established=1" /tmp/tor_control_status; then
        echo "Tor proxy is fully operational."
    else
        echo "Tor proxy validation failed! Circuit not established."
        echo "Tor control port response:"
        cat /tmp/tor_control_status
        exit 1
    fi
}

# ======================================================================== #
# MAIN SCRIPT
# ======================================================================== #

main() {
  create_user
  set_environment_variables
  setup_tor
  verify_tor

  echo "Script complete! Rebooting..."
  reboot
}

main "$@"
