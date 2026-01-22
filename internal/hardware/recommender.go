package hardware

// RecommendationsData represents the model recommendations JSON structure
type RecommendationsData struct {
	Version        string          `json:"version"`
	LastUpdated    string          `json:"last_updated"`
	Source         string          `json:"source"`
	Description    string          `json:"description,omitempty"`
	HardwareGroups []HardwareGroup `json:"hardware_groups"`
}

// HardwareGroup represents a hardware group with criteria and models
type HardwareGroup struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Criteria    HardwareCriteria      `json:"criteria"`
	Models      []ModelRecommendation `json:"models"`
}

// HardwareCriteria defines the criteria for matching hardware to a group
type HardwareCriteria struct {
	MinRAMGB    *int `json:"min_ram_gb,omitempty"`
	MaxRAMGB    *int `json:"max_ram_gb,omitempty"`
	MinVRAMGB   *int `json:"min_vram_gb,omitempty"`
	MaxVRAMGB   *int `json:"max_vram_gb,omitempty"`
	RequiresGPU bool `json:"requires_gpu"`
}

// ModelRecommendation represents a recommended model
type ModelRecommendation struct {
	Name                  string   `json:"name"`
	OllamaName            string   `json:"ollama_name"`
	Priority              int      `json:"priority"`
	Description           string   `json:"description"`
	IntelligenceScore     *float64 `json:"intelligence_score,omitempty"`
	ParametersB           *float64 `json:"parameters_b,omitempty"`
	MinRAMGB              int      `json:"min_ram_gb"`
	MinVRAMGB             int      `json:"min_vram_gb"`
	Quantized             bool     `json:"quantized"`
	OllamaCommand         string   `json:"ollama_command"`
	UseCases              []string `json:"use_cases"`
	ArtificialAnalysisURL string  `json:"artificial_analysis_url,omitempty"`
}

// Recommender struct and methods are now in data.go (using embedded data)
