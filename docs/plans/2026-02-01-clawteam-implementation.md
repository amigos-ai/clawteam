# ClawTeam Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a CLI tool that orchestrates isolated OpenClaw instances in Docker containers with per-instance credentials and configurable persistence.

**Architecture:** Go CLI using Cobra for commands, Docker SDK for container management. Credentials stored in `~/.clawteam/vault/`, instance state in `~/.clawteam/instances.json`. Custom Docker image extends `openclaw:local` with SSH/git support.

**Tech Stack:** Go 1.21+, Cobra (CLI), Docker SDK for Go, JSON for config storage

---

## Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/clawteam/main.go`
- Create: `.gitignore`

**Step 1: Initialize Go module**

Run: `go mod init github.com/amigos-ai/clawteam`
Expected: Creates `go.mod` file

**Step 2: Create main.go skeleton**

Create `cmd/clawteam/main.go`:

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("clawteam - OpenClaw orchestration tool")
	return nil
}
```

**Step 3: Create .gitignore**

Create `.gitignore`:

```
# Binaries
clawteam
*.exe

# Go
vendor/

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store

# Local config (for testing)
.clawteam/
```

**Step 4: Verify it compiles**

Run: `go build -o clawteam ./cmd/clawteam`
Expected: Binary `clawteam` created without errors

**Step 5: Verify it runs**

Run: `./clawteam`
Expected: Output `clawteam - OpenClaw orchestration tool`

**Step 6: Commit**

```bash
git add go.mod cmd/ .gitignore
git commit -m "chore: initial project scaffolding"
```

---

## Task 2: Add Cobra CLI Framework

**Files:**
- Modify: `cmd/clawteam/main.go`
- Modify: `go.mod`, `go.sum`

**Step 1: Add Cobra dependency**

Run: `go get github.com/spf13/cobra@latest`
Expected: Cobra added to go.mod

**Step 2: Rewrite main.go with Cobra root command**

Replace `cmd/clawteam/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clawteam",
	Short: "OpenClaw orchestration tool",
	Long:  "Manage isolated OpenClaw instances in Docker containers",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Verify it compiles and runs**

Run: `go build -o clawteam ./cmd/clawteam && ./clawteam --help`
Expected: Shows Cobra help output with "OpenClaw orchestration tool"

**Step 4: Commit**

```bash
git add go.mod go.sum cmd/
git commit -m "feat: add Cobra CLI framework"
```

---

## Task 3: Vault Data Types and Storage

**Files:**
- Create: `internal/vault/types.go`
- Create: `internal/vault/storage.go`
- Create: `internal/vault/storage_test.go`

**Step 1: Create vault types**

Create `internal/vault/types.go`:

```go
package vault

import "time"

type CredentialType string

const (
	TypeAPIKey CredentialType = "api-key"
	TypeSSHKey CredentialType = "ssh-key"
)

type Credential struct {
	Name      string         `json:"name"`
	Type      CredentialType `json:"type"`
	Value     string         `json:"value"`
	CreatedAt time.Time      `json:"created_at"`
}
```

**Step 2: Write failing test for Save/Load**

Create `internal/vault/storage_test.go`:

```go
package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	cred := &Credential{
		Name:      "test-key",
		Type:      TypeAPIKey,
		Value:     "sk-test-123",
		CreatedAt: time.Now(),
	}

	// Save
	err := store.Save(cred)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := store.Load("test-key")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Name != cred.Name {
		t.Errorf("Name mismatch: got %s, want %s", loaded.Name, cred.Name)
	}
	if loaded.Value != cred.Value {
		t.Errorf("Value mismatch: got %s, want %s", loaded.Value, cred.Value)
	}
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/vault/... -v`
Expected: FAIL - NewStorage not defined

**Step 4: Implement Storage**

Create `internal/vault/storage.go`:

