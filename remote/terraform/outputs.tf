# Output the public IP address of the created instance
output "instance_ip" {
  description = "Public IP of the created Vultr instance"
  value       = vultr_instance.micro.main_ip
}

# Output the label of the created instance
output "instance_label" {
  description = "Label of the created Vultr instance"
  value       = vultr_instance.micro.label
}
