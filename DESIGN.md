# Eskimo Design

## Overview
Eskimo is a Go-based command line tool that builds a repeatable pipeline for running multiple security scanners across every repository in a GitHub organization. Users authenticate via GitHub’s device flow, define scanners in YAML, and let the CLI orchestrate repository discovery, cloning, and scanner execution in parallel. Supporting assets such as Terraform modules, runbooks, and install scripts provide deployment options ranging from laptops to AWS Fargate.

## Runtime Architecture

- **Entry Point (`main.go`)** – boots the Cobra root command and delegates all logic to the CLI layer.
- **CLI Layer (`cmd/`)** – exposes two commands:
  - `eskimo` (root) pulls the GitHub token, loads scanner configuration, discovers repositories, clones them to `/tmp/github-repos`, and runs every configured scanner per repo.
  - `eskimo auth` drives the GitHub device-flow exchange, writing the resulting token to `~/.config/eskimo/token`.
- **Domain Packages (`internal/`)** – the root command stays thin by leaning on focused packages:
  - `internal/auth` handles device flow, token persistence, and default browser invocation.
  - `internal/config` loads `scanners.yaml`, drops disabled entries, and surfaces runnable scanner definitions.
  - `internal/github` wraps `go-github` to list organization repositories and clone or update them via `git`.
  - `internal/scanner` executes the pre-command/command pairs with environment inheritance and combined output capture.

The CLI coordinates these pieces using context propagation so cancellation signals flow through to subprocesses.

### Scanner Pipeline
1. **Configuration** – `config.Load` returns active scanner definitions, each with optional `pre_command`, `command`, environment variable names, and a disable flag.
2. **Repository Discovery** – `github.Client.ListRepos` paginates through the organization via the GitHub API.
3. **Clone/Update** – `CloneRepo` clones (depth 1) into `/tmp/github-repos/<name>` or runs `git pull` when the repo already exists.
4. **Execution** – For each repository:
   - A worker pool (sized to `runtime.NumCPU`) gates concurrent repo processing.
   - Each repo fan-outs scanners in goroutines so independent scanner runtimes do not block one another.
   - `scanner.Scanner.Run` injects requested environment variables and runs the commands via `exec.CommandContext`, honoring context cancellation.
5. **Logging** – scanner output and errors funnel into a buffered channel consumed by a single goroutine for orderly console logging.

### Authentication Flow
`eskimo auth` reads `GITHUB_CLIENT_ID`, requests a device code, opens a browser via `DefaultBrowser`, and polls for completion. Tokens persist under `~/.config/eskimo/token` (0600 permissions); later `eskimo` runs reuse the `GITHUB_TOKEN` environment variable or the stored file.

### Error Handling & Resilience
- API and Git operations return wrapped errors with context.
- Cloning failures log and skip the offending repo without halting the entire run.
- Pre-command failures short-circuit the scanner but still capture stderr/stdout for visibility.
- Device flow uses exponential backoff when GitHub asks clients to slow down.

### External Dependencies
- `github.com/spf13/cobra` for CLI surfaces.
- `github.com/google/go-github` plus `oauth2` for API access.
- System `git` binary for cloning/pulling.
- Scanner binaries (Semgrep, Wiz CLI, etc.) supplied externally and referenced in `scanners.yaml`.

## Infrastructure & Deployment

- **Terraform (`terraform/`)**
  - `bootstrap/` provisions the remote state backend (versioned S3 bucket and DynamoDB lock tables) with KMS encryption.
  - `aws/` builds the production stack:
    - VPC with public subnets for Fargate tasks.
    - ECS cluster and Fargate task definition mounting an encrypted EFS volume at `/tmp` so multiple scans can share cached clones.
    - ECR repository, EventBridge schedules (weekly cron + manual trigger rule), and IAM roles for task execution, OIDC-based image pushes, and Secrets Manager access.
    - Secrets Manager secret with automatic rotation through a placeholder Lambda (`rotate.py`) and DLQ handling.
  - `terraform/README.md` plus the runbook in `docs/Runbooks/aws-deploy.md` explain bootstrap, deployment, and image publishing steps.
- **Dockerfile** builds a static Eskimo binary and bundles common scanners (Semgrep, Scharf, Trivy) to match the configuration defaults.
- **`install.sh`** provides a cross-platform installer that fetches the latest release artifact for macOS and Linux.

## Folder Structure

- `main.go` – binary entry point.
- `cmd/` – Cobra commands:
  - `root.go` – scan orchestration.
  - `auth.go` – device-flow authentication command.
- `internal/auth/` – device flow client, token load/save, default browser helper.
- `internal/config/` – scanner YAML parsing and filtering.
- `internal/github/` – GitHub client wrapper and git clone/pull helpers.
- `internal/scanner/` – execution harness for pre-commands and scanners.
- `docs/adr/` – project decisions (e.g., architecture, AWS infra updates, parallel scanning).
- `docs/Runbooks/` – operational playbooks (currently AWS deployment).
- `terraform/` – infrastructure as code (bootstrap + AWS stack).
- `Dockerfile` – container build (multi-stage).
- `scanners.yaml` – sample scanner configuration consumed by `internal/config`.
- `results.sarif` – example aggregated scan results.
- `install.sh` – release installation helper script.

The repository also ships a prebuilt `eskimo` binary for convenience during development.

## Testing & Quality
- Unit tests cover authentication helpers, GitHub interactions (via fakes), configuration loading, and scanner execution contracts.
- Recommended validation commands: `go test ./...`, `golangci-lint`, `go vet`, and `govulncheck` (per AGENTS guidelines).
- Terraform modules expect `terraform validate`, `tflint`, `tfsec`, and `checkov` before deployment.

## Extending Eskimo
- **Adding a scanner** – amend `scanners.yaml`; code changes are unnecessary unless new orchestration logic is required.
- **New CLI commands** – place command definitions under `cmd/` and wire them into `rootCmd` or sibling commands.
- **Infrastructure changes** – capture non-trivial updates in `docs/adr/` and extend Terraform modules, keeping module boundaries (`bootstrap` vs `aws`) intact.

