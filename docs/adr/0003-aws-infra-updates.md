# 3. AWS Infrastructure Updates

## Status
Accepted

## Context
After implementing scheduled scans on AWS Fargate, we extended the
Terraform configuration to support image publishing from GitHub Actions
and to provide an on‑demand scan trigger. The bootstrap module was also
updated to use the S3 backend so state is shared across environments.

## Decision
- Added an IAM role and OIDC provider allowing GitHub Actions to push
  container images to the ECR repository. The ECS task now pulls the
  latest image digest from ECR to ensure deterministic deployments.
- Introduced a CloudWatch Event rule named `eskimo-manual` so scans can
  be executed on demand via `aws events put-events`.
- Enabled lifecycle policy on the ECR repository to expire untagged
  images after 14 days.
- Migrated the bootstrap configuration to store its state in the same S3
  bucket with DynamoDB locking.

## Consequences
- Container images built in CI can be pushed securely to ECR without
  long-lived credentials.
- Teams can trigger ad‑hoc scans without changing the schedule.
- State management is consistent across bootstrap and infrastructure
  deployments.
- Unused container images are cleaned up automatically.
