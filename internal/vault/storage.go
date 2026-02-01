package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Storage struct {
	dir string
}

func NewStorage(dir string) *Storage {
	return &Storage{dir: dir}
}

// validateName checks that a credential name is safe for use in file paths.
// It rejects empty names, names containing path separators, and special directory names.
func validateName(name string) error {
	if name == "" {
		return errors.New("credential name cannot be empty")
	}
	if strings.ContainsAny(name, "/\\") {
		return errors.New("credential name cannot contain path separators")
	}
	if name == "." || name == ".." {
		return errors.New("credential name cannot be '.' or '..'")
	}
	return nil
}

func (s *Storage) Save(cred *Credential) error {
	if cred == nil {
		return errors.New("credential cannot be nil")
	}
	if err := validateName(cred.Name); err != nil {
		return fmt.Errorf("invalid credential name: %w", err)
	}

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
	if err := validateName(name); err != nil {
		return nil, fmt.Errorf("invalid credential name: %w", err)
	}

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
	if err := validateName(name); err != nil {
		return fmt.Errorf("invalid credential name: %w", err)
	}

	path := filepath.Join(s.dir, name+".json")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove credential: %w", err)
	}
	return nil
}
