# 2. AWS Infrastructure for Scheduled Scans

## Status
Accepted

## Context
Eskimo needs to run weekly security scans without manual intervention. Deploying the scanner as a container on AWS Fargate keeps operations simple while leveraging managed services like CloudWatch Events and Secrets Manager.

## Decision
We provision an ECS Fargate cluster and a task definition that runs the container built from this repository. A CloudWatch Event rule triggers `ecs:RunTask` on a cron schedule every Monday. Secrets required by the scanner (GitHub token and Wiz credentials) are stored in AWS Secrets Manager and passed to the task. Terraform state is kept in an S3 bucket with locking in DynamoDB. Network resources, the ECS cluster, and the ECR repository are created via community Terraform modules.

## Consequences
- Scans run automatically each week with logs stored in CloudWatch.
- Using Fargate means tasks stop after completion, so there is no compute cost outside of scan runs.
- The infrastructure can be recreated consistently from Terraform state.
