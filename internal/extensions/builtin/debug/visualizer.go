package debug

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

// Visualizer provides visual debugging output for workflows
type Visualizer struct {
	name     string
	version  string
	enabled  bool
	registry *core.Registry
	workflows map[string]*WorkflowVisualization
}

// WorkflowVisualization tracks workflow state for visualization
type WorkflowVisualization struct {
	WorkflowID string
	Steps      []*StepVisualization
	StartTime  time.Time
	Status     string
}

// StepVisualization represents a step in a workflow
type StepVisualization struct {
	Name      string
	Status    string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Data      map[string]interface{}
}

// NewVisualizer creates a new visualizer
func NewVisualizer(name, version string) *Visualizer {
	return &Visualizer{
		name:      name,
		version:   version,
		enabled:   true,
		registry:  core.GetRegistry(),
		workflows: make(map[string]*WorkflowVisualization),
	}
}

// Name returns the name of the extension
func (v *Visualizer) Name() string {
	return v.name
}

// Version returns the version of the extension
func (v *Visualizer) Version() string {
	return v.version
}

// Initialize initializes the visualizer
func (v *Visualizer) Initialize(ctx context.Context, config map[string]interface{}) error {
	if enabled, ok := config["enabled"].(bool); ok {
		v.enabled = enabled
	}

	if !v.enabled {
		return nil
	}

	// Subscribe to workflow events
	publisher, err := v.registry.GetEventPublisher("default")
	if err == nil {
		filter := &core.EventFilter{
			Types: []string{
				"workflow.started",
				"workflow.step.started",
				"workflow.step.completed",
				"workflow.completed",
				"workflow.failed",
				"state.saved",
				"state.status_changed",
				"agent.call.started",
				"agent.call.completed",
				"tool.execution.started",
				"tool.execution.completed",
			},
		}
		_, _ = publisher.Subscribe(ctx, filter, v.handleEvent)
	}

	v.printHeader("Visual Debugger Initialized")
	return nil
}

// Shutdown shuts down the visualizer
func (v *Visualizer) Shutdown(ctx context.Context) error {
	if v.enabled {
		v.printHeader("Visual Debugger Shutting Down")
		v.printSummary()
	}
	return nil
}

// handleEvent handles events for visualization
func (v *Visualizer) handleEvent(ctx context.Context, event *core.Event) error {
	if !v.enabled || event == nil {
		return nil
	}

	switch event.Type {
	case "workflow.started":
		v.handleWorkflowStarted(event)
	case "workflow.step.started":
		v.handleStepStarted(event)
	case "workflow.step.completed":
		v.handleStepCompleted(event)
	case "workflow.completed", "workflow.failed":
		v.handleWorkflowCompleted(event)
	case "state.saved":
		v.handleStateSaved(event)
	case "state.status_changed":
		v.handleStatusChanged(event)
	case "agent.call.started":
		v.handleAgentCallStarted(event)
	case "agent.call.completed":
		v.handleAgentCallCompleted(event)
	case "tool.execution.started":
		v.handleToolStarted(event)
	case "tool.execution.completed":
		v.handleToolCompleted(event)
	}

	return nil
}

func (v *Visualizer) handleWorkflowStarted(event *core.Event) {
	workflowID, _ := event.Data["workflow_id"].(string)
	if workflowID == "" {
		return
	}

	v.workflows[workflowID] = &WorkflowVisualization{
		WorkflowID: workflowID,
		Steps:      []*StepVisualization{},
		StartTime:  event.Timestamp,
		Status:     "running",
	}

	v.printBox(fmt.Sprintf("🚀 Workflow Started: %s", workflowID), "green")
}

func (v *Visualizer) handleStepStarted(event *core.Event) {
	workflowID, _ := event.Data["workflow_id"].(string)
	stepName, _ := event.Data["step"].(string)
	if workflowID == "" || stepName == "" {
		return
	}

	wf, exists := v.workflows[workflowID]
	if !exists {
		return
	}

	step := &StepVisualization{
		Name:      stepName,
		Status:    "running",
		StartTime: event.Timestamp,
		Data:      event.Data,
	}
	wf.Steps = append(wf.Steps, step)

	v.printIndent(fmt.Sprintf("  ⏳ Step: %s", stepName), "yellow")
}

func (v *Visualizer) handleStepCompleted(event *core.Event) {
	workflowID, _ := event.Data["workflow_id"].(string)
	stepName, _ := event.Data["step"].(string)
	if workflowID == "" || stepName == "" {
		return
	}

	wf, exists := v.workflows[workflowID]
	if !exists {
		return
	}

	// Find step
	for _, step := range wf.Steps {
		if step.Name == stepName && step.Status == "running" {
			step.Status = "completed"
			step.EndTime = event.Timestamp
			step.Duration = step.EndTime.Sub(step.StartTime)
			v.printIndent(fmt.Sprintf("  ✅ Step: %s (%.2fs)", stepName, step.Duration.Seconds()), "green")
			break
		}
	}
}

