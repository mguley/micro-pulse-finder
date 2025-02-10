# API key for Vultr account authentication
variable "vultr_api_key" {
  description = "API key for Vultr account"
  type        = string
  sensitive   = true
}

# Region to deploy the Vultr instance
variable "region" {
  description = "Region to deploy the server"
  type        = string
  default     = "waw"                     # Poland
}

# Server plan specifying CPU, memory, and storage
variable "plan" {
  description = "Server plan"
  type        = string
  default     = "vc2-1c-1gb"              # Default plan: 1 vCPU, 1 GB RAM, 25 GB SSD
}

# Operating System ID for the instance
variable "os_id" {
  description = "Operating System ID for Ubuntu 24.04"
  type        = number
  default     = 2284                      # Default OS ID for Ubuntu 24.04 LTS x64
}

# Label for identifying the Vultr instance
variable "label" {
  description = "Label prefix"
  type        = string
  default     = "micro"                   # Default label: micro
}

# Hostname for the Vultr instance
variable "hostname" {
  description = "Hostname for Vultr"
  type        = string
  default     = "micro"                  # Default hostname: micro
}
