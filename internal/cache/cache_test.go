package cache

import (
	"testing"
)

func TestCache_GetSet(t *testing.T) {
	c := New()

	params := CacheKeyParams{
		Model: "llama2",
		Messages: []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}

	// Test cache miss
	_, found := c.Get(params)
	if found {
		t.Error("Expected cache miss, got cache hit")
	}

	// Set cache
	response := []byte(`{"response": "Hello, how can I help?"}`)
	if err := c.Set(params, response); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Test cache hit
	cached, found := c.Get(params)
	if !found {
		t.Error("Expected cache hit, got cache miss")
	}

	if string(cached) != string(response) {
		t.Errorf("Cached response mismatch: got %s, want %s", string(cached), string(response))
	}
}

func TestCache_DifferentModels(t *testing.T) {
	c := New()

	messages := []map[string]interface{}{
		{"role": "user", "content": "Hello"},
	}

	params1 := CacheKeyParams{
		Model:    "llama2",
		Messages: messages,
	}
	params2 := CacheKeyParams{
		Model:    "mistral",
		Messages: messages,
	}

	response1 := []byte(`{"response": "Response 1"}`)
	response2 := []byte(`{"response": "Response 2"}`)

	if err := c.Set(params1, response1); err != nil {
		t.Fatalf("Failed to set cache for model1: %v", err)
	}
	if err := c.Set(params2, response2); err != nil {
		t.Fatalf("Failed to set cache for model2: %v", err)
	}

	// Check model1
	cached1, found1 := c.Get(params1)
	if !found1 || string(cached1) != string(response1) {
		t.Error("Cache miss for model1")
	}

	// Check model2
	cached2, found2 := c.Get(params2)
	if !found2 || string(cached2) != string(response2) {
		t.Error("Cache miss for model2")
	}
}

func TestCache_DifferentMessages(t *testing.T) {
	c := New()

	model := "llama2"
	messages1 := []map[string]interface{}{
		{"role": "user", "content": "Hello"},
	}
	messages2 := []map[string]interface{}{
		{"role": "user", "content": "Goodbye"},
	}

	params1 := CacheKeyParams{
		Model:    model,
		Messages: messages1,
	}
	params2 := CacheKeyParams{
		Model:    model,
		Messages: messages2,
	}

	response1 := []byte(`{"response": "Response 1"}`)
	response2 := []byte(`{"response": "Response 2"}`)

	if err := c.Set(params1, response1); err != nil {
		t.Fatalf("Failed to set cache for messages1: %v", err)
	}
	if err := c.Set(params2, response2); err != nil {
		t.Fatalf("Failed to set cache for messages2: %v", err)
	}

	// Check messages1
	cached1, found1 := c.Get(params1)
	if !found1 || string(cached1) != string(response1) {
		t.Error("Cache miss for messages1")
	}

	// Check messages2
	cached2, found2 := c.Get(params2)
	if !found2 || string(cached2) != string(response2) {
		t.Error("Cache miss for messages2")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New()

	params := CacheKeyParams{
		Model: "llama2",
		Messages: []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}

	response := []byte(`{"response": "Hello"}`)
	if err := c.Set(params, response); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Verify it's cached
	_, found := c.Get(params)
	if !found {
		t.Error("Expected cache hit before clear")
	}

	// Clear cache
	c.Clear()

	// Verify it's gone
	_, found = c.Get(params)
	if found {
		t.Error("Expected cache miss after clear")
	}
}

// TestCache_DifferentParameters verifies that cache keys include all parameters
func TestCache_DifferentParameters(t *testing.T) {
	c := New()

	baseParams := CacheKeyParams{
		Model: "llama2",
		Messages: []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}

	// Test with different temperatures
	params1 := baseParams
	params1.Temperature = 0.7
	params2 := baseParams
	params2.Temperature = 0.9

	response1 := []byte(`{"response": "Response 1"}`)
	response2 := []byte(`{"response": "Response 2"}`)

	if err := c.Set(params1, response1); err != nil {
		t.Fatalf("Failed to set cache for params1: %v", err)
	}
	if err := c.Set(params2, response2); err != nil {
		t.Fatalf("Failed to set cache for params2: %v", err)
	}

	// Verify different temperatures produce different cache keys
	cached1, found1 := c.Get(params1)
	if !found1 || string(cached1) != string(response1) {
		t.Error("Cache miss for params1")
	}

	cached2, found2 := c.Get(params2)
	if !found2 || string(cached2) != string(response2) {
		t.Error("Cache miss for params2")
	}

	// Verify params1 doesn't return params2's response
	if string(cached1) == string(response2) {
		t.Error("Cache incorrectly reused response for different temperature")
	}
}

// TestCache_IdenticalParameters verifies correct cache reuse for identical requests
func TestCache_IdenticalParameters(t *testing.T) {
	c := New()

	params := CacheKeyParams{
		Model:       "llama2",
		Temperature: 0.7,
		MaxTokens:   100,
		Messages: []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		Tools: []map[string]interface{}{
			{"type": "function", "function": map[string]interface{}{"name": "test"}},
		},
	}

	response := []byte(`{"response": "Hello, how can I help?"}`)

	// Set cache
	if err := c.Set(params, response); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Create identical params (should hit cache)
	identicalParams := CacheKeyParams{
		Model:       "llama2",
		Temperature: 0.7,
		MaxTokens:   100,
		Messages: []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		Tools: []map[string]interface{}{
			{"type": "function", "function": map[string]interface{}{"name": "test"}},
		},
	}

	cached, found := c.Get(identicalParams)
	if !found {
		t.Error("Expected cache hit for identical parameters, got cache miss")
	}
	if string(cached) != string(response) {
		t.Errorf("Cached response mismatch: got %s, want %s", string(cached), string(response))
	}
}