```go
package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Storage struct {
	dir string
}

func NewStorage(dir string) *Storage {
	return &Storage{dir: dir}
}

func (s *Storage) Save(cred *Credential) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}

	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credential: %w", err)
	}

	path := filepath.Join(s.dir, cred.Name+".json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write credential: %w", err)
	}

	return nil
}

func (s *Storage) Load(name string) (*Credential, error) {
	path := filepath.Join(s.dir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read credential: %w", err)
	}

	var cred Credential
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, fmt.Errorf("unmarshal credential: %w", err)
	}

	return &cred, nil
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/vault/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/vault/
git commit -m "feat: add vault credential storage"
```

---

## Task 4: Vault List and Remove

**Files:**
- Modify: `internal/vault/storage.go`
- Modify: `internal/vault/storage_test.go`

**Step 1: Write failing test for List**

Add to `internal/vault/storage_test.go`:

```go
func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	// Add two credentials
	store.Save(&Credential{Name: "key1", Type: TypeAPIKey, Value: "v1"})
	store.Save(&Credential{Name: "key2", Type: TypeSSHKey, Value: "v2"})

	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 credentials, got %d", len(list))
	}
}

func TestRemove(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	store.Save(&Credential{Name: "to-delete", Type: TypeAPIKey, Value: "v"})

	err := store.Remove("to-delete")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	_, err = store.Load("to-delete")
	if err == nil {
		t.Error("Expected error loading removed credential")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vault/... -v`
Expected: FAIL - List and Remove not defined

**Step 3: Implement List and Remove**

Add to `internal/vault/storage.go`:

```go
func (s *Storage) List() ([]*Credential, error) {
	entries, err := os.ReadDir(s.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read vault dir: %w", err)
	}

	var creds []*Credential
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		name := entry.Name()[:len(entry.Name())-5] // remove .json
		cred, err := s.Load(name)
		if err != nil {
			continue
		}
		creds = append(creds, cred)
	}

	return creds, nil
}

func (s *Storage) Remove(name string) error {
	path := filepath.Join(s.dir, name+".json")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove credential: %w", err)
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/vault/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/vault/
git commit -m "feat: add vault list and remove"
```

---

## Task 5: Vault CLI Commands

**Files:**
- Create: `cmd/clawteam/vault.go`
- Modify: `cmd/clawteam/main.go`

**Step 1: Create vault command group**

Create `cmd/clawteam/vault.go`:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/amigos-ai/clawteam/internal/vault"
	"github.com/spf13/cobra"
)

func getVaultStorage() *vault.Storage {
	home, _ := os.UserHomeDir()
	return vault.NewStorage(filepath.Join(home, ".clawteam", "vault"))
}

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage credentials",
}

var vaultAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a credential to the vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		credType, _ := cmd.Flags().GetString("type")
		value, _ := cmd.Flags().GetString("value")
		file, _ := cmd.Flags().GetString("file")

		if value == "" && file == "" {
			return fmt.Errorf("must provide --value or --file")
		}

		if file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
			value = string(data)
		}

		cred := &vault.Credential{
			Name:      name,
			Type:      vault.CredentialType(credType),
			Value:     value,
			CreatedAt: time.Now(),
		}

		store := getVaultStorage()
		if err := store.Save(cred); err != nil {
			return err
		}

		fmt.Printf("Added credential '%s' (%s)\n", name, credType)
		return nil
	},
}

var vaultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := getVaultStorage()
		creds, err := store.List()
		if err != nil {
			return err
		}

		if len(creds) == 0 {
			fmt.Println("No credentials in vault")
			return nil
		}

		for _, c := range creds {
			fmt.Printf("%-20s %s\n", c.Name, c.Type)
		}
		return nil
	},
}

var vaultRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a credential from the vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store := getVaultStorage()
		if err := store.Remove(name); err != nil {
			return err
		}
		fmt.Printf("Removed credential '%s'\n", name)
		return nil
	},
}

