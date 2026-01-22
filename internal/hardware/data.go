package hardware

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

//go:embed data/model-recommendations.json
var embeddedRecommendationsData []byte

// loadEmbeddedRecommendations loads the embedded recommendations data
func loadEmbeddedRecommendations() (*RecommendationsData, error) {
	var data RecommendationsData
	if err := json.Unmarshal(embeddedRecommendationsData, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// Recommender provides model recommendations based on hardware
type Recommender struct {
	mu                  sync.RWMutex
	recommendationsData *RecommendationsData
}

// NewRecommender creates a new hardware recommender with embedded data
func NewRecommender() *Recommender {
	r := &Recommender{}
	// Load embedded data immediately
	data, err := loadEmbeddedRecommendations()
	if err == nil {
		r.mu.Lock()
		r.recommendationsData = data
		r.mu.Unlock()
	}
	return r
}

// LoadFromFile loads recommendations from a file (for testing purposes)
// This allows tests to use custom data files
func (r *Recommender) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read data file %s: %w", filePath, err)
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

// LoadRecommendations is a no-op for embedded data (data is loaded in NewRecommender)
// Kept for API compatibility
func (r *Recommender) LoadRecommendations() error {
	if r.recommendationsData == nil {
		data, err := loadEmbeddedRecommendations()
		if err != nil {
			return err
		}
		r.mu.Lock()
		r.recommendationsData = data
		r.mu.Unlock()
	}
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
