# ClawTeam

ClawTeam is an orchestration CLI for managing isolated [OpenClaw](https://github.com/openclaw/openclaw) instances in Docker containers. It handles the full lifecycle — creating containers with injected credentials, managing persistent storage, allocating ports, and pairing browser sessions — so you can run multiple independent bots from one machine.

## Prerequisites

- **Docker** running locally
- **ClawTeam Docker image** built from the included Dockerfile:
  ```
  clawteam build
  ```
  This extends the official `ghcr.io/openclaw/openclaw:2026.1.30` image with git and SSH support.

## Quick start

```bash
# 1. Store your API key in the vault
clawteam vault add prod --type api-key --value "sk-ant-..."

# 2. Create an instance
clawteam create mybot --anthropic prod

# 3. Open the printed URL in your browser, then approve the pairing
clawteam pair mybot
```

## Data storage

All state lives under `~/.clawteam/`:

```
~/.clawteam/
├── instances.json          # Registry of all instances
├── vault/                  # Credential files (one JSON per credential)
├── ssh/<name>/id_rsa       # SSH keys copied for container mounting
└── volumes/<name>/
    ├── config/             # OpenClaw config (always persisted)
    ├── memory/             # Agent memory (persistence: full or memory)
    └── workspace/          # Project files (persistence: full only)
```

---

## Commands

### `clawteam vault` — Credential management

Store API keys and SSH keys in a local vault. Credentials are referenced by name when creating instances, so one key can be shared across multiple bots.

#### `clawteam vault add <name>`

Add a credential to the vault.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | `-t` | `api-key` | Credential type: `api-key` or `ssh-key` |
| `--value` | `-v` | | The credential value as a string |
| `--file` | `-f` | | Read the value from a file instead |

You must provide either `--value` or `--file`.

```
$ clawteam vault add prod --type api-key --value "sk-ant-abc123..."
Added credential 'prod' (api-key)

$ clawteam vault add my-ssh --type ssh-key --file ~/.ssh/id_ed25519
Added credential 'my-ssh' (ssh-key)
```

#### `clawteam vault list`

List all stored credentials (names and types only — values are never printed).

```
$ clawteam vault list
prod                 api-key
my-ssh               ssh-key
```

When the vault is empty:

```
$ clawteam vault list
No credentials in vault
```

#### `clawteam vault remove <name>`

Remove a credential from the vault.

```
$ clawteam vault remove prod
Removed credential 'prod'
```

---

### `clawteam create <name>` — Create a new instance

Creates a Docker container, starts it, waits for the OpenClaw gateway to become ready (showing a lobster spinner animation), then prints the access URL.

| Flag | Default | Description |
|------|---------|-------------|
| `--anthropic` | | Vault credential name for the Anthropic API key |
| `--openai` | | Vault credential name for the OpenAI API key |
| `--ssh` | | Vault credential name for an SSH key (mounted into the container) |
| `--git-name` | | `git config user.name` inside the container |
| `--git-email` | | `git config user.email` inside the container |
| `--persistence` | `full` | Persistence level: `full`, `memory`, or `minimal` |
| `--port` | auto | Host port (auto-assigned from 18789–19000 if omitted) |

**Persistence levels:**

| Level | What survives restarts | Use case |
|-------|------------------------|----------|
| `full` | Config + agent memory + workspace files | Long-running bot with project state |
| `memory` | Config + agent memory | Bot that remembers context but gets a fresh workspace |
| `minimal` | Config only | Disposable bot, clean slate each restart |

**Output:**

```
$ clawteam create mybot --anthropic prod --ssh my-ssh --git-name "My Bot" --git-email "bot@example.com"
🦞 Starting gateway...
Gateway ready!
Instance 'mybot' created:
  URL: http://localhost:18789/?token=a1b2c3d4...

Open the URL in your browser, then run:
  clawteam pair mybot
```

If the gateway doesn't become ready within 60 seconds (or the container crashes), you get a warning but the URL is still printed:

```
Warning: gateway may not be ready: gateway did not become ready: context deadline exceeded
Instance 'mybot' created:
  URL: http://localhost:18789/?token=a1b2c3d4...

Open the URL in your browser, then run:
  clawteam pair mybot
```

When piped or in a non-TTY context, the spinner animation is suppressed — only the final text output is printed.

---

### `clawteam list` — List all instances

Shows a table of all instances with their status and access URL.

```
$ clawteam list
NAME    STATUS   URL
mybot   running  http://localhost:18789/?token=a1b2c3d4...
other   stopped  -
```

Stopped instances show `-` for the URL. When no instances exist:

```
$ clawteam list
No instances
```

---

### `clawteam start <name>` — Start a stopped instance

Starts the container and waits for the gateway to become ready (same spinner behavior as `create`).

```
$ clawteam start mybot
🦞 Starting gateway...
Gateway ready!
Instance 'mybot' started
```

---

### `clawteam stop <name>` — Stop a running instance

Stops the container. Persistent data (volumes) is preserved.

```
$ clawteam stop mybot
Instance 'mybot' stopped
```

---

### `clawteam delete <name>` — Delete an instance

Removes the container (force-stops if running) and deletes the instance from the registry. Also cleans up the SSH key file for this instance.

Note: volume data under `~/.clawteam/volumes/<name>/` is **not** deleted automatically.

```
$ clawteam delete mybot
Instance 'mybot' deleted
```

---

### `clawteam logs <name>` — View instance logs

Streams stdout and stderr from the container.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--follow` | `-f` | `false` | Continuously follow new log output (like `tail -f`) |

```
$ clawteam logs mybot
[gateway] Listening on 0.0.0.0:18789
[gateway] Token authentication enabled
...

$ clawteam logs mybot -f
# Streams continuously until Ctrl+C
```

---

### `clawteam pair <name>` — Approve device pairings

When you open the gateway URL in a browser, the device needs to be approved before it can connect. This command lists pending pairing requests and approves them.

| Flag | Default | Description |
|------|---------|-------------|
| `--id` | | Approve only a specific pairing request by its request ID |

**Approve all pending devices:**

```
$ clawteam pair mybot
Approving device req_abc123 (192.168.1.42)
Approving device req_def456 (192.168.1.43)

Approved 2 device(s). Refresh your browser to connect.
```

**Approve a specific device:**

```
$ clawteam pair mybot --id req_abc123
Approved device req_abc123
```

**No pending requests:**

```
$ clawteam pair mybot
No pending pairing requests
```

---

### `clawteam build` — Build the Docker image

Builds the ClawTeam Docker image (`clawteam:latest`) from the project's Dockerfile. This image extends the official OpenClaw image with git and SSH support.

```
$ clawteam build
Building ClawTeam Docker image...
Note: This requires 'openclaw:local' base image.
Run OpenClaw's docker-setup.sh first if you haven't.
...docker build output...
ClawTeam image built successfully!
```

---

## Typical workflow

```bash
# One-time setup: store credentials
clawteam vault add prod --type api-key --value "sk-ant-..."
clawteam vault add github-key --type ssh-key --file ~/.ssh/id_ed25519

# Create an instance
clawteam create alice \
  --anthropic prod \
  --ssh github-key \
  --git-name "Alice Bot" \
  --git-email "alice@example.com" \
  --persistence full

# Open the URL in your browser, then pair
clawteam pair alice

# Day-to-day management
clawteam list
clawteam logs alice -f
clawteam stop alice
clawteam start alice

# Clean up
clawteam delete alice

# You can run multiple instances simultaneously
clawteam create bob --anthropic prod --persistence minimal --port 18800
clawteam create carol --anthropic prod --persistence memory
clawteam list
```

## How it works under the hood

When you run `clawteam create mybot --anthropic prod`:

1. **Port allocation** — Finds the next unused port in the 18789–19000 range
2. **Token generation** — Creates a random 64-character hex gateway token
3. **Credential loading** — Reads the `prod` credential from `~/.clawteam/vault/`
4. **Container creation** — Calls the Docker API to create a container from `clawteam:latest` with:
   - Environment variables: `ANTHROPIC_API_KEY`, `OPENCLAW_GATEWAY_TOKEN`, etc.
   - Volume mounts for persistence and SSH keys
   - Port binding: container `18789/tcp` → host allocated port
5. **Container start** — Starts the container; the entrypoint script sets up SSH keys and git config
6. **Health polling** — Polls `GET http://localhost:<port>/health` every 500ms until it returns 200, with a 60s timeout. Checks that the container is still running between polls to fail fast on crashes.
7. **Registry update** — Saves instance metadata to `~/.clawteam/instances.json`

Each container runs the OpenClaw gateway process (`node openclaw.mjs gateway --allow-unconfigured --bind lan`), which serves a web UI on the bound port. The gateway token is included in the URL as a query parameter for authentication.
