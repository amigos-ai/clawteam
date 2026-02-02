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
