package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateChatRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     ChatCompletionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: ChatCompletionRequest{
				Model: "llama2",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing model",
			req: ChatCompletionRequest{
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
			errMsg:  "Model is required",
		},
		{
			name: "empty messages",
			req: ChatCompletionRequest{
				Model:    "llama2",
				Messages: []Message{},
			},
			wantErr: true,
			errMsg:  "Messages are required",
		},
		{
			name: "nil messages",
			req: ChatCompletionRequest{
				Model: "llama2",
			},
			wantErr: true,
			errMsg:  "Messages are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChatRequest(&tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if valErr, ok := err.(*ValidationError); ok {
					assert.Equal(t, tt.errMsg, valErr.Message)
				} else {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
