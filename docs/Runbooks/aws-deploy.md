# Deploying Eskimo on AWS

This runbook walks through deploying the Eskimo scanner in an AWS account. We start by creating the Terraform backend with a local state file, deploy the main infrastructure using that backend and then migrate the bootstrap state into S3.

## Prerequisites

- Terraform 1.5 or newer
- AWS CLI configured with an account that can create S3 buckets, DynamoDB tables, IAM roles, ECS resources and ECR repositories
- Docker installed if you plan to build the image yourself

## 1. Clone the repository

```bash
git clone https://github.com/cybrota/eskimo.git
cd eskimo
```

## 2. Bootstrap the Terraform backend (local state)

The bootstrap module creates the S3 bucket and DynamoDB tables that store Terraform state. Run it with a **local** backend the first time:

```bash
cd terraform/bootstrap
terraform init -backend=false
terraform apply
```

After this step you will have:

- `eskimo-tf-state` S3 bucket
- `eskimo-infra-tf-lock` DynamoDB table
- `eskimo-bootstrap-tf-lock` DynamoDB table

The bootstrap `terraform.tfstate` stays local for now.

## 3. Deploy AWS infrastructure with remote state

Next initialise the main module with the remote backend that bootstrap created and apply it.

```bash
cd ../aws
terraform init \
  -backend-config="bucket=eskimo-tf-state" \
  -backend-config="key=infra/terraform.tfstate" \
  -backend-config="region=us-west-2" \
  -backend-config="dynamodb_table=eskimo-infra-tf-lock"
terraform apply
```

This creates the VPC, ECS cluster, task definition, EventBridge rules and the Secrets Manager secret used by Eskimo.

## 4. Migrate bootstrap state to S3

To keep all state files in S3, migrate the bootstrap state now:

```bash
cd ../bootstrap
terraform init \
  -migrate-state \
  -backend-config="bucket=eskimo-tf-state" \
  -backend-config="key=bootstrap/terraform.tfstate" \
  -backend-config="region=us-west-2" \
  -backend-config="dynamodb_table=eskimo-bootstrap-tf-lock"
```

## 5. Build and push the Docker image

A ready-to-use `Dockerfile` is provided in the repository. The first part builds the Go binary and then installs common scanners:

```Dockerfile
# ─────────────── Build your Go binary ───────────────
FROM --platform=linux/amd64 golang:1.23 AS builder

WORKDIR /app

# Download deps
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o eskimo .

# ────────── Runtime with Python & Git and HomeBrew installed ─────────
FROM --platform=linux/amd64 python:3.11-slim-bullseye AS runtime

# Install Git and Python
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    git \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# ────────── Install Scanners ──────────
# 1. Semgrep
RUN pip3 install semgrep

# 2. Scharf
RUN curl -sf https://raw.githubusercontent.com/cybrota/scharf/refs/heads/main/install.sh | sh

# 3. Trivy
RUN curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.63.0

# Copy in your Go binary and config
WORKDIR /app
COPY --from=builder /app/eskimo .
COPY scanners.yaml .

# By default, run your Go CLI;
# but Python (and Git) are also available in the container shell if you need them.
ENTRYPOINT ["./eskimo"]
CMD ["-h"]
```

Build the image and push it to the ECR repository created by Terraform:

```bash
docker build -t eskimo .
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-west-2.amazonaws.com

docker tag eskimo:latest <account-id>.dkr.ecr.us-west-2.amazonaws.com/eskimo:latest
docker push <account-id>.dkr.ecr.us-west-2.amazonaws.com/eskimo:latest
```

## 6. Configure scanners

`scanners.yaml` lists all available scanners. Enable or disable them as needed. A sample configuration looks like:

```yaml
scanners:
  - name: semgrep
    command: ["semgrep", "scan"]
    env: ["SEMGREP_PAT_TOKEN"]
  - name: wiz
    pre_command: ["wizcli", "auth"]
    command: ["wizcli", "dir", "scan"]
    env: ["WIZ_CLIENT_ID", "WIZ_CLIENT_SECRET"]
    disable: true
  - name: cycode
    pre_command: ["cycode", "auth"]
    command: ["cycode", "scan", "path", "."]
    env: ["CYCODE_CLIENT_ID", "CYCODE_CLIENT_SECRET"]
    disable: true
  - name: scharf
    command: ["scharf", "audit"]
    env: []
  - name: trivy
    command: ["trivy", "fs", "."]
    env: []
```

Store the required secrets (such as `GITHUB_TOKEN`, `WIZ_CLIENT_ID` and `WIZ_CLIENT_SECRET`) in AWS Secrets Manager under the name configured by the `secret_name` variable (default `eskimo-config`).

## 7. Triggering scans

Scans run automatically according to the cron expression in `scan_schedule_expression` (defaults to every Monday). To run a scan manually use:

```bash
aws events put-events --entries '[{"Source":"eskimo.manual","DetailType":"manual trigger","Detail":"{}"}]'
```

## Cleanup

To remove all resources:

```bash
cd terraform/aws
terraform destroy
cd ../bootstrap
terraform destroy
```

This will remove the ECS resources along with the S3 bucket and DynamoDB tables used for state storage.
