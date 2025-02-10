#!/bin/bash

set -euo pipefail

# ======================================================================== #
# VARIABLES
# ======================================================================== #

TIMEZONE="Europe/Berlin"            # Timezone to configure on the server

# Export locale to ensure consistent system behavior
export LC_ALL=en_US.UTF-8

# ======================================================================== #
# FUNCTIONS
# ======================================================================== #

# ------------------------------------------------------------------------------
# update_system
#
# Updates and upgrades all system packages to ensure software is up-to-date.
# ------------------------------------------------------------------------------
update_system() {
    echo "Updating system packages..."
    apt-get update -q
    apt-get --yes -o Dpkg::Options::="--force-confnew" upgrade
}

# ------------------------------------------------------------------------------
# enable_repositories
#
# Enables necessary repositories (e.g., universe repository) to support
# installation of required packages.
# ------------------------------------------------------------------------------
enable_repositories() {
    echo "Enabling universe repository..."
    add-apt-repository --yes universe
}

# ------------------------------------------------------------------------------
# setup_time_and_locale
#
# Configures the server timezone and installs locales to ensure compatibility
# with internationalization needs.
# ------------------------------------------------------------------------------
setup_time_and_locale() {
    echo "Setting timezone to ${TIMEZONE}..."
    timedatectl set-timezone "${TIMEZONE}"
    echo "Installing all locales..."
    apt-get --yes install locales-all
}

# ------------------------------------------------------------------------------
# configure_firewall
#
# Configures the firewall (ufw) to allow specific services and enable the firewall.
# ------------------------------------------------------------------------------
configure_firewall() {
    echo "Configuring firewall to allow services..."
    ufw allow 22           # SSH
    ufw --force enable
}

# ------------------------------------------------------------------------------
# upgrade_system
#
# Upgrade system packages.
# ------------------------------------------------------------------------------
upgrade_system() {
    echo "Upgrading all system packages..."
    apt-get --yes -o Dpkg::Options::="--force-confnew" upgrade
}

# ======================================================================== #
# MAIN SCRIPT
# ======================================================================== #

main() {
  enable_repositories
  update_system
  setup_time_and_locale
  configure_firewall
  upgrade_system

  echo "Script complete! Rebooting..."
  reboot
}

main "$@"
