variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "cluster_name" {
  description = "ECS cluster name"
  type        = string
  default     = "eskimo-scanner"
}

variable "github_org" {
  description = "GitHub organization to scan"
  type        = string
}

variable "github_repository" {
  description = "GitHub repository using OIDC"
  type        = string
  default     = "cybrota/eskimo"
}

variable "secret_name" {
  description = "Secrets Manager secret name"
  type        = string
  default     = "eskimo-config"
}

variable "secret_values" {
  description = "Key-value map for scanner configuration"
  type        = map(string)
  default     = {}
}

variable "scan_schedule_expression" {
  description = "Cron expression for scheduled scans"
  type        = string
  default     = "0 0 ? * MON *"
}