func init() {
	vaultAddCmd.Flags().StringP("type", "t", "api-key", "Credential type (api-key, ssh-key)")
	vaultAddCmd.Flags().StringP("value", "v", "", "Credential value")
	vaultAddCmd.Flags().StringP("file", "f", "", "Read value from file")

	vaultCmd.AddCommand(vaultAddCmd)
	vaultCmd.AddCommand(vaultListCmd)
	vaultCmd.AddCommand(vaultRemoveCmd)
}
```

**Step 2: Register vault command in main.go**

Add to `cmd/clawteam/main.go` before `func main()`:

```go
func init() {
	rootCmd.AddCommand(vaultCmd)
}
```

**Step 3: Verify it compiles**

Run: `go build -o clawteam ./cmd/clawteam`
Expected: Compiles without errors

**Step 4: Test vault commands manually**

Run: `./clawteam vault add test-key --type api-key --value "sk-test"`
Expected: `Added credential 'test-key' (api-key)`

Run: `./clawteam vault list`
Expected: Shows `test-key              api-key`

Run: `./clawteam vault remove test-key`
Expected: `Removed credential 'test-key'`

**Step 5: Commit**

```bash
git add cmd/clawteam/
git commit -m "feat: add vault CLI commands"
```

---

## Task 6: Instance Data Types

**Files:**
- Create: `internal/instance/types.go`
- Create: `internal/instance/registry.go`
- Create: `internal/instance/registry_test.go`

**Step 1: Create instance types**

Create `internal/instance/types.go`:

```go
package instance

import "time"

type PersistenceLevel string

const (
	PersistFull    PersistenceLevel = "full"
	PersistMemory  PersistenceLevel = "memory"
	PersistMinimal PersistenceLevel = "minimal"
)

type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
)

type Credentials struct {
	AnthropicKeyRef string `json:"anthropic_key_ref,omitempty"`
	OpenAIKeyRef    string `json:"openai_key_ref,omitempty"`
	SSHKeyRef       string `json:"ssh_key_ref,omitempty"`
}

type Instance struct {
	Name        string           `json:"name"`
	Status      Status           `json:"status"`
	Port        int              `json:"port"`
	CreatedAt   time.Time        `json:"created_at"`
	Persistence PersistenceLevel `json:"persistence"`
	Credentials Credentials      `json:"credentials"`
	ContainerID string           `json:"container_id"`
	GitName     string           `json:"git_name,omitempty"`
	GitEmail    string           `json:"git_email,omitempty"`
}
```

**Step 2: Write failing test for registry**

Create `internal/instance/registry_test.go`:

```go
package instance

import (
	"testing"
	"time"
)

func TestRegistrySaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	reg := NewRegistry(tmpDir)

	inst := &Instance{
		Name:        "test-bot",
		Status:      StatusRunning,
		Port:        18789,
		CreatedAt:   time.Now(),
		Persistence: PersistFull,
		ContainerID: "abc123",
	}

	err := reg.Save(inst)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := reg.Get("test-bot")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if loaded.Name != inst.Name {
		t.Errorf("Name mismatch: got %s, want %s", loaded.Name, inst.Name)
	}
	if loaded.Port != inst.Port {
		t.Errorf("Port mismatch: got %d, want %d", loaded.Port, inst.Port)
	}
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/instance/... -v`
Expected: FAIL - NewRegistry not defined

**Step 4: Implement Registry**

Create `internal/instance/registry.go`:

```go
package instance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Registry struct {
	dir  string
	path string
}

type registryData struct {
	Instances map[string]*Instance `json:"instances"`
}

func NewRegistry(dir string) *Registry {
	return &Registry{
		dir:  dir,
		path: filepath.Join(dir, "instances.json"),
	}
}

func (r *Registry) load() (*registryData, error) {
	data, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return &registryData{Instances: make(map[string]*Instance)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read registry: %w", err)
	}

	var reg registryData
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("unmarshal registry: %w", err)
	}
	if reg.Instances == nil {
		reg.Instances = make(map[string]*Instance)
	}

	return &reg, nil
}

