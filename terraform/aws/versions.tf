terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "eskimo-tf-state"
    key            = "infra/terraform.tfstate"
    region         = "us-west-2"
    dynamodb_table = "eskimo-infra-tf-lock"
  }
}

provider "aws" {
  region = var.region
}
