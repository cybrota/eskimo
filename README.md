# github-scanner

A pluggable security scanner written in Go. It fetches all repositories in a GitHub organization and runs configured scanners against each repository.

## Usage

```bash
go run ./cmd/github-scanner -org my-org -config scanners.yaml
```

Environment variables required:

- `GITHUB_TOKEN` – Personal access token with rights to read organization repositories
- Scanner specific variables (see `scanners.yaml`)

## Docker

Build and run using Docker:

```bash
docker build -t github-scanner .

docker run -e GITHUB_TOKEN=xxxx -v $HOME/.config:/root/.config github-scanner -org my-org
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

Set `disable: true` to skip running a scanner. If the flag is omitted or set to `false`, the scanner will run.
