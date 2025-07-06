variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "state_bucket_name" {
  description = "Name of S3 bucket for Terraform state"
  type        = string
  default     = "eskimo-tf-state"
}

variable "infra_lock_table_name" {
  description = "DynamoDB table for Terraform state locking"
  type        = string
  default     = "eskimo-infra-tf-lock"
}

variable "bootstrap_lock_table_name" {
  description = "DynamoDB table for Terraform state locking"
  type        = string
  default     = "eskimo-bootstrap-tf-lock"
}

variable "state_log_bucket_name" {
  description = "Name of S3 bucket to store state bucket access logs"
  type        = string
  default     = "eskimo-tf-logs"
}
