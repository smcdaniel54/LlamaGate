package proxy

import (
	"fmt"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidateChatRequest validates a chat completion request
func ValidateChatRequest(req *ChatCompletionRequest) error {
	if req.Model == "" {
		return &ValidationError{
			Field:   "model",
			Message: "Model is required",
		}
	}

	if len(req.Messages) == 0 {
		return &ValidationError{
			Field:   "messages",
			Message: "Messages are required",
		}
	}

	return nil
}
