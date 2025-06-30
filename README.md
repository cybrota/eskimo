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

Bring Your Own Scanners (BYOS) to your organization’s CI/CD pipelines. Eskimo discovers all repositories in a GitHub organization, clones them locally, and runs each scanner you configure—generating a snapshot baseline of your security posture on demand or on a schedule.

## Why Use Eskimo?

- **Centralized Scanning**
  Automatically fetches every repository in your organization, so you never miss a repo or branch.

- **Configurable Scanners**
  Define any scanner with its commands and environment variables in a simple YAML file.

- **Baseline & Audit**
  Run daily or weekly to capture a baseline, then compare over time or spot new issues immediately.

- **Lightweight & Portable**
  A single Go binary with zero external dependencies—runs anywhere Docker runs or as a native executable.

## Key Features

- **Auto-clone & Pull**
  Clones each repo under `/tmp/github-repos`; if already present, pulls the latest changes.

- **Flexible Configuration**
  Define multiple scanners in `scanners.yaml`, toggle them on or off, set pre-commands, and pass custom env vars.

- **Device-Flow Authentication**
  Securely authenticate via GitHub’s device flow—no browser embeds in CI required.


## Supported Platforms

**Binary**
- Linux
- macOS

**Cloud**
- AWS

## Installation

**Option 1: Script Install**
```sh
curl -sfL https://raw.githubusercontent.com/cybrota/eskimo/refs/heads/main/install.sh | sh
```

This will download the latest Eskimo binary and make it available on your $PATH.

**Option 2: Download Prebuilt Binary**

1. Go to the Releases page

2. Download the archive for your OS and architecture

3. Unpack and move eskimo into a directory on your $PATH

## Usage Examples

1. Run Scanners Against an Organization
```sh
# Uses scanners.yaml in current directory
eskimo scan --org my-org
```

Or explicitly specify your config file:

```sh
eskimo scan --org my-org --config /path/to/scanners.yaml
```

2. Authenticate via Device Flow
```sh
eskimo auth --org my-org
```

Follows GitHub’s device-flow: you’ll get a code to paste at github.com/device.

## The Risk of Unscanned Repositories

Without a centralized scanner, it’s easy to overlook new or forked repos—exposing your organization to unpatched vulnerabilities, drifted dependencies, or misconfigured workflows. Eskimo ensures every repo is covered by automating manual process of scanning code everytime there is a new scanner or a repo.

## TODO

 - Webhook integration for real-time alerts to MS Teams or Slack

 - Send to SIEM systems

## Further Reading
- GitHub Supply-Chain Security: https://docs.github.com/en/code-security/supply-chain-security

- Device Flow Authentication: https://docs.github.com/en/developers/apps/authorizing-oauth-apps#device-flow
