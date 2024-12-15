# Variables configuration
# Typically saved as variables.tf
variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "craft"
}

variable "environment" {
  description = "Environment (dev/staging/prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "k8s_config_path" {
  description = "Path to Kubernetes config file"
  type        = string
  default     = "~/.kube/config"
}