package introspection

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// CaptureHealth checks backend (Ollama) reachability and returns a HealthSnapshot.
func CaptureHealth(ctx context.Context, ollamaHost string, timeout time.Duration) HealthSnapshot {
	snap := HealthSnapshot{
		Backend:   "ollama",
		Status:    "unknown",
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
	u, err := url.Parse(ollamaHost)
	if err != nil {
		snap.Status = "unknown"
		snap.Message = "invalid backend URL"
		return snap
	}
	u.Path = "/api/tags"
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		snap.Message = err.Error()
		return snap
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		snap.Status = "unreachable"
		snap.Message = err.Error()
		return snap
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		snap.Status = "ok"
	} else {
		snap.Status = "unreachable"
		snap.Message = fmt.Sprintf("status %d", resp.StatusCode)
	}
	return snap
}
