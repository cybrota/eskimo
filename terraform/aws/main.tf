# VPC for Fargate tasks
locals {
  azs = ["${var.region}a", "${var.region}b"]
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "eskimo-vpc"
  cidr = "10.0.0.0/16"

  azs            = local.azs
  public_subnets = ["10.0.0.0/24", "10.0.1.0/24"]

  enable_nat_gateway = false
  single_nat_gateway = false
  enable_vpn_gateway = false
}

# ECS Cluster
module "ecs_cluster" {
  source  = "terraform-aws-modules/ecs/aws//modules/cluster"
  version = "~> 5.0"

  cluster_name = var.cluster_name
}

# ECR Repository for the scanner image
module "ecr" {
  source  = "terraform-aws-modules/ecr/aws"
  version = "~> 1.5"

  repository_name         = "eskimo"
  create_lifecycle_policy = true
  repository_lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Expire untagged images older than 14 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 14
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

data "aws_caller_identity" "current" {}

resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]
}

data "aws_iam_policy_document" "github_ecr_trust" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }
    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${var.github_repository}:*"]
    }
  }
}

resource "aws_iam_role" "github_ecr" {
  name               = "eskimo-github-ecr"
  assume_role_policy = data.aws_iam_policy_document.github_ecr_trust.json
}

resource "aws_iam_role_policy_attachment" "github_ecr" {
  role       = aws_iam_role.github_ecr.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"
}

# Secrets Manager
resource "aws_secretsmanager_secret" "scanner" {
  name = var.secret_name
}

resource "aws_secretsmanager_secret_version" "scanner" {
  secret_id     = aws_secretsmanager_secret.scanner.id
  secret_string = jsonencode(var.secret_values)
}

# IAM role for ECS task execution
data "aws_iam_policy_document" "task_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "task_exec" {
  name               = "eskimo-task-exec"
  assume_role_policy = data.aws_iam_policy_document.task_assume.json
}

resource "aws_iam_role_policy_attachment" "task_exec" {
  role       = aws_iam_role.task_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Allow task to read secrets
data "aws_iam_policy_document" "secret_access" {
  statement {
    actions   = ["secretsmanager:GetSecretValue"]
    resources = [aws_secretsmanager_secret.scanner.arn]
  }
}

resource "aws_iam_policy" "secret_access" {
  name   = "eskimo-secret-access"
  policy = data.aws_iam_policy_document.secret_access.json
}

resource "aws_iam_role_policy_attachment" "secret_access" {
  role       = aws_iam_role.task_exec.name
  policy_arn = aws_iam_policy.secret_access.arn
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/eskimo"
  retention_in_days = 30
}

locals {
  repository_url = module.ecr.repository_url
  repository_name = element(
    split("/", local.repository_url),
    length(split("/", local.repository_url)) - 1
  )
  image = "${module.ecr.repository_url}@${data.aws_ecr_image.latest.image_digest}"

}

# Latest image digest
data "aws_ecr_image" "latest" {
  repository_name = local.repository_name
  most_recent     = true
}

# ECS Task Definition

resource "aws_ecs_task_definition" "scan" {
  family                   = "eskimo-scan"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "512"
  memory                   = "1024"
  execution_role_arn       = aws_iam_role.task_exec.arn
  task_role_arn            = aws_iam_role.task_exec.arn

  container_definitions = jsonencode([
    {
      name      = "eskimo"
      image     = local.image
      essential = true
      command   = ["--org", var.github_org, "--config", "/app/scanners.yaml"]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.ecs.name
          awslogs-region        = var.region
          awslogs-stream-prefix = "eskimo"
        }
      }
      secrets = [
        { name = "GITHUB_TOKEN", valueFrom = "${aws_secretsmanager_secret.scanner.arn}:GITHUB_TOKEN::" },
        { name = "WIZ_CLIENT_ID", valueFrom = "${aws_secretsmanager_secret.scanner.arn}:WIZ_CLIENT_ID::" },
        { name = "WIZ_CLIENT_SECRET", valueFrom = "${aws_secretsmanager_secret.scanner.arn}:WIZ_CLIENT_SECRET::" }
      ]
    }
  ])
}

# Security group for tasks
resource "aws_security_group" "ecs_tasks" {
  name_prefix = "ecs-tasks-"
  vpc_id      = module.vpc.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# CloudWatch Event rule to trigger weekly
resource "aws_cloudwatch_event_rule" "weekly" {
  name                = "eskimo-weekly"
  schedule_expression = "cron(${var.scan_schedule_expression})"
}

# IAM role for EventBridge to run tasks
data "aws_iam_policy_document" "event_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "event" {
  name               = "ecs-scan-event"
  assume_role_policy = data.aws_iam_policy_document.event_assume.json
}

data "aws_iam_policy_document" "event" {
  statement {
    actions   = ["ecs:RunTask"]
    resources = [aws_ecs_task_definition.scan.arn]
  }
  statement {
    actions   = ["iam:PassRole"]
    resources = [aws_iam_role.task_exec.arn]
  }
}

resource "aws_iam_role_policy" "event" {
  role   = aws_iam_role.event.id
  policy = data.aws_iam_policy_document.event.json
}

resource "aws_cloudwatch_event_target" "ecs" {
  rule      = aws_cloudwatch_event_rule.weekly.name
  target_id = "EcsTask"
  arn       = module.ecs_cluster.arn
  role_arn  = aws_iam_role.event.arn

  ecs_target {
    task_definition_arn = aws_ecs_task_definition.scan.arn
    launch_type         = "FARGATE"
    network_configuration {
      subnets          = module.vpc.public_subnets
      security_groups  = [aws_security_group.ecs_tasks.id]
      assign_public_ip = true
    }
    platform_version = "LATEST"
  }
}

# CloudWatch Event rule for manual trigger
resource "aws_cloudwatch_event_rule" "manual" {
  name = "eskimo-manual"
  event_pattern = jsonencode({
    source = ["eskimo.manual"]
  })
}

resource "aws_cloudwatch_event_target" "manual" {
  rule      = aws_cloudwatch_event_rule.manual.name
  target_id = "ManualEcsTask"
  arn       = module.ecs_cluster.arn
  role_arn  = aws_iam_role.event.arn

  ecs_target {
    task_definition_arn = aws_ecs_task_definition.scan.arn
    launch_type         = "FARGATE"
    network_configuration {
      subnets          = module.vpc.public_subnets
      security_groups  = [aws_security_group.ecs_tasks.id]
      assign_public_ip = true
    }
    platform_version = "LATEST"
  }
}
