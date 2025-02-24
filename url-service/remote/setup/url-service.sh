#!/bin/bash

set -euo pipefail

# ======================================================================== #
# VARIABLES
# ======================================================================== #

USERNAME="url-service"

NATS_RPC_HOST=127.0.0.1
NATS_RPC_PORT=61355

TLS_CERTIFICATE=/opt/grpc-certs/fullchain.pem
TLS_KEY=/opt/grpc-certs/privkey.pem

NATS_RPC_SERVER_PORT=61355

MONGO_HOST=127.0.0.1
MONGO_PORT=27017
MONGO_USER=user
MONGO_PASS=pass
MONGO_DB=url
MONGO_COLLECTION=list

INBOUND_MESSAGE_BATCH_SIZE=25
INBOUND_MESSAGE_QUEUE_GROUP=

OUTBOUND_MESSAGE_BATCH_SIZE=25

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
# setup_mongodb
#
# Installs and configures MongoDB, enabling authorization and creating the
# necessary MongoDB admin user for the application.
# ------------------------------------------------------------------------------
setup_mongodb() {
    echo "Updating package list and installing prerequisites..."
    apt-get update -q
    apt-get install -y gnupg curl

    echo "Importing MongoDB GPG key..."
    curl -fsSL https://www.mongodb.org/static/pgp/server-7.0.asc | gpg --dearmor -o /usr/share/keyrings/mongodb-server-7.0.gpg

    echo "Creating MongoDB source list for Ubuntu 22.04 (Jammy)..."
    echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" | tee /etc/apt/sources.list.d/mongodb-org-7.0.list

    echo "Updating package list..."
    apt-get update -q

    echo "Installing MongoDB..."
    apt-get install -y mongodb-org

    echo "Configuring MongoDB..."
    cp /etc/mongod.conf /etc/mongod.conf.bak
    echo -e "\nsecurity:\n  authorization: enabled" >> /etc/mongod.conf

    echo "Starting and enabling MongoDB service..."
    systemctl restart mongod
    systemctl enable mongod

    echo "Waiting for MongoDB service to initialize..."
    sleep 10  # Adjust this delay if needed

    echo "Checking MongoDB service status..."
    if ! systemctl is-active --quiet mongod; then
        echo "MongoDB service failed to start. Check /var/log/mongodb/mongod.log for details. Exiting..."
        exit 1
    fi

    echo "Creating MongoDB user..."
    retry=0
    until mongosh --eval "db.runCommand({ connectionStatus: 1 })" &>/dev/null || [ $retry -ge 5 ]; do
        echo "Waiting for MongoDB to accept connections..."
        sleep 5
        retry=$((retry + 1))
    done

    if [ $retry -ge 5 ]; then
        echo "MongoDB did not become available. Exiting..."
        exit 1
    fi

    mongosh <<EOF
use admin
db.createUser({
  user: "${MONGO_USER}",
  pwd: "${MONGO_PASS}",
  roles: [{ role: "root", db: "admin" }]
})
EOF

    echo "MongoDB setup complete!"
}

# ------------------------------------------------------------------------------
# initialize_mongodb
#
# Creates the application database and initializes collections.
# ------------------------------------------------------------------------------
initialize_mongodb() {
    echo "Creating database '${MONGO_DB}' and initializing collections..."
    mongosh -u "$MONGO_USER" -p "$MONGO_PASS" --authenticationDatabase admin <<EOF
use ${MONGO_DB}
db.createCollection("${MONGO_COLLECTION}")
EOF
    echo "Database and collections initialized successfully!"
}

# ------------------------------------------------------------------------------
# verify_mongodb
#
# Verify MongoDB connection and user authentication.
# ------------------------------------------------------------------------------
verify_mongodb() {
    echo "Verifying MongoDB connection and user authentication..."

    mongosh --host 127.0.0.1 --port 27017 -u "$MONGO_USER" -p "$MONGO_PASS" --authenticationDatabase admin <<EOF
use admin
db.runCommand({ connectionStatus: 1 })
EOF

    if [ $? -eq 0 ]; then
        echo "MongoDB connection verified successfully!"
    else
        echo "MongoDB connection verification failed!"
        exit 1
    fi
}

# ------------------------------------------------------------------------------
# set_url_service_env
#
# Sets essential environment variables in /etc/service/url-service for global access.
# ------------------------------------------------------------------------------
set_url_service_env() {
    local env_file="/etc/service/url-service"
    echo "Writing url-service environment variables to ${env_file}..."
    mkdir -p "$(dirname "${env_file}")"
    cat <<EOF > "${env_file}"
# ====================================================
# url-service Environment Variables
# ====================================================

# NATS RPC configuration
NATS_RPC_HOST=${NATS_RPC_HOST}
NATS_RPC_PORT=${NATS_RPC_PORT}

# TLS configuration for secure connections
TLS_CERTIFICATE=${TLS_CERTIFICATE}
TLS_KEY=${TLS_KEY}

# NATS RPC Server configuration
NATS_RPC_SERVER_PORT=${NATS_RPC_SERVER_PORT}

# MongoDB configuration
MONGO_HOST=${MONGO_HOST}
MONGO_PORT=${MONGO_PORT}
MONGO_USER=${MONGO_USER}
MONGO_PASS=${MONGO_PASS}
MONGO_DB=${MONGO_DB}
MONGO_COLLECTION=${MONGO_COLLECTION}

# Inbound message processing settings
INBOUND_MESSAGE_BATCH_SIZE=${INBOUND_MESSAGE_BATCH_SIZE}
INBOUND_MESSAGE_QUEUE_GROUP=${INBOUND_MESSAGE_QUEUE_GROUP}

# Outbound message processing settings
OUTBOUND_MESSAGE_BATCH_SIZE=${OUTBOUND_MESSAGE_BATCH_SIZE}

# Application environment
ENV=${ENV}
EOF
}

# ======================================================================== #
# MAIN SCRIPT
# ======================================================================== #

main() {
  create_user
  set_url_service_env
  setup_mongodb
  initialize_mongodb
  verify_mongodb

  echo "Script complete! Rebooting..."
  reboot
}

main "$@"