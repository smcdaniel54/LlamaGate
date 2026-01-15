package extensions

import "context"

// LLMHandlerFunc is a function type for making LLM calls
// Returns the LLM response as a map, or an error
type LLMHandlerFunc func(ctx context.Context, model string, messages []map[string]interface{}, options map[string]interface{}) (map[string]interface{}, error)
