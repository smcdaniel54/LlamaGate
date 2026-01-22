// Package hardware provides hardware detection and model recommendation capabilities.
package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
	"github.com/rs/zerolog/log"
)

// Recommender implements hardware detection and model recommendations
type Recommender struct {
	name                string
	version             string
	dataFilePath        string
	mu                  sync.RWMutex
	recommendationsData *RecommendationsData
	scriptPath          string
}

// HardwareSpecs represents detected hardware specifications
type HardwareSpecs struct {
	CPUCores     int    `json:"cpu_cores"`
	CPUModel     string `json:"cpu_model"`
	TotalRAMGB   int    `json:"total_ram_gb"`
	GPUDetected  bool   `json:"gpu_detected"`
	GPUName      string `json:"gpu_name"`
	GPUVRAMGB    int    `json:"gpu_vram_gb"`
	DetectionMethod string `json:"detection_method"`
}

// RecommendationsData represents the model recommendations JSON structure
type RecommendationsData struct {
	Version       string         `json:"version"`
	LastUpdated   string         `json:"last_updated"`
	Source        string         `json:"source"`
	HardwareGroups []HardwareGroup `json:"hardware_groups"`
}

// HardwareGroup represents a hardware group with criteria and models
type HardwareGroup struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Criteria    HardwareCriteria      `json:"criteria"`
	Models      []ModelRecommendation  `json:"models"`
}

// HardwareCriteria defines the criteria for matching hardware to a group
type HardwareCriteria struct {
	MinRAMGB    *int  `json:"min_ram_gb,omitempty"`
	MaxRAMGB    *int  `json:"max_ram_gb,omitempty"`
	MinVRAMGB   *int  `json:"min_vram_gb,omitempty"`
	MaxVRAMGB   *int  `json:"max_vram_gb,omitempty"`
	RequiresGPU bool  `json:"requires_gpu"`
}

// ModelRecommendation represents a recommended model
type ModelRecommendation struct {
	Name         string   `json:"name"`
	OllamaName   string   `json:"ollama_name"`
	Priority     int      `json:"priority"`
	Description  string   `json:"description"`
	MinRAMGB     int      `json:"min_ram_gb"`
	MinVRAMGB    int      `json:"min_vram_gb"`
	Quantized    bool     `json:"quantized"`
	OllamaCommand string  `json:"ollama_command"`
	UseCases     []string `json:"use_cases"`
}

// NewRecommender creates a new hardware recommender extension
func NewRecommender(name, version, dataFilePath string) *Recommender {
	// Determine script path based on OS
	var scriptPath string
	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join("internal", "extensions", "builtin", "hardware", "scripts", "detect-hardware.ps1")
	} else {
		scriptPath = filepath.Join("internal", "extensions", "builtin", "hardware", "scripts", "detect-hardware.sh")
	}

	return &Recommender{
		name:         name,
		version:      version,
		dataFilePath: dataFilePath,
		scriptPath:   scriptPath,
	}
}

// Name returns the name of the extension
func (r *Recommender) Name() string {
	return r.name
}

// Version returns the version of the extension
func (r *Recommender) Version() string {
	return r.version
}

// Initialize initializes the recommender and loads the recommendations data
func (r *Recommender) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Override data file path if provided in config
	if dataPath, ok := config["data_file_path"].(string); ok && dataPath != "" {
		r.dataFilePath = dataPath
	}

	// Override script path if provided in config
	if scriptPath, ok := config["script_path"].(string); ok && scriptPath != "" {
		r.scriptPath = scriptPath
	}

	// Load recommendations data
	if err := r.loadRecommendations(); err != nil {
		return fmt.Errorf("failed to load recommendations data: %w", err)
	}

	log.Info().
		Str("extension", r.name).
		Str("data_file", r.dataFilePath).
		Msg("Hardware recommender initialized")

	return nil
}

// Shutdown shuts down the recommender
func (r *Recommender) Shutdown(ctx context.Context) error {
	return nil
}

// Execute executes the hardware recommendation tool
func (r *Recommender) Execute(ctx context.Context, toolName string, params map[string]interface{}) (*core.ToolResult, error) {
	if toolName != "hardware_recommend_models" {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unknown tool: %s", toolName),
		}, fmt.Errorf("unknown tool: %s", toolName)
	}

	// Detect hardware
	specs, err := r.detectHardware(ctx)
	if err != nil {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("hardware detection failed: %v", err),
		}, err
	}

	// Classify hardware group
	groupID, err := r.classifyHardwareGroup(specs)
	if err != nil {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("hardware classification failed: %v", err),
		}, err
	}

	// Get recommendations
	recommendations, err := r.getRecommendations(groupID)
	if err != nil {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get recommendations: %v", err),
		}, err
	}

	// Build output
	output := map[string]interface{}{
		"hardware":        specs,
		"hardware_group":   groupID,
		"recommendations": recommendations,
	}

	return &core.ToolResult{
		Success: true,
		Output:  output,
	}, nil
}

// ListTools returns the available tools
func (r *Recommender) ListTools(ctx context.Context) ([]*core.ToolDefinition, error) {
	return []*core.ToolDefinition{
		{
			Name:        "hardware_recommend_models",
			Description: "Detects system hardware (CPU, RAM, GPU, VRAM) and recommends local LLM models based on hardware capabilities",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}, nil
}

// detectHardware executes the OS-specific hardware detection script
func (r *Recommender) detectHardware(ctx context.Context) (*HardwareSpecs, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// PowerShell script
		cmd = exec.CommandContext(ctx, "powershell.exe", "-ExecutionPolicy", "Bypass", "-File", r.scriptPath)
	} else {
		// Bash script
		cmd = exec.CommandContext(ctx, "bash", r.scriptPath)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %w", err)
	}

	// Parse JSON output
	var specs HardwareSpecs
	if err := json.Unmarshal(output, &specs); err != nil {
		return nil, fmt.Errorf("failed to parse script output: %w", err)
	}

	// Validate specs
	if specs.CPUCores <= 0 {
		return nil, fmt.Errorf("invalid CPU cores: %d", specs.CPUCores)
	}
	if specs.TotalRAMGB <= 0 {
		return nil, fmt.Errorf("invalid RAM: %d GB", specs.TotalRAMGB)
	}

	return &specs, nil
}

// classifyHardwareGroup matches hardware specs to a hardware group
func (r *Recommender) classifyHardwareGroup(specs *HardwareSpecs) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.recommendationsData == nil {
		return "", fmt.Errorf("recommendations data not loaded")
	}

	// Primary check: GPU/VRAM detection
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

// getRecommendations returns model recommendations for a hardware group
func (r *Recommender) getRecommendations(groupID string) ([]ModelRecommendation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.recommendationsData == nil {
		return nil, fmt.Errorf("recommendations data not loaded")
	}

	for _, group := range r.recommendationsData.HardwareGroups {
		if group.ID == groupID {
			return group.Models, nil
		}
	}

	return nil, fmt.Errorf("hardware group not found: %s", groupID)
}

// loadRecommendations loads the recommendations data from JSON file
func (r *Recommender) loadRecommendations() error {
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
