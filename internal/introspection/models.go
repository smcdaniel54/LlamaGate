package introspection

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// CaptureModels fetches model list from Ollama and returns a ModelsSnapshot.
func CaptureModels(ctx context.Context, ollamaHost string, timeout time.Duration) (ModelsSnapshot, error) {
	snap := ModelsSnapshot{
		Provider: "ollama",
		Endpoint: redactURL(ollamaHost),
		Models:   nil,
	}
	u, err := url.Parse(ollamaHost)
	if err != nil {
		return snap, err
	}
	u.Path = "/api/tags"
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return snap, err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return snap, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return snap, fmt.Errorf("ollama returned %d", resp.StatusCode)
	}
	var out struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return snap, err
	}
	for _, m := range out.Models {
		snap.Models = append(snap.Models, ModelEntry{
			ID:      m.Name,
			Name:    m.Name,
			OwnedBy: "ollama",
			Enabled: true,
		})
	}
	if len(snap.Models) > 0 {
		snap.DefaultModel = snap.Models[0].ID
	}
	return snap, nil
}

// redactURL removes credentials from URL for display.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid>"
	}
	if u.User != nil {
		u.User = nil
	}
	return u.String()
}
