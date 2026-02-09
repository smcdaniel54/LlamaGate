package introspection

import "time"

// RuntimeSnapshot holds app runtime info (version, uptime, etc.).
type RuntimeSnapshot struct {
	AppVersion   string    `json:"app_version"`
	GitCommit    string    `json:"git_commit,omitempty"`
	BuildDate    string    `json:"build_date,omitempty"`
	GoVersion    string    `json:"go_version"`
	UptimeSecs   float64   `json:"uptime_secs"`
	ProcessRSS   int64     `json:"process_rss_bytes,omitempty"`
	CurrentTime  time.Time `json:"current_time"`
}

// HardwareSnapshot holds hardware info (best-effort, redacted).
type HardwareSnapshot struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	CPU  struct {
		LogicalCores int    `json:"logical_cores"`
		ModelName    string `json:"model_name,omitempty"`
	} `json:"cpu"`
	Memory struct {
		TotalBytes     int64 `json:"total_bytes,omitempty"`
		AvailableBytes int64 `json:"available_bytes,omitempty"`
	} `json:"memory"`
	Disk struct {
		DataDirTotalBytes int64 `json:"data_dir_total_bytes,omitempty"`
		DataDirFreeBytes  int64 `json:"data_dir_free_bytes,omitempty"`
	} `json:"disk"`
	GPU struct {
		Present   bool   `json:"gpu_present"`
		GPUName   string `json:"gpu_name,omitempty"`
		VRAMBytes int64  `json:"vram_bytes,omitempty"`
	} `json:"gpu"`
}

// ModelsSnapshot holds model/provider info.
type ModelsSnapshot struct {
	Provider   string         `json:"provider"`
	Endpoint   string         `json:"endpoint,omitempty"` // sanitized
	Models     []ModelEntry   `json:"models"`
	DefaultModel string       `json:"default_model,omitempty"`
}

// ModelEntry is a single model in the snapshot.
type ModelEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	OwnedBy  string `json:"owned_by,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// HealthSnapshot holds backend health.
type HealthSnapshot struct {
	Backend   string `json:"backend"`
	Status    string `json:"status"` // "ok", "unreachable", "unknown"
	Message   string `json:"message,omitempty"`
	CheckedAt string `json:"checked_at"`
}

// DetailLevel is the hardware redaction level.
type DetailLevel string

const (
	DetailMinimal  DetailLevel = "minimal"
	DetailStandard DetailLevel = "standard"
	DetailFull     DetailLevel = "full"
)
