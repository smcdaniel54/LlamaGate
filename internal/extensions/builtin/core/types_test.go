package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgentRequest(t *testing.T) {
	t.Run("create agent request", func(t *testing.T) {
		temp := 0.7
		maxTokens := 100
		req := &AgentRequest{
			Model:       "test-model",
			Messages:    []Message{{Role: "user", Content: "test"}},
			Temperature: &temp,
			MaxTokens:   &maxTokens,
			Stream:      false,
			Metadata:    map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "test-model", req.Model)
		assert.Len(t, req.Messages, 1)
		assert.Equal(t, 0.7, *req.Temperature)
		assert.Equal(t, 100, *req.MaxTokens)
		assert.False(t, req.Stream)
		assert.Equal(t, "value", req.Metadata["key"])
	})
}

func TestAgentResponse(t *testing.T) {
	t.Run("create agent response", func(t *testing.T) {
		usage := &Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		}

		resp := &AgentResponse{
			ID:        "resp-1",
			Model:     "test-model",
			Content:   "response content",
			ToolCalls: []*ToolCall{{ID: "call-1", Name: "test-tool"}},
			Usage:     usage,
			Metadata:  map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "resp-1", resp.ID)
		assert.Equal(t, "test-model", resp.Model)
		assert.Equal(t, "response content", resp.Content)
		assert.Len(t, resp.ToolCalls, 1)
		assert.Equal(t, 30, resp.Usage.TotalTokens)
	})
}

func TestStreamChunk(t *testing.T) {
	t.Run("create stream chunk", func(t *testing.T) {
		chunk := &StreamChunk{
			Content:  "chunk content",
			Done:     false,
			Metadata: map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "chunk content", chunk.Content)
		assert.False(t, chunk.Done)
		assert.Nil(t, chunk.Error)
	})

	t.Run("create done chunk", func(t *testing.T) {
		chunk := &StreamChunk{
			Content: "",
			Done:    true,
		}

		assert.True(t, chunk.Done)
	})
}

func TestMessage(t *testing.T) {
	t.Run("create message", func(t *testing.T) {
		msg := Message{
			Role:    "user",
			Content: "test message",
		}

		assert.Equal(t, "user", msg.Role)
		assert.Equal(t, "test message", msg.Content)
	})
}

func TestToolDefinition(t *testing.T) {
	t.Run("create tool definition", func(t *testing.T) {
		tool := &ToolDefinition{
			Name:        "test-tool",
			Description: "A test tool",
			Parameters: map[string]interface{}{
				"param1": "string",
			},
			Metadata: map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "test-tool", tool.Name)
		assert.Equal(t, "A test tool", tool.Description)
		assert.Equal(t, "string", tool.Parameters["param1"])
	})
}

func TestToolCall(t *testing.T) {
	t.Run("create tool call", func(t *testing.T) {
		call := &ToolCall{
			ID:       "call-1",
			Name:     "test-tool",
			Arguments: map[string]interface{}{"arg1": "value1"},
		}

		assert.Equal(t, "call-1", call.ID)
		assert.Equal(t, "test-tool", call.Name)
		assert.Equal(t, "value1", call.Arguments["arg1"])
	})
}

func TestToolResult(t *testing.T) {
	t.Run("create successful tool result", func(t *testing.T) {
		result := &ToolResult{
			Success:  true,
			Output:   "result output",
			Duration: 100 * time.Millisecond,
			Metadata: map[string]interface{}{"key": "value"},
		}

		assert.True(t, result.Success)
		assert.Equal(t, "result output", result.Output)
		assert.Equal(t, 100*time.Millisecond, result.Duration)
		assert.Empty(t, result.Error)
	})

	t.Run("create failed tool result", func(t *testing.T) {
		result := &ToolResult{
			Success:  false,
			Error:    "tool execution failed",
			Duration: 50 * time.Millisecond,
		}

		assert.False(t, result.Success)
		assert.Equal(t, "tool execution failed", result.Error)
	})
}

