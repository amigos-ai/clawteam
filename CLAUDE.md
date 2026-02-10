# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Build the CLI binary
go build -o clawteam ./cmd/clawteam

# Build the Docker image (used by instances)
docker build -t clawteam:latest .
# or via the CLI itself:
./clawteam build

# Run all tests
go test ./...

# Run a single package's tests
go test ./internal/vault/... -v

# Tidy dependencies
go mod tidy
```

## Architecture

ClawTeam is a CLI tool that manages multiple isolated OpenClaw bot instances running in Docker containers. Each instance gets its own container, port, credentials, and persistent storage.

### Package structure

- **`cmd/clawteam/`** — Cobra CLI commands. Each file maps to a command group: `vault.go` (credential management), `instance.go` (create/start/stop/delete/logs/pair), `build.go` (Docker image build).
- **`internal/instance/`** — Core orchestration. `Manager` is the central coordinator that wires together Docker, Vault, and the Registry. `Registry` persists instance metadata to `~/.clawteam/instances.json`. `health.go` polls the gateway `/health` endpoint with container crash detection.
- **`internal/docker/`** — Thin wrapper over the Docker SDK (`github.com/docker/docker`). Exposes simplified container CRUD, log streaming, and exec.
- **`internal/vault/`** — JSON file-based credential storage in `~/.clawteam/vault/`. Supports API keys and SSH keys. Includes path traversal protection.
- **`internal/ui/`** — Reusable terminal components. Spinner with TTY detection (no-op when piped), lobster-themed animation frames.

### Key data flow

`CLI command` → `instance.Manager` → orchestrates `docker.Client` + `vault.Storage` + `Registry`

When creating an instance, the Manager: allocates a port (range 18789–19000), generates a gateway token, loads credentials from the vault, creates the Docker container with env vars and volume mounts, starts it, then polls `/health` until ready.

### Docker setup

The Dockerfile extends `ghcr.io/openclaw/openclaw:2026.1.30` with git and SSH support. The entrypoint script (`scripts/entrypoint.sh`) copies SSH keys and configures git before starting the OpenClaw gateway. Container port 18789 is mapped to the allocated host port.

### Local data layout

```
~/.clawteam/
├── vault/              # Credential JSON files
├── instances.json      # Instance registry
├── ssh/{name}/         # SSH keys per instance
└── volumes/{name}/     # Persistent data (config, memory, workspace)
```