func (v *Visualizer) handleWorkflowCompleted(event *core.Event) {
	workflowID, _ := event.Data["workflow_id"].(string)
	if workflowID == "" {
		return
	}

	wf, exists := v.workflows[workflowID]
	if !exists {
		return
	}

	wf.Status = "completed"
	if event.Type == "workflow.failed" {
		wf.Status = "failed"
	}

	duration := time.Since(wf.StartTime)
	icon := "✅"
	color := "green"
	if wf.Status == "failed" {
		icon = "❌"
		color = "red"
	}

	v.printBox(fmt.Sprintf("%s Workflow %s: %s (%.2fs)", icon, wf.Status, workflowID, duration.Seconds()), color)
	v.printWorkflowDetails(wf)
}

func (v *Visualizer) handleStateSaved(event *core.Event) {
	workflowID, _ := event.Data["workflow_id"].(string)
	status, _ := event.Data["status"].(string)
	if workflowID == "" {
		return
	}

	v.printIndent(fmt.Sprintf("  💾 State Saved: %s (status: %s)", workflowID, status), "blue")
}

func (v *Visualizer) handleStatusChanged(event *core.Event) {
	workflowID, _ := event.Data["workflow_id"].(string)
	status, _ := event.Data["status"].(string)
	if workflowID == "" {
		return
	}

	icon := getStatusIcon(status)
	v.printIndent(fmt.Sprintf("  %s Status Changed: %s → %s", icon, workflowID, status), "cyan")
}

func (v *Visualizer) handleAgentCallStarted(event *core.Event) {
	model, _ := event.Data["model"].(string)
	v.printIndent(fmt.Sprintf("  🤖 Agent Call Started: %s", model), "magenta")
}

func (v *Visualizer) handleAgentCallCompleted(event *core.Event) {
	responseID, _ := event.Data["response_id"].(string)
	v.printIndent(fmt.Sprintf("  🤖 Agent Call Completed: %s", responseID), "magenta")
}

func (v *Visualizer) handleToolStarted(event *core.Event) {
	tool, _ := event.Data["tool"].(string)
	v.printIndent(fmt.Sprintf("  🔧 Tool Started: %s", tool), "blue")
}

func (v *Visualizer) handleToolCompleted(event *core.Event) {
	tool, _ := event.Data["tool"].(string)
	duration, _ := event.Data["duration"].(string)
	v.printIndent(fmt.Sprintf("  🔧 Tool Completed: %s (%s)", tool, duration), "blue")
}

// printHeader prints a header
func (v *Visualizer) printHeader(message string) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  %s\n", message)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
}

// printBox prints a boxed message
func (v *Visualizer) printBox(message, color string) {
	width := 60
	padding := (width - len(message) - 2) / 2
	leftPad := padding
	rightPad := width - len(message) - 2 - leftPad

	colorCode := getColorCode(color)
	reset := "\033[0m"

	fmt.Printf("%s┌%s┐%s\n", colorCode, strings.Repeat("─", width-2), reset)
	fmt.Printf("%s│%s%s%s%s│%s\n", colorCode, strings.Repeat(" ", leftPad), message, strings.Repeat(" ", rightPad), reset)
	fmt.Printf("%s└%s┘%s\n", colorCode, strings.Repeat("─", width-2), reset)
}

// printIndent prints an indented message
func (v *Visualizer) printIndent(message, color string) {
	colorCode := getColorCode(color)
	reset := "\033[0m"
	fmt.Printf("%s%s%s\n", colorCode, message, reset)
}

// printWorkflowDetails prints workflow details
func (v *Visualizer) printWorkflowDetails(wf *WorkflowVisualization) {
	fmt.Println()
	fmt.Println("  Workflow Steps:")
	for i, step := range wf.Steps {
		statusIcon := "✅"
		if step.Status == "running" {
			statusIcon = "⏳"
		}
		fmt.Printf("    %d. %s %s", i+1, statusIcon, step.Name)
		if step.Duration > 0 {
			fmt.Printf(" (%.2fs)", step.Duration.Seconds())
		}
		fmt.Println()
	}
	fmt.Println()
}

// printSummary prints a summary of all workflows
func (v *Visualizer) printSummary() {
	if len(v.workflows) == 0 {
		return
	}

	v.printHeader("Workflow Summary")
	for _, wf := range v.workflows {
		duration := time.Since(wf.StartTime)
		fmt.Printf("  %s: %s (%.2fs, %d steps)\n", wf.WorkflowID, wf.Status, duration.Seconds(), len(wf.Steps))
	}
	fmt.Println()
}

// getColorCode returns ANSI color code
func getColorCode(color string) string {
	colors := map[string]string{
		"black":   "\033[30m",
		"red":     "\033[31m",
		"green":   "\033[32m",
		"yellow":  "\033[33m",
		"blue":    "\033[34m",
		"magenta": "\033[35m",
		"cyan":    "\033[36m",
		"white":   "\033[37m",
	}
	return colors[color]
}

// getStatusIcon returns icon for status
func getStatusIcon(status string) string {
	icons := map[string]string{
		"running":  "▶️",
		"paused":   "⏸️",
		"completed": "✅",
		"failed":   "❌",
	}
	if icon, ok := icons[status]; ok {
		return icon
	}
	return "⚪"
}
