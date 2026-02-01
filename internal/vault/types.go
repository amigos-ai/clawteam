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
