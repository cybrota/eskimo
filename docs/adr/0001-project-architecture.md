# 1. Project Architecture

## Status
Accepted

## Context

Eskimo is a command line tool for orchestrating security scanning across all repositories in a GitHub organization. Users configure a list of scanners via `scanners.yaml` and authorize access through GitHub's device flow. The tool clones repositories locally and executes each configured scanner.

## Decision

We separate concerns into small packages to keep the codebase easy to maintain:

- **cmd** – Cobra-based CLI commands for authentication and running scans.
- **internal/auth** – GitHub device flow implementation and token storage helpers.
- **internal/config** – YAML loader that filters out disabled scanners.
- **internal/github** – Lightweight GitHub wrapper for listing repositories and cloning them using the configured token.
- **internal/scanner** – Runs shell commands for each scanner with environment variables passed through.

This layout keeps dependencies clear: the CLI coordinates actions while internal packages expose reusable logic. Each component has unit tests for core functionality.

## Consequences

- Adding new scanner types only requires updating `scanners.yaml`; no code changes are needed.
- GitHub API usage is isolated so authentication and repository fetching can evolve independently.
- The design favors simplicity and composability over complex orchestration frameworks.
