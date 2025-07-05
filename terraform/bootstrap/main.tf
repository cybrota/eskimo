
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "kms" {
  statement {
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    resources = ["arn:aws:kms:${var.region}:${data.aws_caller_identity.current.account_id}:key/*"]
  }
}

module "state_bucket" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-s3-bucket.git?ref=e1fb51bce8008b0d5fd660e81d97ff7a101bdbc6"

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
  enable_key_rotation = true
  policy              = data.aws_iam_policy_document.kms.json
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
