package cache

import (
	"testing"
)

func TestCache_GetSet(t *testing.T) {
	c := New()

	model := "llama2"
	messages := []map[string]interface{}{
		{"role": "user", "content": "Hello"},
	}

	// Test cache miss
	_, found := c.Get(model, messages)
	if found {
		t.Error("Expected cache miss, got cache hit")
	}

	// Set cache
	response := []byte(`{"response": "Hello, how can I help?"}`)
	if err := c.Set(model, messages, response); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Test cache hit
	cached, found := c.Get(model, messages)
	if !found {
		t.Error("Expected cache hit, got cache miss")
	}

	if string(cached) != string(response) {
		t.Errorf("Cached response mismatch: got %s, want %s", string(cached), string(response))
	}
}

func TestCache_DifferentModels(t *testing.T) {
	c := New()

	model1 := "llama2"
	model2 := "mistral"
	messages := []map[string]interface{}{
		{"role": "user", "content": "Hello"},
	}

	response1 := []byte(`{"response": "Response 1"}`)
	response2 := []byte(`{"response": "Response 2"}`)

	c.Set(model1, messages, response1)
	c.Set(model2, messages, response2)

	// Check model1
	cached1, found1 := c.Get(model1, messages)
	if !found1 || string(cached1) != string(response1) {
		t.Error("Cache miss for model1")
	}

	// Check model2
	cached2, found2 := c.Get(model2, messages)
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

	response1 := []byte(`{"response": "Response 1"}`)
	response2 := []byte(`{"response": "Response 2"}`)

	c.Set(model, messages1, response1)
	c.Set(model, messages2, response2)

	// Check messages1
	cached1, found1 := c.Get(model, messages1)
	if !found1 || string(cached1) != string(response1) {
		t.Error("Cache miss for messages1")
	}

	// Check messages2
	cached2, found2 := c.Get(model, messages2)
	if !found2 || string(cached2) != string(response2) {
		t.Error("Cache miss for messages2")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New()

	model := "llama2"
	messages := []map[string]interface{}{
		{"role": "user", "content": "Hello"},
	}

	response := []byte(`{"response": "Hello"}`)
	c.Set(model, messages, response)

	// Verify it's cached
	_, found := c.Get(model, messages)
	if !found {
		t.Error("Expected cache hit before clear")
	}

	// Clear cache
	c.Clear()

	// Verify it's gone
	_, found = c.Get(model, messages)
	if found {
		t.Error("Expected cache miss after clear")
	}
}

