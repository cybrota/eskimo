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

  tags = var.tags
}

# Access point rooted at "/" with 0777 perms
resource "aws_efs_access_point" "tmp" {
  file_system_id = aws_efs_file_system.tmp.id

  posix_user {
    uid = 1000
    gid = 1000
  }

  # Use a dedicated /tmp directory so tasks can write to the mount root
  root_directory {
    path = "/tmp"
    creation_info {
      owner_uid   = 1000
      owner_gid   = 1000
      permissions = "0777"
    }
  }

  tags = var.tags
}

# Security group and mount targets to make the file system reachable from ECS tasks
resource "aws_security_group" "efs" {
  name_prefix = "efs-"
  description = "Security group for EFS"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "NFS from ECS tasks"
    from_port       = 2049
    to_port         = 2049
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs_tasks.id]
  }

  egress {
    description = "Allow return traffic to ECS tasks"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["10.0.0.0/16"]
  }

  tags = var.tags
}

resource "aws_efs_mount_target" "tmp" {
  for_each = toset(module.vpc.public_subnets)

  file_system_id  = aws_efs_file_system.tmp.id
  subnet_id       = each.value
  security_groups = [aws_security_group.efs.id]
}
