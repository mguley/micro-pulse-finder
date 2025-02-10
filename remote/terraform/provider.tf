# Specify the required provider and its version
terraform {
  required_providers {
    vultr = {
      source = "vultr/vultr"          # Provider source for Vultr
      version = "2.23.1"              # Specific version of the Vultr provider
    }
  }
}

# Configuration for the Vultr provider
provider "vultr" {
  api_key = var.vultr_api_key         # API key for authentication with Vultr
}