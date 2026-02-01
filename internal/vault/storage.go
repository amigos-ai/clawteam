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
