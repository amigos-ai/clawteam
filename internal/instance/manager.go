package instance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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

	// Generate gateway token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate gateway token: %w", err)
	}
	gatewayToken := hex.EncodeToString(tokenBytes)

	// Build environment variables from vault refs
	envVars := make(map[string]string)
	envVars["OPENCLAW_GATEWAY_TOKEN"] = gatewayToken

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
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	mounts = append(mounts, docker.MountConfig{
		Source: configDir,
		Target: "/home/node/.openclaw",
	})

	if persistence == PersistFull || persistence == PersistMemory {
		memoryDir := filepath.Join(volumeDir, "memory")
		if err := os.MkdirAll(memoryDir, 0755); err != nil {
			return nil, fmt.Errorf("create memory dir: %w", err)
		}
		mounts = append(mounts, docker.MountConfig{
			Source: memoryDir,
			Target: "/home/node/memory",
		})
	}

	if persistence == PersistFull {
		workspaceDir := filepath.Join(volumeDir, "workspace")
		if err := os.MkdirAll(workspaceDir, 0755); err != nil {
			return nil, fmt.Errorf("create workspace dir: %w", err)
		}
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
		ContainerID:  containerID,
		GatewayToken: gatewayToken,
		GitName:      opts.GitName,
		GitEmail:     opts.GitEmail,
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

func (m *Manager) Pair(ctx context.Context, name string) ([]PendingDevice, error) {
	inst, err := m.registry.Get(name)
	if err != nil {
		return nil, err
	}

	output, err := m.docker.ExecContainer(ctx, inst.ContainerID, []string{
		"node", "openclaw.mjs", "devices", "list", "--json",
	})
	if err != nil {
		return nil, fmt.Errorf("list pending devices: %w", err)
	}

	var result devicesListResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("parse devices list: %w", err)
	}

	devices := make([]PendingDevice, len(result.Pending))
	for i, p := range result.Pending {
		devices[i] = PendingDevice{
			ID:       p.RequestID,
			DeviceID: p.DeviceID,
			Role:     p.Role,
			IP:       p.RemoteIP,
		}
	}
	return devices, nil
}

type devicesListResult struct {
	Pending []devicesListPending `json:"pending"`
}

type devicesListPending struct {
	RequestID string `json:"requestId"`
	DeviceID  string `json:"deviceId"`
	Role      string `json:"role"`
	RemoteIP  string `json:"remoteIp"`
}

func (m *Manager) ApprovePairing(ctx context.Context, name string, requestID string) error {
	inst, err := m.registry.Get(name)
	if err != nil {
		return err
	}

	_, err = m.docker.ExecContainer(ctx, inst.ContainerID, []string{
		"node", "openclaw.mjs", "devices", "approve", requestID,
	})
	if err != nil {
		return fmt.Errorf("approve pairing: %w", err)
	}

	return nil
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
