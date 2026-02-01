package vault

import (
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
