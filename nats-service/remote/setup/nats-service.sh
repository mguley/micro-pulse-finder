#!/bin/bash

set -euo pipefail

# ======================================================================== #
# VARIABLES
# ======================================================================== #

USERNAME="nats-service"

NATS_HOST=127.0.0.1
NATS_PORT=4222

TLS_CERTIFICATE=/opt/grpc-certs/fullchain.pem
TLS_KEY=/opt/grpc-certs/privkey.pem

NATS_RPC_SERVER_PORT=61355

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
# Sets essential environment variables in /etc/service/nats-service for global access.
# ------------------------------------------------------------------------------
set_environment_variables() {
    local env_file="/etc/service/nats-service"
    echo "Writing nats-service environment variables to ${env_file}..."
    mkdir -p "$(dirname "${env_file}")"
    cat <<EOF > "${env_file}"
# ====================================================
# nats-service Environment Variables
# ====================================================

# NATS server configuration
NATS_HOST=${NATS_HOST}
NATS_PORT=${NATS_PORT}

# TLS configuration for secure connections
TLS_CERTIFICATE=${TLS_CERTIFICATE}
TLS_KEY=${TLS_KEY}

# RPC configuration for the NATS service
NATS_RPC_SERVER_PORT=${NATS_RPC_SERVER_PORT}

# Application environment
ENV=${ENV}
EOF
}

# ------------------------------------------------------------------------------
# setup_nats
#
# Install and configure NATS server.
# ------------------------------------------------------------------------------
setup_nats() {
      echo "Installing NATS server..."
      # Download the latest NATS Server (nats-server) binary
      curl -L https://github.com/nats-io/nats-server/releases/download/v2.10.25/nats-server-v2.10.25-linux-amd64.tar.gz | tar xz
      mv nats-server-v2.10.25-linux-amd64/nats-server /usr/local/bin/
      chmod +x /usr/local/bin/nats-server
}

# ======================================================================== #
# MAIN SCRIPT
# ======================================================================== #

main() {
  create_user
  set_environment_variables
  setup_nats

  echo "Script complete! Rebooting..."
  reboot
}

main "$@"
