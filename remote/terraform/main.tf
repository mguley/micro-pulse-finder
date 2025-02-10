# Resource to create a Vultr instance (virtual machine)
resource "vultr_instance" "micro" {
  plan = var.plan                           # Specifies the server plan
  region = var.region                       # The region where the server will be deployed (e.g., Frankfurt)
  os_id = var.os_id                         # Operating System ID (e.g., Ubuntu 24.04 LTS)
  label = var.label                         # A label for identifying the instance
  hostname = var.hostname                   # The hostname for the server
  backups = "disabled"                      # Auto backups are disabled
  enable_ipv6 = false                       # IPv6 is disabled
  ssh_key_ids = [vultr_ssh_key.default.id]  # SSH key IDs for secure access to the instance
}

# Resource to add an SSH key to Vultr for authentication
resource "vultr_ssh_key" "default" {
  name = "default_ssh_key"                  # Name for the SSH key in Vultr
  ssh_key = file("~/.ssh/id_ed25519.pub")   # Path to the public SSH key file
}