func TestUsage(t *testing.T) {
	t.Run("create usage", func(t *testing.T) {
		usage := &Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		}

		assert.Equal(t, 10, usage.PromptTokens)
		assert.Equal(t, 20, usage.CompletionTokens)
		assert.Equal(t, 30, usage.TotalTokens)
		assert.Equal(t, 30, usage.PromptTokens+usage.CompletionTokens)
	})
}

func TestCondition(t *testing.T) {
	t.Run("create condition", func(t *testing.T) {
		cond := &Condition{
			Type:       "llm",
			Expression: "test > 5",
			Context:    map[string]interface{}{"test": 10},
			Metadata:   map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "llm", cond.Type)
		assert.Equal(t, "test > 5", cond.Expression)
		assert.Equal(t, 10, cond.Context["test"])
	})
}

func TestDecision(t *testing.T) {
	t.Run("create decision", func(t *testing.T) {
		decision := &Decision{
			Result:     true,
			Confidence: 0.95,
			Reason:     "condition met",
			Metadata:   map[string]interface{}{"key": "value"},
		}

		assert.True(t, decision.Result)
		assert.Equal(t, 0.95, decision.Confidence)
		assert.Equal(t, "condition met", decision.Reason)
	})
}

func TestBranch(t *testing.T) {
	t.Run("create branch", func(t *testing.T) {
		branch := &Branch{
			ID:        "branch-1",
			Name:      "test branch",
			Condition: &Condition{Type: "expression", Expression: "x > 0"},
			Priority:  10,
			Metadata:  map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "branch-1", branch.ID)
		assert.Equal(t, "test branch", branch.Name)
		assert.Equal(t, 10, branch.Priority)
		assert.NotNil(t, branch.Condition)
	})
}

func TestWorkflowState(t *testing.T) {
	t.Run("create workflow state", func(t *testing.T) {
		now := time.Now()
		state := &WorkflowState{
			WorkflowID: "wf-1",
			Status:     "running",
			Step:       "step-1",
			Context:    map[string]interface{}{"key": "value"},
			History: []*StateHistory{
				{
					Timestamp: now,
					Step:      "step-1",
					Action:    "started",
				},
			},
			CreatedAt: now,
			UpdatedAt: now,
			Metadata:  map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "wf-1", state.WorkflowID)
		assert.Equal(t, "running", state.Status)
		assert.Equal(t, "step-1", state.Step)
		assert.Len(t, state.History, 1)
	})
}

func TestStateHistory(t *testing.T) {
	t.Run("create state history", func(t *testing.T) {
		now := time.Now()
		history := &StateHistory{
			Timestamp: now,
			Step:      "step-1",
			Action:    "completed",
			Data:      map[string]interface{}{"result": "success"},
		}

		assert.Equal(t, now, history.Timestamp)
		assert.Equal(t, "step-1", history.Step)
		assert.Equal(t, "completed", history.Action)
		assert.Equal(t, "success", history.Data["result"])
	})
}

func TestEvent(t *testing.T) {
	t.Run("create event", func(t *testing.T) {
		now := time.Now()
		event := &Event{
			ID:        "event-1",
			Type:      "test.event",
			Source:    "test-source",
			Timestamp: now,
			Data:      map[string]interface{}{"key": "value"},
			Metadata:  map[string]interface{}{"meta": "data"},
		}

		assert.Equal(t, "event-1", event.ID)
		assert.Equal(t, "test.event", event.Type)
		assert.Equal(t, "test-source", event.Source)
		assert.Equal(t, now, event.Timestamp)
	})
}

func TestEventFilter(t *testing.T) {
	t.Run("create event filter", func(t *testing.T) {
		filter := &EventFilter{
			Types:   []string{"type1", "type2"},
			Sources: []string{"source1"},
			Match:   map[string]interface{}{"key": "value"},
		}

		assert.Len(t, filter.Types, 2)
		assert.Len(t, filter.Sources, 1)
		assert.Equal(t, "value", filter.Match["key"])
	})
}

