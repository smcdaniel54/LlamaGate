package memory

import "time"

// SystemMemory holds system-level memory (capabilities, notes, optional snapshots).
type SystemMemory struct {
	UpdatedAt           time.Time       `json:"updated_at"`
	Notes               []string       `json:"notes,omitempty"`
	Capabilities        *Capabilities   `json:"capabilities,omitempty"`
	LastHardwareSnapshot interface{}   `json:"last_hardware_snapshot,omitempty"`
	LastModelsSnapshot   interface{}   `json:"last_models_snapshot,omitempty"`
}

// Capabilities describes what the gateway supports.
type Capabilities struct {
	Endpoints       []string `json:"endpoints,omitempty"`
	MCPEnabled      bool     `json:"mcp_enabled"`
	MemoryEnabled   bool     `json:"memory_enabled"`
	IntrospectionEnabled bool `json:"introspection_enabled"`
}

// UserMemory holds per-user memory (pinned, recent, tags).
type UserMemory struct {
	UserID    string    `json:"user_id"`
	UpdatedAt time.Time `json:"updated_at"`
	Pinned    []string  `json:"pinned,omitempty"`
	Recent    []string  `json:"recent,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
}

// MemoryStatus is the summary returned by MemoryStatus() for /v1/system/memory.
type MemoryStatus struct {
	SystemUpdatedAt   *time.Time `json:"system_updated_at,omitempty"`
	UserCount         int       `json:"user_count"`
	SessionCount      int       `json:"session_count"`
	TotalSizeBytes    int64     `json:"total_size_bytes,omitempty"`
}
