package instance

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// WaitReady polls the gateway's /health endpoint until it returns 200 or the
// context is cancelled. Before each poll it checks that the container is still
// running so it can fail fast on crashes.
func (m *Manager) WaitReady(ctx context.Context, name string) error {
	inst, err := m.registry.Get(name)
	if err != nil {
		return err
	}

	healthURL := fmt.Sprintf("http://localhost:%d/health", inst.Port)

	client := &http.Client{Timeout: 2 * time.Second}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Check container is still running.
		running, err := m.docker.IsContainerRunning(ctx, inst.ContainerID)
		if err != nil {
			return fmt.Errorf("check container: %w", err)
		}
		if !running {
			return fmt.Errorf("container exited unexpectedly — check logs with: clawteam logs %s", name)
		}

		// Probe the health endpoint.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("gateway did not become ready: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
