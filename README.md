# Eskimo

<picture width="500">
  <source
    width="100%"
    media="(prefers-color-scheme: dark)"
    src="https://raw.githubusercontent.com/cybrota/eskimo/refs/heads/main/logo.png"
    alt="Cybrota Eskimo Logo (dark)"
  />
  <img
    width="100%"
    src="https://raw.githubusercontent.com/cybrota/eskimo/refs/heads/main/logo.png"
    alt="Cybrota Eskimo Logo (light)"
  />
</picture>

A pluggable security scanner written in Go. It fetches all repositories in a GitHub organization and runs configured scanners against each repository.
The scanners can be configured with commands to run, and their environment variables. Eskimo is useful for setting up weekly scans on cloud environments to generate a baseline scan for organization.

Eskimo is a BYOS (Bring Your Own Scanner) tool that performs security scans based on given configuration. Plug in or out your favorite scanners.

## Usage

```bash
go run ./cmd/eskimo -org my-org -config scanners.yaml
```

Repositories are cloned under `/tmp/github-repos`. If a repository directory
already exists, the latest changes will be pulled before running scanners.

Environment variables required:

- `GITHUB_TOKEN` – Personal access token with rights to read organization repositories
- Scanner specific variables (see `scanners.yaml`)

## Docker

Build and run using Docker:

```bash
docker build -t eskimo .

docker run -e GITHUB_TOKEN=xxxx -v $HOME/.config:/root/.config eskimo -org my-org
```

## Configuration

`scanners.yaml` defines scanners and the commands executed for each repository. Example:

```yaml
scanners:
  - name: semgrep
    command: ["semgrep", "ci", "--pro"]
    env: ["SEMGREP_PAT_TOKEN"]
    disable: false
  - name: wiz
    pre_command: ["wizcli", "auth"]
    command: ["wizcli", "dir", "scan"]
    env: ["WIZ_CLIENT_ID", "WIZ_CLIENT_SECRET"]
    disable: true
```

Set `disable: true` to skip running a scanner. If the flag is omitted or set to `false`, the scanner will run by default.
