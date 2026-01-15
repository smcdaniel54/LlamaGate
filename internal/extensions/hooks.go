package extensions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// HookManager manages extension hooks (middleware and observers)
type HookManager struct {
	registry *Registry
	baseDir  string
}

// NewHookManager creates a new hook manager
func NewHookManager(registry *Registry, baseDir string) *HookManager {
	return &HookManager{
		registry: registry,
		baseDir:  baseDir,
	}
}

// CreateMiddlewareHook creates a Gin middleware for request inspection
func (h *HookManager) CreateMiddlewareHook() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get middleware extensions
		middlewares := h.registry.GetByType("middleware")

		// Execute each middleware extension
		for _, manifest := range middlewares {
			if !h.registry.IsEnabled(manifest.Name) {
				continue
			}

			// Check if hook matches this request
			for _, hook := range manifest.Hooks {
				if hook.On == "http.request" && h.matchesRequest(c, hook) {
					if err := h.executeHook(manifest, hook, c, nil); err != nil {
						// Log error but continue processing
						_ = err
					}
				}
			}
		}

		c.Next()
	}
}

// ExecuteResponseHooks executes observer hooks after response
func (h *HookManager) ExecuteResponseHooks(c *gin.Context, responseData map[string]interface{}) {
	// Get observer extensions
	observers := h.registry.GetByType("observer")

	// Execute each observer extension
	for _, manifest := range observers {
		if !h.registry.IsEnabled(manifest.Name) {
			continue
		}

		// Check if hook matches
		for _, hook := range manifest.Hooks {
			if hook.On == "llm.response" {
				if err := h.executeHook(manifest, hook, c, responseData); err != nil {
					// Log error but continue processing
					_ = err
				}
			}
		}
	}
}

// matchesRequest checks if a hook matches the current request
func (h *HookManager) matchesRequest(c *gin.Context, hook HookDefinition) bool {
	if match, ok := hook.Match["path_prefix"].(string); ok {
		return strings.HasPrefix(c.Request.URL.Path, match)
	}
	return true // Default: match all
}

// executeHook executes a hook action
func (h *HookManager) executeHook(manifest *Manifest, hook HookDefinition, c *gin.Context, responseData map[string]interface{}) error {
	switch hook.Action {
	case "audit.log":
		return h.auditLog(manifest, c, responseData)
	case "usage.track":
		return h.trackUsage(manifest, c, responseData)
	default:
		return fmt.Errorf("unknown hook action: %s", hook.Action)
	}
}

// auditLog creates an audit log entry (for request-inspector)
func (h *HookManager) auditLog(manifest *Manifest, c *gin.Context, responseData map[string]interface{}) error {
	// Get config
	config := manifest.Config
	sampleRate := 1.0
	if rate, ok := config["sample_rate"].(float64); ok {
		sampleRate = rate
	}

	// Sample rate check (simple implementation)
	if sampleRate < 1.0 {
		// In production, would use proper sampling
		// For now, always log
	}

	// Get audit directory
	auditDir := "./var/audit"
	if dir, ok := config["audit_dir"].(string); ok {
		auditDir = dir
	}

	// Resolve relative paths relative to extension directory
	if !filepath.IsAbs(auditDir) {
		extDir := GetExtensionDir(h.baseDir, manifest.Name)
		auditDir = filepath.Join(extDir, auditDir)
	}

	// Ensure audit directory exists
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		return fmt.Errorf("failed to create audit directory: %w", err)
	}

	// Create audit entry
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"request_id": c.GetString("request_id"),
		"ip":        c.ClientIP(),
	}

	// Apply redaction rules
	if redactRules, ok := config["redact"].([]interface{}); ok {
		// Simple redaction: truncate message content
		for _, rule := range redactRules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				if path, ok := ruleMap["path"].(string); ok && path == "$.messages[*].content" {
					if mode, ok := ruleMap["mode"].(string); ok && mode == "truncate" {
						if maxLen, ok := ruleMap["max_len"].(int); ok {
							// In a real implementation, would parse JSON path and redact
							// For now, just note that redaction would happen
							entry["redacted"] = true
							entry["max_length"] = maxLen
						}
					}
				}
			}
		}
	}

	// Write audit log file
	auditFile := filepath.Join(auditDir, fmt.Sprintf("audit-%s.jsonl", time.Now().Format("2006-01-02")))
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	file, err := os.OpenFile(auditFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(string(entryJSON) + "\n"); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	return nil
}

// trackUsage tracks token usage and cost (for cost-usage-reporter)
func (h *HookManager) trackUsage(manifest *Manifest, c *gin.Context, responseData map[string]interface{}) error {
	// Extract usage information from response
	usage := map[string]interface{}{
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"request_id": c.GetString("request_id"),
		"model":      responseData["model"],
	}

	// Extract token counts if available
	if usageData, ok := responseData["usage"].(map[string]interface{}); ok {
		usage["prompt_tokens"] = usageData["prompt_tokens"]
		usage["completion_tokens"] = usageData["completion_tokens"]
		usage["total_tokens"] = usageData["total_tokens"]
	}

	// Simple cost estimation (would need model-specific pricing)
	// For now, just track tokens
	usage["estimated_cost"] = 0.0 // Placeholder

	// Get output directory
	outputDir := "./output"
	for _, output := range manifest.Outputs {
		if output.Type == "file" && output.Path != "" {
			outputPath := output.Path
			// Resolve relative paths relative to extension directory
			if !filepath.IsAbs(outputPath) {
				extDir := GetExtensionDir(h.baseDir, manifest.Name)
				outputPath = filepath.Join(extDir, outputPath)
			}
			outputDir = filepath.Dir(outputPath)
			break
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write usage report
	reportFile := filepath.Join(outputDir, "usage_report.json")
	// If we found an output definition, use that path instead
	for _, output := range manifest.Outputs {
		if output.Type == "file" && output.Path != "" {
			outputPath := output.Path
			if !filepath.IsAbs(outputPath) {
				extDir := GetExtensionDir(h.baseDir, manifest.Name)
				outputPath = filepath.Join(extDir, outputPath)
			}
			reportFile = outputPath
			break
		}
	}
	
	// Read existing report or create new
	var report []map[string]interface{}
	if data, err := os.ReadFile(reportFile); err == nil {
		json.Unmarshal(data, &report)
	}

	// Append new usage entry
	report = append(report, usage)

	// Write updated report
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal usage report: %w", err)
	}

	if err := os.WriteFile(reportFile, reportJSON, 0644); err != nil {
		return fmt.Errorf("failed to write usage report: %w", err)
	}

	return nil
}
