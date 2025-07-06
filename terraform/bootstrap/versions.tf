terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  # Local TF state is migrated to the current backend using `terraform init -migrate-state=false`
  backend "s3" {
    bucket         = "eskimo-tf-state"
    key            = "bootstrap/terraform.tfstate"
    region         = "us-west-2"
    dynamodb_table = "eskimo-bootstrap-tf-lock"
  }
}

provider "aws" {
  region = var.region
}
