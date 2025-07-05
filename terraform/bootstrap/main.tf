data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "kms" {
  statement {
    actions   = ["kms:*"]
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    resources = ["*"]
  }
}

module "state_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "~> 4.0"

  bucket = var.state_bucket_name

  versioning = {
    enabled = true
  }

  server_side_encryption_configuration = {
    rule = {
      apply_server_side_encryption_by_default = {
        sse_algorithm = "AES256"
      }
    }
  }
}

resource "aws_kms_key" "dynamo" {
  description         = "Key for DynamoDB tables"
  policy              = data.aws_iam_policy_document.kms.json
  enable_key_rotation = true
}

resource "aws_dynamodb_table" "infra_tf_lock" {
  name         = var.infra_lock_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.dynamo.arn
  }

  point_in_time_recovery {
    enabled = true
  }

  attribute {
    name = "LockID"
    type = "S"
  }
}

resource "aws_dynamodb_table" "bootstrap_tf_lock" {
  name         = var.bootstrap_lock_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.dynamo.arn
  }

  point_in_time_recovery {
    enabled = true
  }

  attribute {
    name = "LockID"
    type = "S"
  }
}
