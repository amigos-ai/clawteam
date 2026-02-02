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
	if port1 != BasePort {
		t.Errorf("First port should be %d, got %d", BasePort, port1)
	}

	// Save an instance using that port
	if err := reg.Save(&Instance{Name: "bot1", Port: port1, Status: StatusRunning}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	port2, err := AllocatePort(reg)
	if err != nil {
		t.Fatalf("Second AllocatePort failed: %v", err)
	}
	if port2 != BasePort+1 {
		t.Errorf("Second port should be %d, got %d", BasePort+1, port2)
	}
}
