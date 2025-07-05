# AGENTS Guidelines

This repository follows these guidelines for contributions by AI agents or humans:

## General

1. **Commit Messages**: Use [Conventional Commits](https://www.conventionalcommits.org/) format. Examples include:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `test:` for test-related changes
   - `chore:` for maintenance tasks

2. **Simplicity First**: Prefer simpler implementations over overly complex solutions.

3. **Uniform Structure**: Maintain a consistent code structure across modules so files and packages are easy to navigate.

4. **Explain Why**: Add comments explaining *why* something is done if it is not obvious from code alone.

5. **Copyright Header**: Add the following header at the beginning of every new `.go` code file created as part of PR:

  ```
  Copyright (c) 2025 Naren Yellavula & Cybrota contributors
  Apache License, Version 2.0

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  ```
6. **Branch Names**: Use '_type_/_short_topic_' convention for new branches (e.g. feat/add-s3-backup).

7. **Architectural Decision Records (ADRs)**: For non-trivial design choices, add a short ADR (docs/adr/NNN-*.md) explaining context, the decision, and alternatives.

## Go Related

1. **Run Tests**: Always run tests before committing to ensure functionality and catch regressions. Use `go test ./...` for Go modules.

2. **Style & Formatting**: Use opinionated formatters/lints (e.g. gofmt + goimports, golangci-lint) and run them.

3. **Security**: Run go vet, govulncheck to make sure code is free from basic security issues.

4. **Other**:
   * By convention, one-method interfaces are named by the method name plus an -er suffix or similar modification to construct an agent noun: Reader, Writer, Formatter, CloseNotifier etc.
   * When feasible, error strings should identify their origin, such as by having a prefix naming the operation or package that generated the error


## Terraform Related

1. **Style & Formatting**: Run TFLint to make sure Terraform code is linted.

2. **Security**: Run TFSec & Checkov to make sure Terraform is free from critical and high vulnerabilities.

3. **Other**: Follow Least Privilige Principle for creating IAM Roles & Policies.
