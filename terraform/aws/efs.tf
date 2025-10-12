###############
# EFS setup
###############

# Filesystem with Intelligent-Tiering to reduce costs
resource "aws_efs_file_system" "tmp" {
  creation_token = "eskimo-tmp"
  encrypted      = true
  kms_key_id     = aws_kms_key.secrets.arn

  # Transition to Infrequent Access after 30 days (94% cost savings: $0.30 -> $0.016/GB-month)
  lifecycle_policy {
    transition_to_ia = "AFTER_30_DAYS"
  }

  # Automatically transition back to Standard when accessed
  lifecycle_policy {
    transition_to_primary_storage_class = "AFTER_1_ACCESS"
  }
}

# Access point rooted at "/" with 0777 perms
resource "aws_efs_access_point" "tmp" {
  file_system_id = aws_efs_file_system.tmp.id

  posix_user {
    uid = 1000
    gid = 1000
  }

  root_directory {
    path = "/"
    creation_info {
      owner_uid   = 1000
      owner_gid   = 1000
      permissions = "0777"
    }
  }
}
