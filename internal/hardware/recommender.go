package hardware

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

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

// Recommender provides model recommendations based on hardware
type Recommender struct {
	dataFilePath        string
	mu                  sync.RWMutex
	recommendationsData *RecommendationsData
}

// NewRecommender creates a new hardware recommender
func NewRecommender(dataFilePath string) *Recommender {
	return &Recommender{
		dataFilePath: dataFilePath,
	}
}

// LoadRecommendations loads the recommendations data from JSON file
func (r *Recommender) LoadRecommendations() error {
	// Try relative path first, then absolute
	dataPath := r.dataFilePath
	if !filepath.IsAbs(dataPath) {
		// Try to find the file relative to current working directory or executable
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			// Try common locations
			possiblePaths := []string{
				filepath.Join("internal", "extensions", "builtin", "hardware", "data", "model-recommendations.json"),
				filepath.Join(".", "internal", "extensions", "builtin", "hardware", "data", "model-recommendations.json"),
			}
			for _, path := range possiblePaths {
				if _, err := os.Stat(path); err == nil {
					dataPath = path
					break
				}
			}
		}
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		return fmt.Errorf("failed to read data file %s: %w", dataPath, err)
	}

	var recommendationsData RecommendationsData
	if err := json.Unmarshal(data, &recommendationsData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	r.mu.Lock()
	r.recommendationsData = &recommendationsData
	r.mu.Unlock()

	return nil
}

// ClassifyHardwareGroup matches hardware specs to a hardware group
func (r *Recommender) ClassifyHardwareGroup(specs *Specs) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.recommendationsData == nil {
		return "", fmt.Errorf("recommendations data not loaded")
	}

	// Primary check: GPU/VRAM detection (most important for model selection)
	if specs.GPUDetected && specs.GPUVRAMGB > 0 {
		// GPU-based groups
		for _, group := range r.recommendationsData.HardwareGroups {
			if !group.Criteria.RequiresGPU {
				continue
			}

			// Check VRAM criteria
			if group.Criteria.MinVRAMGB != nil && specs.GPUVRAMGB < *group.Criteria.MinVRAMGB {
				continue
			}
			if group.Criteria.MaxVRAMGB != nil && specs.GPUVRAMGB > *group.Criteria.MaxVRAMGB {
				continue
			}

			// Check RAM criteria if specified
			if group.Criteria.MinRAMGB != nil && specs.TotalRAMGB < *group.Criteria.MinRAMGB {
				continue
			}
			if group.Criteria.MaxRAMGB != nil && specs.TotalRAMGB > *group.Criteria.MaxRAMGB {
				continue
			}

			return group.ID, nil
		}
	} else {
		// CPU-only groups
		for _, group := range r.recommendationsData.HardwareGroups {
			if group.Criteria.RequiresGPU {
				continue
			}

			// Check RAM criteria
			if group.Criteria.MinRAMGB != nil && specs.TotalRAMGB < *group.Criteria.MinRAMGB {
				continue
			}
			if group.Criteria.MaxRAMGB != nil && specs.TotalRAMGB > *group.Criteria.MaxRAMGB {
				continue
			}

			return group.ID, nil
		}
	}

	// Fallback: return most common group (CPU-only, 32-64GB)
	return "cpu_only_32_64gb", nil
}

// GetRecommendations returns model recommendations for a hardware group, sorted by priority (1 = highest)
func (r *Recommender) GetRecommendations(groupID string) ([]ModelRecommendation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.recommendationsData == nil {
		return nil, fmt.Errorf("recommendations data not loaded")
	}

	for _, group := range r.recommendationsData.HardwareGroups {
		if group.ID == groupID {
			// Return models sorted by priority (1 = highest priority, 2 = second, etc.)
			// Models are already stored in priority order in JSON, but we ensure they're sorted
			models := make([]ModelRecommendation, len(group.Models))
			copy(models, group.Models)
			
			// Sort by priority (ascending: 1, 2, 3...)
			// Using insertion sort since the list is typically small (< 10 items)
			for i := 1; i < len(models); i++ {
				key := models[i]
				j := i - 1
				for j >= 0 && models[j].Priority > key.Priority {
					models[j+1] = models[j]
					j--
				}
				models[j+1] = key
			}
			
			return models, nil
		}
	}

	return nil, fmt.Errorf("hardware group not found: %s", groupID)
}