func (r *Registry) save(data *registryData) error {
	if err := os.MkdirAll(r.dir, 0755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(r.path, bytes, 0644); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}

	return nil
}

func (r *Registry) Save(inst *Instance) error {
	data, err := r.load()
	if err != nil {
		return err
	}

	data.Instances[inst.Name] = inst
	return r.save(data)
}

func (r *Registry) Get(name string) (*Instance, error) {
	data, err := r.load()
	if err != nil {
		return nil, err
	}

	inst, ok := data.Instances[name]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", name)
	}

	return inst, nil
}

func (r *Registry) List() ([]*Instance, error) {
	data, err := r.load()
	if err != nil {
		return nil, err
	}

	var instances []*Instance
	for _, inst := range data.Instances {
		instances = append(instances, inst)
	}

	return instances, nil
}

func (r *Registry) Delete(name string) error {
	data, err := r.load()
	if err != nil {
		return err
	}

	delete(data.Instances, name)
	return r.save(data)
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/instance/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/instance/
git commit -m "feat: add instance registry"
```

---

## Task 7: Port Allocator

**Files:**
- Create: `internal/instance/ports.go`
- Create: `internal/instance/ports_test.go`

**Step 1: Write failing test for port allocation**

Create `internal/instance/ports_test.go`:

```go
package instance

import (
	"testing"
)

func TestAllocatePort(t *testing.T) {
	tmpDir := t.TempDir()
	reg := NewRegistry(tmpDir)

	port1, err := AllocatePort(reg)
	if err != nil {
		t.Fatalf("AllocatePort failed: %v", err)
	}
	if port1 != 18789 {
		t.Errorf("First port should be 18789, got %d", port1)
	}

	// Save an instance using that port
	reg.Save(&Instance{Name: "bot1", Port: port1, Status: StatusRunning})

	port2, err := AllocatePort(reg)
	if err != nil {
		t.Fatalf("Second AllocatePort failed: %v", err)
	}
	if port2 != 18790 {
		t.Errorf("Second port should be 18790, got %d", port2)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/instance/... -v -run TestAllocatePort`
Expected: FAIL - AllocatePort not defined

**Step 3: Implement AllocatePort**

Create `internal/instance/ports.go`:

```go
package instance

const (
	BasePort = 18789
	MaxPort  = 19000
)

func AllocatePort(reg *Registry) (int, error) {
	instances, err := reg.List()
	if err != nil {
		return 0, err
	}

	usedPorts := make(map[int]bool)
	for _, inst := range instances {
		usedPorts[inst.Port] = true
	}

	for port := BasePort; port < MaxPort; port++ {
		if !usedPorts[port] {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", BasePort, MaxPort)
}
```

**Step 4: Add missing import**

Add `"fmt"` to the imports in `internal/instance/ports.go`.

**Step 5: Run test to verify it passes**

Run: `go test ./internal/instance/... -v -run TestAllocatePort`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/instance/
git commit -m "feat: add port allocator"
```

---

## Task 8: Docker Image (Dockerfile and Entrypoint)

**Files:**
- Create: `Dockerfile`
- Create: `scripts/entrypoint.sh`

**Step 1: Create entrypoint script**

Create `scripts/entrypoint.sh`:

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

**Step 2: Create Dockerfile**

Create `Dockerfile`:

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
    printf "Host *\n  IdentityFile /home/node/.ssh/id_rsa\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null\n" \
    > /home/node/.ssh/config && \
    chown -R node:node /home/node/.ssh

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
```

**Step 3: Make entrypoint executable**

Run: `chmod +x scripts/entrypoint.sh`
Expected: No output, file now executable

**Step 4: Commit**

```bash
git add Dockerfile scripts/
git commit -m "feat: add Docker image files"
```

---

## Task 9: Docker Client Wrapper

**Files:**
- Create: `internal/docker/client.go`
- Modify: `go.mod`, `go.sum`

**Step 1: Add Docker SDK dependency**

Run: `go get github.com/docker/docker@latest`
Expected: Docker SDK added to go.mod

**Step 2: Create Docker client wrapper**

Create `internal/docker/client.go`:

```go
package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

type CreateOptions struct {
	Name         string
	Image        string
	Port         int
	EnvVars      map[string]string
	Mounts       []MountConfig
}

type MountConfig struct {
	Source   string
	Target   string
	ReadOnly bool
}

func (c *Client) CreateContainer(ctx context.Context, opts CreateOptions) (string, error) {
	// Build environment variables
	var env []string
	for k, v := range opts.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Build mounts
	var mounts []mount.Mount
	for _, m := range opts.Mounts {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	// Port binding
	containerPort := nat.Port("18789/tcp")
	hostPort := fmt.Sprintf("%d", opts.Port)

	resp, err := c.cli.ContainerCreate(ctx,
		&container.Config{
			Image: opts.Image,
			Env:   env,
			ExposedPorts: nat.PortSet{
				containerPort: struct{}{},
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				containerPort: []nat.PortBinding{
					{HostIP: "127.0.0.1", HostPort: hostPort},
				},
			},
			Mounts: mounts,
		},
		nil,
		nil,
		opts.Name,
	)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}

	return resp.ID, nil
}

func (c *Client) StartContainer(ctx context.Context, id string) error {
	return c.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (c *Client) StopContainer(ctx context.Context, id string) error {
	return c.cli.ContainerStop(ctx, id, container.StopOptions{})
}

func (c *Client) RemoveContainer(ctx context.Context, id string) error {
	return c.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

func (c *Client) ContainerLogs(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	return c.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
	})
}
```

**Step 3: Verify it compiles**

Run: `go build ./internal/docker/...`
Expected: Compiles without errors

**Step 4: Commit**

```bash
git add internal/docker/ go.mod go.sum
git commit -m "feat: add Docker client wrapper"
```

---

## Task 10: Instance Manager

**Files:**
- Create: `internal/instance/manager.go`

**Step 1: Create instance manager**

Create `internal/instance/manager.go`:

```go
package instance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/amigos-ai/clawteam/internal/docker"
	"github.com/amigos-ai/clawteam/internal/vault"
)

const ImageName = "clawteam:latest"

type Manager struct {
	registry *Registry
	vault    *vault.Storage
	docker   *docker.Client
	dataDir  string
}

func NewManager(dataDir string) (*Manager, error) {
	dockerClient, err := docker.NewClient()
	if err != nil {
		return nil, err
	}

	return &Manager{
		registry: NewRegistry(dataDir),
		vault:    vault.NewStorage(filepath.Join(dataDir, "vault")),
		docker:   dockerClient,
		dataDir:  dataDir,
	}, nil
}

func (m *Manager) Close() error {
	return m.docker.Close()
}

type CreateOptions struct {
	Name           string
	AnthropicRef   string
	OpenAIRef      string
	SSHRef         string
	GitName        string
	GitEmail       string
	Persistence    PersistenceLevel
	Port           int
}

func (m *Manager) Create(ctx context.Context, opts CreateOptions) (*Instance, error) {
	// Check if instance already exists
	if _, err := m.registry.Get(opts.Name); err == nil {
		return nil, fmt.Errorf("instance '%s' already exists", opts.Name)
	}

	// Allocate port if not specified
	port := opts.Port
	if port == 0 {
		var err error
		port, err = AllocatePort(m.registry)
		if err != nil {
			return nil, err
		}
	}

	// Build environment variables from vault refs
	envVars := make(map[string]string)

	if opts.AnthropicRef != "" {
		cred, err := m.vault.Load(opts.AnthropicRef)
		if err != nil {
			return nil, fmt.Errorf("load anthropic key: %w", err)
		}
		envVars["ANTHROPIC_API_KEY"] = cred.Value
	}

	if opts.OpenAIRef != "" {
		cred, err := m.vault.Load(opts.OpenAIRef)
		if err != nil {
			return nil, fmt.Errorf("load openai key: %w", err)
		}
		envVars["OPENAI_API_KEY"] = cred.Value
	}

	if opts.GitName != "" {
		envVars["GIT_USER_NAME"] = opts.GitName
	}
	if opts.GitEmail != "" {
		envVars["GIT_USER_EMAIL"] = opts.GitEmail
	}

	// Build mounts
	var mounts []docker.MountConfig

	// SSH key mount
	if opts.SSHRef != "" {
		cred, err := m.vault.Load(opts.SSHRef)
		if err != nil {
			return nil, fmt.Errorf("load ssh key: %w", err)
		}

		// Write SSH key to temp file for mounting
		sshDir := filepath.Join(m.dataDir, "ssh", opts.Name)
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return nil, fmt.Errorf("create ssh dir: %w", err)
		}
		sshKeyPath := filepath.Join(sshDir, "id_rsa")
		if err := os.WriteFile(sshKeyPath, []byte(cred.Value), 0600); err != nil {
			return nil, fmt.Errorf("write ssh key: %w", err)
		}

		mounts = append(mounts, docker.MountConfig{
			Source:   sshKeyPath,
			Target:   "/secrets/ssh/id_rsa",
			ReadOnly: true,
		})
	}

	// Persistence mounts
	volumeDir := filepath.Join(m.dataDir, "volumes", opts.Name)
	if err := os.MkdirAll(volumeDir, 0755); err != nil {
		return nil, fmt.Errorf("create volume dir: %w", err)
	}

	persistence := opts.Persistence
	if persistence == "" {
		persistence = PersistFull
	}

	// Config always persisted
	configDir := filepath.Join(volumeDir, "config")
	os.MkdirAll(configDir, 0755)
	mounts = append(mounts, docker.MountConfig{
		Source: configDir,
		Target: "/home/node/.openclaw",
	})

	if persistence == PersistFull || persistence == PersistMemory {
		memoryDir := filepath.Join(volumeDir, "memory")
		os.MkdirAll(memoryDir, 0755)
		mounts = append(mounts, docker.MountConfig{
			Source: memoryDir,
			Target: "/home/node/memory",
		})
	}

	if persistence == PersistFull {
		workspaceDir := filepath.Join(volumeDir, "workspace")
		os.MkdirAll(workspaceDir, 0755)
		mounts = append(mounts, docker.MountConfig{
			Source: workspaceDir,
			Target: "/home/node/workspace",
		})
	}

	// Create container
	containerID, err := m.docker.CreateContainer(ctx, docker.CreateOptions{
		Name:    "clawteam-" + opts.Name,
		Image:   ImageName,
		Port:    port,
		EnvVars: envVars,
		Mounts:  mounts,
	})
	if err != nil {
		return nil, err
	}

	// Start container
	if err := m.docker.StartContainer(ctx, containerID); err != nil {
		m.docker.RemoveContainer(ctx, containerID)
		return nil, fmt.Errorf("start container: %w", err)
	}

	// Save to registry
	inst := &Instance{
		Name:        opts.Name,
		Status:      StatusRunning,
		Port:        port,
		CreatedAt:   time.Now(),
		Persistence: persistence,
		Credentials: Credentials{
			AnthropicKeyRef: opts.AnthropicRef,
			OpenAIKeyRef:    opts.OpenAIRef,
			SSHKeyRef:       opts.SSHRef,
		},
		ContainerID: containerID,
		GitName:     opts.GitName,
		GitEmail:    opts.GitEmail,
	}

	if err := m.registry.Save(inst); err != nil {
		return nil, err
	}

	return inst, nil
}

func (m *Manager) List() ([]*Instance, error) {
	return m.registry.List()
}

func (m *Manager) Get(name string) (*Instance, error) {
	return m.registry.Get(name)
}

func (m *Manager) Start(ctx context.Context, name string) error {
	inst, err := m.registry.Get(name)
	if err != nil {
		return err
	}

	if err := m.docker.StartContainer(ctx, inst.ContainerID); err != nil {
		return err
	}

	inst.Status = StatusRunning
	return m.registry.Save(inst)
}

func (m *Manager) Stop(ctx context.Context, name string) error {
	inst, err := m.registry.Get(name)
	if err != nil {
		return err
	}

	if err := m.docker.StopContainer(ctx, inst.ContainerID); err != nil {
		return err
	}

	inst.Status = StatusStopped
	return m.registry.Save(inst)
}

func (m *Manager) Delete(ctx context.Context, name string) error {
	inst, err := m.registry.Get(name)
	if err != nil {
		return err
	}

	// Remove container
	if err := m.docker.RemoveContainer(ctx, inst.ContainerID); err != nil {
		// Continue even if container removal fails
	}

	// Remove SSH key file
	sshDir := filepath.Join(m.dataDir, "ssh", name)
	os.RemoveAll(sshDir)

	// Remove from registry
	return m.registry.Delete(name)
}

func (m *Manager) Logs(ctx context.Context, name string, follow bool) error {
	inst, err := m.registry.Get(name)
	if err != nil {
		return err
	}

	logs, err := m.docker.ContainerLogs(ctx, inst.ContainerID, follow)
	if err != nil {
		return err
	}
	defer logs.Close()

	_, err = io.Copy(os.Stdout, logs)
	return err
}
```

**Step 2: Add missing import**

Add `"io"` to the imports.

**Step 3: Verify it compiles**

Run: `go build ./internal/instance/...`
Expected: Compiles without errors

**Step 4: Commit**

```bash
git add internal/instance/
git commit -m "feat: add instance manager"
```

---

## Task 11: Instance CLI Commands

**Files:**
- Create: `cmd/clawteam/instance.go`
- Modify: `cmd/clawteam/main.go`

**Step 1: Create instance commands**

Create `cmd/clawteam/instance.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/amigos-ai/clawteam/internal/instance"
	"github.com/spf13/cobra"
)

func getManager() (*instance.Manager, error) {
	home, _ := os.UserHomeDir()
	return instance.NewManager(filepath.Join(home, ".clawteam"))
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new OpenClaw instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		anthropic, _ := cmd.Flags().GetString("anthropic")
		openai, _ := cmd.Flags().GetString("openai")
		ssh, _ := cmd.Flags().GetString("ssh")
		gitName, _ := cmd.Flags().GetString("git-name")
		gitEmail, _ := cmd.Flags().GetString("git-email")
		persistence, _ := cmd.Flags().GetString("persistence")
		port, _ := cmd.Flags().GetInt("port")

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		inst, err := mgr.Create(context.Background(), instance.CreateOptions{
			Name:         name,
			AnthropicRef: anthropic,
			OpenAIRef:    openai,
			SSHRef:       ssh,
			GitName:      gitName,
			GitEmail:     gitEmail,
			Persistence:  instance.PersistenceLevel(persistence),
			Port:         port,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Instance '%s' running at http://localhost:%d\n", inst.Name, inst.Port)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		instances, err := mgr.List()
		if err != nil {
			return err
		}

		if len(instances) == 0 {
			fmt.Println("No instances")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS\tPORT\tPERSISTENCE")
		for _, inst := range instances {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", inst.Name, inst.Status, inst.Port, inst.Persistence)
		}
		w.Flush()

		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a stopped instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		if err := mgr.Start(context.Background(), args[0]); err != nil {
			return err
		}

		fmt.Printf("Instance '%s' started\n", args[0])
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a running instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		if err := mgr.Stop(context.Background(), args[0]); err != nil {
			return err
		}

		fmt.Printf("Instance '%s' stopped\n", args[0])
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		if err := mgr.Delete(context.Background(), args[0]); err != nil {
			return err
		}

		fmt.Printf("Instance '%s' deleted\n", args[0])
		return nil
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "View instance logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		follow, _ := cmd.Flags().GetBool("follow")

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		return mgr.Logs(context.Background(), args[0], follow)
	},
}

func init() {
	createCmd.Flags().String("anthropic", "", "Vault ref for Anthropic API key")
	createCmd.Flags().String("openai", "", "Vault ref for OpenAI API key")
	createCmd.Flags().String("ssh", "", "Vault ref for SSH key")
	createCmd.Flags().String("git-name", "", "Git user.name")
	createCmd.Flags().String("git-email", "", "Git user.email")
	createCmd.Flags().String("persistence", "full", "Persistence level (full, memory, minimal)")
	createCmd.Flags().Int("port", 0, "Specific port (default: auto-assign)")

	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
}
```

**Step 2: Register commands in main.go**

Replace the `init()` function in `cmd/clawteam/main.go`:

```go
func init() {
	rootCmd.AddCommand(vaultCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(logsCmd)
}
```

**Step 3: Verify it compiles**

Run: `go build -o clawteam ./cmd/clawteam`
Expected: Compiles without errors

**Step 4: Verify help output**

Run: `./clawteam --help`
Expected: Shows all commands (vault, create, list, start, stop, delete, logs)

**Step 5: Commit**

```bash
git add cmd/clawteam/
git commit -m "feat: add instance CLI commands"
```

---

## Task 12: Build Image Command

**Files:**
- Create: `cmd/clawteam/build.go`

**Step 1: Create build command**

Create `cmd/clawteam/build.go`:

```go
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the ClawTeam Docker image",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Building ClawTeam Docker image...")

		// First, ensure openclaw:local exists
		fmt.Println("Note: This requires 'openclaw:local' base image.")
		fmt.Println("Run OpenClaw's docker-setup.sh first if you haven't.")

		buildCmd := exec.Command("docker", "build", "-t", "clawteam:latest", ".")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr

		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("docker build failed: %w", err)
		}

		fmt.Println("ClawTeam image built successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
```

**Step 2: Add to main.go init**

The build command is auto-registered via its `init()` function.

**Step 3: Verify it compiles**

Run: `go build -o clawteam ./cmd/clawteam`
Expected: Compiles without errors

**Step 4: Commit**

```bash
git add cmd/clawteam/build.go
git commit -m "feat: add build command for Docker image"
```

---

## Task 13: Final Integration Test

**Step 1: Build the binary**

Run: `go build -o clawteam ./cmd/clawteam`
Expected: Binary created

**Step 2: Run all unit tests**

Run: `go test ./... -v`
Expected: All tests pass

**Step 3: Test CLI help**

Run: `./clawteam --help`
Expected: Shows all commands

Run: `./clawteam vault --help`
Expected: Shows vault subcommands

Run: `./clawteam create --help`
Expected: Shows create flags

**Step 4: Commit final state**

```bash
git add -A
git commit -m "chore: final cleanup"
```

---

## Summary

After completing all tasks, you will have:

1. **CLI binary** (`clawteam`) with commands:
   - `vault add/list/remove` - Credential management
   - `create/list/start/stop/delete/logs` - Instance lifecycle
   - `build` - Docker image building

2. **Docker image** (`clawteam:latest`) extending OpenClaw with SSH/git support

3. **Data storage** in `~/.clawteam/`:
   - `vault/` - Encrypted credentials
   - `instances.json` - Instance registry
   - `volumes/<name>/` - Persistent data per instance
   - `ssh/<name>/` - SSH keys per instance

4. **Test coverage** for vault and registry modules