func TestValidationRules(t *testing.T) {
	t.Run("create validation rules", func(t *testing.T) {
		rules := &ValidationRules{
			Type:     "schema",
			Schema:   map[string]interface{}{"type": "object"},
			Rules:    []string{"rule1", "rule2"},
			Metadata: map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "schema", rules.Type)
		assert.Equal(t, "object", rules.Schema["type"])
		assert.Len(t, rules.Rules, 2)
	})
}

func TestValidationResult(t *testing.T) {
	t.Run("create valid validation result", func(t *testing.T) {
		result := &ValidationResult{
			Valid:    true,
			Score:    0.95,
			Metadata: map[string]interface{}{"key": "value"},
		}

		assert.True(t, result.Valid)
		assert.Equal(t, 0.95, result.Score)
		assert.Empty(t, result.Errors)
	})

	t.Run("create invalid validation result", func(t *testing.T) {
		result := &ValidationResult{
			Valid:   false,
			Errors:  []string{"error1", "error2"},
			Warnings: []string{"warning1"},
		}

		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 2)
		assert.Len(t, result.Warnings, 1)
	})
}

func TestApprovalRequest(t *testing.T) {
	t.Run("create approval request", func(t *testing.T) {
		timeout := 5 * time.Minute
		req := &ApprovalRequest{
			RequestID:   "req-1",
			WorkflowID:  "wf-1",
			Title:       "Test Approval",
			Description: "Please approve this",
			Data:        map[string]interface{}{"key": "value"},
			Options:     []string{"approve", "reject"},
			Timeout:     &timeout,
			Metadata:    map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "req-1", req.RequestID)
		assert.Equal(t, "Test Approval", req.Title)
		assert.Len(t, req.Options, 2)
		assert.Equal(t, 5*time.Minute, *req.Timeout)
	})
}

func TestApprovalResponse(t *testing.T) {
	t.Run("create approval response", func(t *testing.T) {
		now := time.Now()
		resp := &ApprovalResponse{
			RequestID: "req-1",
			Approved:  true,
			Choice:    "approve",
			Comment:   "Looks good",
			Timestamp: now,
			Metadata:  map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "req-1", resp.RequestID)
		assert.True(t, resp.Approved)
		assert.Equal(t, "approve", resp.Choice)
		assert.Equal(t, "Looks good", resp.Comment)
	})
}

func TestInputPrompt(t *testing.T) {
	t.Run("create input prompt", func(t *testing.T) {
		timeout := 2 * time.Minute
		prompt := &InputPrompt{
			PromptID:   "prompt-1",
			WorkflowID: "wf-1",
			Type:       "text",
			Prompt:     "Enter your name",
			Options:    []string{"option1", "option2"},
			Required:   true,
			Timeout:    &timeout,
			Metadata:   map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "prompt-1", prompt.PromptID)
		assert.Equal(t, "text", prompt.Type)
		assert.Equal(t, "Enter your name", prompt.Prompt)
		assert.True(t, prompt.Required)
	})
}

func TestInputResponse(t *testing.T) {
	t.Run("create input response", func(t *testing.T) {
		now := time.Now()
		resp := &InputResponse{
			PromptID:  "prompt-1",
			Value:     "user input",
			Timestamp: now,
			Metadata:  map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "prompt-1", resp.PromptID)
		assert.Equal(t, "user input", resp.Value)
		assert.Equal(t, now, resp.Timestamp)
	})
}

func TestTransformation(t *testing.T) {
	t.Run("create transformation", func(t *testing.T) {
		trans := &Transformation{
			Type:     "map",
			Config:   map[string]interface{}{"function": "uppercase"},
			Metadata: map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "map", trans.Type)
		assert.Equal(t, "uppercase", trans.Config["function"])
	})
}
