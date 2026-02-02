# ClawTeam Design

An orchestration tool for managing isolated OpenClaw instances.

## Problem

Running multiple OpenClaw instances requires manual Docker setup, credential management, and port tracking. There's no easy way to spin up isolated bots with controlled access to API keys and git repos.

## Solution

A CLI tool that manages OpenClaw instances as Docker containers with:
- Per-instance credential injection (API keys, SSH keys)
- Configurable persistence (full state, memory only, or minimal)
- Automatic port allocation
- Simple lifecycle management

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   ClawTeam CLI                          │
│  - Docker container lifecycle                           │
│  - Credential injection (API keys, SSH keys)            │
│  - Volume management (persistence)                      │
│  - Port allocation                                      │
└─────────────────────┬───────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        ▼             ▼             ▼
   ┌─────────┐   ┌─────────┐   ┌─────────┐
   │ Instance│   │ Instance│   │ Instance│
   │  "bot1" │   │  "bot2" │   │  "bot3" │
   │ OpenClaw│   │ OpenClaw│   │ OpenClaw│
   │ :18789  │   │ :18790  │   │ :18791  │
   └────┬────┘   └────┬────┘   └────┬────┘
        │             │             │
        ▼             ▼             ▼
   WhatsApp      Telegram       Discord
```

## Instance Container Design

Each OpenClaw instance runs in a container:

```
┌─────────────────────────────────────────────────────┐
│  OpenClaw Instance Container                        │
│                                                     │
│  Environment Variables:                             │
│  ├── ANTHROPIC_API_KEY=sk-ant-...                   │
│  ├── OPENAI_API_KEY=sk-...                          │
│  └── (other provider keys as needed)                │
│                                                     │
│  Mounted Volumes:                                   │
│  ├── /secrets/ssh/        ← Git SSH keys (read-only)│
│  ├── /home/node/.openclaw ← Config (persistence)    │
│  ├── /home/node/workspace ← Workspace (optional)    │
│  └── /home/node/memory    ← Agent memory (optional) │
│                                                     │
│  Exposed Ports:                                     │
│  └── 18789 → <dynamic-host-port>  (Control UI)     │
│                                                     │
│  Base Image: openclaw:local (from official repo)    │
└─────────────────────────────────────────────────────┘
```

### Persistence Levels

| Level | What's saved | Use case |
|-------|--------------|----------|
| `full` | Config + memory + workspace | Long-running bot with project files |
| `memory` | Config + memory only | Bot that remembers but fresh workspace |
| `minimal` | Config only | Stateless bot, fresh each restart |

## Data Model

Instance registry stored in `~/.clawteam/instances.json`:

```json
{
  "instances": {
    "bot1": {
      "name": "bot1",
      "status": "running",
      "port": 18789,
      "created": "2026-02-01T10:00:00Z",
      "persistence": "full",
      "credentials": {
        "anthropic_key_ref": "vault:anthropic-main",
        "openai_key_ref": "vault:openai-dev",
        "ssh_key_ref": "vault:github-bot1"
      },
      "container_id": "abc123..."
    }
  }
}
```

Credential vault stored in `~/.clawteam/vault/`:
- API keys and SSH keys stored as files
- Referenced by name in instance config
- Can be reused across instances

## CLI Interface

### Vault Management

```bash
clawteam vault add <name> --type <api-key|ssh-key> [--value|--file]
clawteam vault list
clawteam vault remove <name>
```

### Instance Lifecycle

```bash
clawteam create <name> [flags]
  --anthropic <vault-ref>     API key for Anthropic
  --openai <vault-ref>        API key for OpenAI
  --ssh <vault-ref>           SSH key for git
  --git-name <name>           Git user.name
  --git-email <email>         Git user.email
  --persistence <level>       full|memory|minimal (default: full)
  --port <port>               Specific port (default: auto-assign)

clawteam list
clawteam start <name>
clawteam stop <name>
clawteam delete <name>
clawteam logs <name> [-f]
clawteam exec <name> [command]
```

### Example Workflow

```bash
# One-time setup
clawteam vault add anthropic-main --type api-key --value "sk-ant-..."
clawteam vault add my-github --type ssh-key --file ~/.ssh/id_ed25519

# Create a bot
clawteam create mybot \
  --anthropic anthropic-main \
  --ssh my-github \
  --git-name "My Bot" \
  --git-email "bot@me.com"
# Output: Instance 'mybot' running at http://localhost:18789

# Manage it
clawteam list
clawteam stop mybot
clawteam start mybot
clawteam delete mybot
```

## Project Structure

```
clawteam/
├── Dockerfile              # Custom image extending openclaw:local
├── docker-compose.yml      # Template for instance creation
├── cmd/
│   └── clawteam/
│       └── main.go         # CLI entrypoint
├── internal/
│   ├── instance/           # Create, start, stop, delete instances
│   ├── vault/              # Credential storage
│   └── docker/             # Docker API wrapper
├── configs/
│   └── templates/          # Default OpenClaw configs
└── scripts/
    └── entrypoint.sh       # Container entrypoint (SSH setup, etc.)
```

## Dockerfile

```dockerfile
FROM openclaw:local

# Git + SSH essentials
RUN apt-get update && apt-get install -y \
    git \
    openssh-client \
    && rm -rf /var/lib/apt/lists/*

# SSH config for mounted keys
RUN mkdir -p /home/node/.ssh && \
    chmod 700 /home/node/.ssh && \
    echo "Host *\n  IdentityFile /secrets/ssh/id_rsa\n  StrictHostKeyChecking no" \
    > /home/node/.ssh/config && \
    chown -R node:node /home/node/.ssh

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
```

## Entrypoint Script

```bash
#!/bin/bash
set -e

# Link SSH key if provided
if [ -f /secrets/ssh/id_rsa ]; then
    cp /secrets/ssh/id_rsa /home/node/.ssh/id_rsa
    chmod 600 /home/node/.ssh/id_rsa
    chown node:node /home/node/.ssh/id_rsa
fi

# Configure git if details provided
if [ -n "$GIT_USER_NAME" ]; then
    git config --global user.name "$GIT_USER_NAME"
fi
if [ -n "$GIT_USER_EMAIL" ]; then
    git config --global user.email "$GIT_USER_EMAIL"
fi

# Start OpenClaw (existing entrypoint)
exec docker-entrypoint.sh "$@"
```

## Out of Scope (for now)

- Web UI (can be added later on top of CLI)
- Network restrictions per instance
- Resource limits (CPU, memory) per instance
- Credential encryption at rest
- Multi-host deployment

## Open Questions

None - ready for implementation.
