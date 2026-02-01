package vault

import (
	"strings"
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

func TestSave_NilCredential(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	err := store.Save(nil)
	if err == nil {
		t.Fatal("expected error for nil credential, got nil")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("expected error to mention 'nil', got: %v", err)
	}
}

func TestSave_InvalidNames(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	tests := []struct {
		name    string
		wantErr string
	}{
		{"", "empty"},
		{"../etc/passwd", "path separator"},
		{"foo/bar", "path separator"},
		{"foo\\bar", "path separator"},
		{".", "'.' or '..'"},
		{"..", "'.' or '..'"},
	}

	for _, tt := range tests {
		t.Run("name="+tt.name, func(t *testing.T) {
			cred := &Credential{
				Name:      tt.name,
				Type:      TypeAPIKey,
				Value:     "test-value",
				CreatedAt: time.Now(),
			}
			err := store.Save(cred)
			if err == nil {
				t.Fatalf("expected error for name %q, got nil", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error to contain %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestLoad_InvalidNames(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	tests := []struct {
		name    string
		wantErr string
	}{
		{"", "empty"},
		{"../etc/passwd", "path separator"},
		{"foo/bar", "path separator"},
		{"foo\\bar", "path separator"},
		{".", "'.' or '..'"},
		{"..", "'.' or '..'"},
	}

	for _, tt := range tests {
		t.Run("name="+tt.name, func(t *testing.T) {
			_, err := store.Load(tt.name)
			if err == nil {
				t.Fatalf("expected error for name %q, got nil", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error to contain %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

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

func TestRemove_InvalidNames(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	invalidNames := []string{
		"",
		"../etc/passwd",
		"foo/bar",
		"foo\\bar",
		".",
		"..",
	}

	for _, name := range invalidNames {
		t.Run("name="+name, func(t *testing.T) {
			err := store.Remove(name)
			if err == nil {
				t.Errorf("Remove(%q) should fail with invalid name", name)
			}
		})
	}
}
