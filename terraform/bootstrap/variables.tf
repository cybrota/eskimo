variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "state_bucket_name" {
  description = "Name of S3 bucket for Terraform state"
  type        = string
}

variable "lock_table_name" {
  description = "DynamoDB table for Terraform state locking"
  type        = string
  default     = "terraform-lock"
}
