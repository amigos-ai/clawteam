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
