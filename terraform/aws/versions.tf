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
    key            = "eksimo/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "eskimo-tf-lock"
  }
}

provider "aws" {
  region = var.region
}
