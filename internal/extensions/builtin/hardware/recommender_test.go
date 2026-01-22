package hardware

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommender_Name(t *testing.T) {
	r := NewRecommender("test-recommender", "1.0.0", "test.json")
	assert.Equal(t, "test-recommender", r.Name())
}

func TestRecommender_Version(t *testing.T) {
	r := NewRecommender("test-recommender", "1.0.0", "test.json")
	assert.Equal(t, "1.0.0", r.Version())
}

func TestRecommender_Initialize(t *testing.T) {
	// Create a temporary data file
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")
	
	// Create minimal valid JSON
	testData := `{
		"version": "1.0.0",
		"last_updated": "2026-01-22",
		"source": "test",
		"hardware_groups": [
			{
				"id": "test_group",
				"name": "Test Group",
				"description": "Test",
				"criteria": {
					"requires_gpu": false,
					"min_ram_gb": 16
				},
				"models": []
			}
		]
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender("test-recommender", "1.0.0", dataFile)
	err := r.Initialize(context.Background(), map[string]interface{}{
		"data_file_path": dataFile,
	})
	require.NoError(t, err)
}

func TestRecommender_ListTools(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")
	
	testData := `{
		"version": "1.0.0",
		"last_updated": "2026-01-22",
		"source": "test",
		"hardware_groups": []
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender("test-recommender", "1.0.0", dataFile)
	require.NoError(t, r.Initialize(context.Background(), nil))

	tools, err := r.ListTools(context.Background())
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, "hardware_recommend_models", tools[0].Name)
}

func TestRecommender_ClassifyHardwareGroup(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")
	
	testData := `{
		"version": "1.0.0",
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
			},
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

	r := NewRecommender("test-recommender", "1.0.0", dataFile)
	require.NoError(t, r.Initialize(context.Background(), nil))

	// Test CPU-only classification
	specs := &HardwareSpecs{
		CPUCores:    8,
		CPUModel:    "Test CPU",
		TotalRAMGB:  32,
		GPUDetected: false,
		GPUVRAMGB:   0,
	}
	
	groupID, err := r.classifyHardwareGroup(specs)
	require.NoError(t, err)
	assert.Equal(t, "cpu_only_32_64gb", groupID)

	// Test GPU classification
	specsGPU := &HardwareSpecs{
		CPUCores:    8,
		CPUModel:    "Test CPU",
		TotalRAMGB:  32,
		GPUDetected: true,
		GPUName:     "Test GPU",
		GPUVRAMGB:   10,
	}
	
	groupID, err = r.classifyHardwareGroup(specsGPU)
	require.NoError(t, err)
	assert.Equal(t, "gpu_8_16gb_vram", groupID)
}

func TestRecommender_GetRecommendations(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")
	
	testData := `{
		"version": "1.0.0",
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

	r := NewRecommender("test-recommender", "1.0.0", dataFile)
	require.NoError(t, r.Initialize(context.Background(), nil))

	recommendations, err := r.getRecommendations("test_group")
	require.NoError(t, err)
	require.Len(t, recommendations, 1)
	assert.Equal(t, "Test Model", recommendations[0].Name)
	assert.Equal(t, "test-model", recommendations[0].OllamaName)
}

func TestRecommender_Execute_UnknownTool(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")
	
	testData := `{
		"version": "1.0.0",
		"last_updated": "2026-01-22",
		"source": "test",
		"hardware_groups": []
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender("test-recommender", "1.0.0", dataFile)
	require.NoError(t, r.Initialize(context.Background(), nil))

	result, err := r.Execute(context.Background(), "unknown_tool", nil)
	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "unknown tool")
}

func TestHardwareSpecs_JSON(t *testing.T) {
	specs := &HardwareSpecs{
		CPUCores:       8,
		CPUModel:       "Test CPU",
		TotalRAMGB:     32,
		GPUDetected:    true,
		GPUName:        "Test GPU",
		GPUVRAMGB:      10,
		DetectionMethod: "nvidia-smi",
	}

	data, err := json.Marshal(specs)
	require.NoError(t, err)

	var unmarshaled HardwareSpecs
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	
	assert.Equal(t, specs.CPUCores, unmarshaled.CPUCores)
	assert.Equal(t, specs.CPUModel, unmarshaled.CPUModel)
	assert.Equal(t, specs.TotalRAMGB, unmarshaled.TotalRAMGB)
	assert.Equal(t, specs.GPUDetected, unmarshaled.GPUDetected)
	assert.Equal(t, specs.GPUName, unmarshaled.GPUName)
	assert.Equal(t, specs.GPUVRAMGB, unmarshaled.GPUVRAMGB)
}

func TestLoadRecommendations(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test-recommendations.json")
	
	testData := `{
		"version": "1.0.0",
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
				"models": []
			}
		]
	}`
	require.NoError(t, os.WriteFile(dataFile, []byte(testData), 0644))

	r := NewRecommender("test-recommender", "1.0.0", dataFile)
	err := r.loadRecommendations()
	require.NoError(t, err)
	
	// Verify data was loaded
	r.mu.RLock()
	assert.NotNil(t, r.recommendationsData)
	assert.Equal(t, "1.0.0", r.recommendationsData.Version)
	assert.Len(t, r.recommendationsData.HardwareGroups, 1)
	r.mu.RUnlock()
}
