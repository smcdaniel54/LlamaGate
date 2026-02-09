package introspection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/llamagate/llamagate/internal/config"
	"github.com/llamagate/llamagate/internal/memory"
)

// SystemCardMaxChars is the hard cap for the injected system card (target ~40 tokens).
const SystemCardMaxChars = 320

// BuildSystemCard returns a short system message for chat injection, or empty if disabled.
// It includes hardware summary, models summary, LlamaGate info, and optionally user pinned memory.
func BuildSystemCard(cfg *config.Config, store memory.Store, ollamaHost string, userID string, timeout time.Duration) string {
	if cfg == nil || cfg.Introspection == nil || !cfg.Introspection.ChatInject.Enabled {
		return ""
	}
	ci := &cfg.Introspection.ChatInject
	var parts []string

	if ci.IncludeLlamaGateInfo {
		rt := CaptureRuntime()
		parts = append(parts, fmt.Sprintf("LlamaGate %s (Go %s).", rt.AppVersion, rt.GoVersion))
		if cfg.MCP != nil && cfg.MCP.Enabled {
			parts = append(parts, "MCP tools enabled.")
		}
		if cfg.Memory != nil && cfg.Memory.Enabled {
			parts = append(parts, "Memory enabled.")
		}
	}

	if ci.IncludeHardware {
		level := DetailMinimal
		if cfg.Introspection.Hardware.DetailLevel == "standard" || cfg.Introspection.Hardware.DetailLevel == "full" {
			level = DetailLevel(cfg.Introspection.Hardware.DetailLevel)
		}
		dataDir := ""
		if cfg.Memory != nil {
			dataDir = cfg.Memory.Dir
		}
		hw := CaptureHardwareWithTimeout(dataDir, level, timeout)
		parts = append(parts, fmt.Sprintf("Host: %s/%s, %d cores.", hw.OS, hw.Arch, hw.CPU.LogicalCores))
		if hw.Memory.TotalBytes > 0 {
			parts = append(parts, fmt.Sprintf("RAM: %d MB total.", hw.Memory.TotalBytes/(1024*1024)))
		}
	}

	if ci.IncludeModels {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		snap, err := CaptureModels(ctx, ollamaHost, timeout)
		cancel()
		if err == nil && len(snap.Models) > 0 {
			names := make([]string, 0, len(snap.Models))
			for _, m := range snap.Models {
				names = append(names, m.Name)
				if len(names) >= 5 {
					break
				}
			}
			parts = append(parts, fmt.Sprintf("Models: %s.", strings.Join(names, ", ")))
		} else {
			parts = append(parts, "Models: (unable to list).")
		}
	}

	if store != nil && cfg.Memory != nil && cfg.Memory.Enabled && userID != "" {
		um, err := store.GetUserMemory(userID)
		if err == nil && len(um.Pinned) > 0 {
			pinned := strings.Join(um.Pinned, "; ")
			if len(pinned) > 80 {
				pinned = pinned[:77] + "..."
			}
			parts = append(parts, "User context: "+pinned)
		}
	}

	if len(parts) == 0 {
		return ""
	}
	s := strings.TrimSpace(strings.Join(parts, " "))
	if len(s) > SystemCardMaxChars {
		s = s[:SystemCardMaxChars-3] + "..."
	}
	return s
}
