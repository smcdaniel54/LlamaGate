package hardware

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommender_LoadRecommendations(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")

	// Create test data with new fields
	testData := `{
		"version": "2.0.0",
		"last_updated": "2026-01-22",
		"source": "artificialanalysis.ai/models/open-source (verified Ollama availability)",
		"description": "Test data",
		"hardware_groups": [
			{
				"id": "test_group",
				"name": "Test Group",
				"description": "Test",
				"criteria": {
					"requires_gpu": false,
					"min_ram_gb": 16
				},
				"models": [
					{
						"name": "Test Model",
						"ollama_name": "test-model",
						"priority": 1,
						"description": "Test model",
						"intelligence_score": 10.5,
						"parameters_b": 7.0,
						"min_ram_gb": 8,
						"min_vram_gb": 0,
						"quantized": true,
						"ollama_command": "ollama pull test-model",
						"use_cases": ["test"],
						"artificial_analysis_url": "https://artificialanalysis.ai/models/test-model"
					}
				]
			}
		]
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender(dataFile)
	err := r.LoadRecommendations()
	require.NoError(t, err)

	// Verify data was loaded
	r.mu.RLock()
	assert.NotNil(t, r.recommendationsData)
	assert.Equal(t, "2.0.0", r.recommendationsData.Version)
	assert.Equal(t, "artificialanalysis.ai/models/open-source (verified Ollama availability)", r.recommendationsData.Source)
	assert.Equal(t, "Test data", r.recommendationsData.Description)
	assert.Len(t, r.recommendationsData.HardwareGroups, 1)
	
	// Verify model with new fields
	model := r.recommendationsData.HardwareGroups[0].Models[0]
	assert.Equal(t, "Test Model", model.Name)
	assert.NotNil(t, model.IntelligenceScore)
	assert.Equal(t, 10.5, *model.IntelligenceScore)
	assert.NotNil(t, model.ParametersB)
	assert.Equal(t, 7.0, *model.ParametersB)
	assert.Equal(t, "https://artificialanalysis.ai/models/test-model", model.ArtificialAnalysisURL)
	r.mu.RUnlock()
}

func TestRecommender_ClassifyHardwareGroup_CPUOnly(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")

	testData := `{
		"version": "2.0.0",
		"last_updated": "2026-01-22",
		"source": "test",
		"hardware_groups": [
			{
				"id": "cpu_only_32_64gb",
				"name": "CPU-only, 32-64GB RAM",
				"description": "Test",
				"criteria": {
					"requires_gpu": false,
					"min_ram_gb": 32,
					"max_ram_gb": 64
				},
				"models": []
			}
		]
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender(dataFile)
	require.NoError(t, r.LoadRecommendations())

	// Test CPU-only classification
	specs := &Specs{
		CPUCores:    8,
		CPUModel:    "Test CPU",
		TotalRAMGB:  32,
		GPUDetected: false,
		GPUVRAMGB:   0,
	}

	groupID, err := r.ClassifyHardwareGroup(specs)
	require.NoError(t, err)
	assert.Equal(t, "cpu_only_32_64gb", groupID)
}

func TestRecommender_ClassifyHardwareGroup_GPU(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")

	testData := `{
		"version": "2.0.0",
		"last_updated": "2026-01-22",
		"source": "test",
		"hardware_groups": [
			{
				"id": "gpu_8_16gb_vram",
				"name": "GPU with 8-16GB VRAM",
				"description": "Test",
				"criteria": {
					"requires_gpu": true,
					"min_vram_gb": 8,
					"max_vram_gb": 16
				},
				"models": []
			}
		]
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender(dataFile)
	require.NoError(t, r.LoadRecommendations())

	// Test GPU classification
	specs := &Specs{
		CPUCores:       8,
		CPUModel:       "Test CPU",
		TotalRAMGB:     32,
		GPUDetected:    true,
		GPUName:        "Test GPU",
		GPUVRAMGB:      10,
		DetectionMethod: "nvidia-smi",
	}

	groupID, err := r.ClassifyHardwareGroup(specs)
	require.NoError(t, err)
	assert.Equal(t, "gpu_8_16gb_vram", groupID)
}

func TestRecommender_GetRecommendations_MultipleModels(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")

	testData := `{
		"version": "2.0.0",
		"last_updated": "2026-01-22",
		"source": "test",
		"hardware_groups": [
			{
				"id": "test_group",
				"name": "Test Group",
				"description": "Test",
				"criteria": {
					"requires_gpu": false
				},
				"models": [
					{
						"name": "Model A",
						"ollama_name": "model-a",
						"priority": 2,
						"description": "Second choice",
						"intelligence_score": 8.0,
						"parameters_b": 7.0,
						"min_ram_gb": 8,
						"min_vram_gb": 0,
						"quantized": true,
						"ollama_command": "ollama pull model-a",
						"use_cases": ["test"],
						"artificial_analysis_url": "https://artificialanalysis.ai/models/model-a"
					},
					{
						"name": "Model B",
						"ollama_name": "model-b",
						"priority": 1,
						"description": "First choice",
						"intelligence_score": 10.0,
						"parameters_b": 7.0,
						"min_ram_gb": 8,
						"min_vram_gb": 0,
						"quantized": true,
						"ollama_command": "ollama pull model-b",
						"use_cases": ["test"],
						"artificial_analysis_url": "https://artificialanalysis.ai/models/model-b"
					},
					{
						"name": "Model C",
						"ollama_name": "model-c",
						"priority": 3,
						"description": "Third choice",
						"intelligence_score": 7.0,
						"parameters_b": 7.0,
						"min_ram_gb": 8,
						"min_vram_gb": 0,
						"quantized": true,
						"ollama_command": "ollama pull model-c",
						"use_cases": ["test"],
						"artificial_analysis_url": "https://artificialanalysis.ai/models/model-c"
					}
				]
			}
		]
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender(dataFile)
	require.NoError(t, r.LoadRecommendations())

	recommendations, err := r.GetRecommendations("test_group")
	require.NoError(t, err)
	require.Len(t, recommendations, 3)
	
	// Verify models are sorted by priority (1, 2, 3)
	assert.Equal(t, "Model B", recommendations[0].Name, "Priority 1 should be first")
	assert.Equal(t, 1, recommendations[0].Priority)
	assert.Equal(t, "Model A", recommendations[1].Name, "Priority 2 should be second")
	assert.Equal(t, 2, recommendations[1].Priority)
	assert.Equal(t, "Model C", recommendations[2].Name, "Priority 3 should be third")
	assert.Equal(t, 3, recommendations[2].Priority)
}

func TestRecommender_GetRecommendations(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")

	testData := `{
		"version": "2.0.0",
		"last_updated": "2026-01-22",
		"source": "test",
		"hardware_groups": [
			{
				"id": "test_group",
				"name": "Test Group",
				"description": "Test",
				"criteria": {
					"requires_gpu": false
				},
				"models": [
					{
						"name": "Test Model",
						"ollama_name": "test-model",
						"priority": 1,
						"description": "Test model",
						"intelligence_score": 10.5,
						"parameters_b": 7.0,
						"min_ram_gb": 8,
						"min_vram_gb": 0,
						"quantized": true,
						"ollama_command": "ollama pull test-model",
						"use_cases": ["test"],
						"artificial_analysis_url": "https://artificialanalysis.ai/models/test-model"
					}
				]
			}
		]
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender(dataFile)
	require.NoError(t, r.LoadRecommendations())

	recommendations, err := r.GetRecommendations("test_group")
	require.NoError(t, err)
	require.Len(t, recommendations, 1)
	
	model := recommendations[0]
	assert.Equal(t, "Test Model", model.Name)
	assert.Equal(t, "test-model", model.OllamaName)
	assert.NotNil(t, model.IntelligenceScore)
	assert.Equal(t, 10.5, *model.IntelligenceScore)
	assert.NotNil(t, model.ParametersB)
	assert.Equal(t, 7.0, *model.ParametersB)
	assert.Equal(t, "https://artificialanalysis.ai/models/test-model", model.ArtificialAnalysisURL)
}

func TestRecommender_GetRecommendations_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")

	testData := `{
		"version": "2.0.0",
		"last_updated": "2026-01-22",
		"source": "test",
		"hardware_groups": []
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender(dataFile)
	require.NoError(t, r.LoadRecommendations())

	_, err := r.GetRecommendations("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hardware group not found")
}

func TestModelRecommendation_OptionalFields(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")

	// Test with optional fields omitted
	testData := `{
		"version": "2.0.0",
		"last_updated": "2026-01-22",
		"source": "test",
		"hardware_groups": [
			{
				"id": "test_group",
				"name": "Test Group",
				"description": "Test",
				"criteria": {
					"requires_gpu": false
				},
				"models": [
					{
						"name": "Test Model",
						"ollama_name": "test-model",
						"priority": 1,
						"description": "Test model",
						"min_ram_gb": 8,
						"min_vram_gb": 0,
						"quantized": true,
						"ollama_command": "ollama pull test-model",
						"use_cases": ["test"]
					}
				]
			}
		]
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender(dataFile)
	require.NoError(t, r.LoadRecommendations())

	recommendations, err := r.GetRecommendations("test_group")
	require.NoError(t, err)
	require.Len(t, recommendations, 1)
	
	model := recommendations[0]
	assert.Equal(t, "Test Model", model.Name)
	assert.Nil(t, model.IntelligenceScore) // Should be nil when omitted
	assert.Nil(t, model.ParametersB)      // Should be nil when omitted
	assert.Empty(t, model.ArtificialAnalysisURL) // Should be empty when omitted
}
