// Package ollama provides utilities for managing Ollama lifecycle.
package ollama

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// DefaultOllamaURL is the default Ollama API URL
	DefaultOllamaURL = "http://localhost:11434"
	// OllamaStartupTimeout is the maximum time to wait for Ollama to start
	OllamaStartupTimeout = 30 * time.Second
	// OllamaCheckInterval is the interval between Ollama health checks
	OllamaCheckInterval = 1 * time.Second
)

// EnsureRunning checks if Ollama is running and starts it if not.
// Returns true if Ollama is running (either already running or successfully started).
func EnsureRunning(ollamaURL string) (bool, error) {
	// Check if Ollama is already running
	if isRunning(ollamaURL) {
		log.Info().Str("ollama_url", ollamaURL).Msg("Ollama is already running")
		return true, nil
	}

	log.Info().Str("ollama_url", ollamaURL).Msg("Ollama is not running - will start it")

	// Check if ollama command exists
	if err := checkOllamaCommand(); err != nil {
		return false, fmt.Errorf("ollama command not found: %w. Please install Ollama from https://ollama.ai", err)
	}

	// Start Ollama
	cmd, err := startOllama()
	if err != nil {
		return false, fmt.Errorf("failed to start Ollama: %w", err)
	}

	// Wait for Ollama to be ready
	ctx, cancel := context.WithTimeout(context.Background(), OllamaStartupTimeout)
	defer cancel()

	ready := waitForOllama(ctx, ollamaURL)
	if !ready {
		// Cleanup: try to kill the process we started
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return false, fmt.Errorf("Ollama failed to start within %v. Please start Ollama manually: ollama serve", OllamaStartupTimeout)
	}

	log.Info().
		Int("pid", cmd.Process.Pid).
		Str("ollama_url", ollamaURL).
		Msg("Ollama started successfully")
	return true, nil
}

// isRunning checks if Ollama is responding at the given URL.
func isRunning(ollamaURL string) bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	url := fmt.Sprintf("%s/api/tags", ollamaURL)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// checkOllamaCommand verifies that the ollama command is available.
func checkOllamaCommand() error {
	cmd := exec.Command("ollama", "--version")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// startOllama starts the Ollama server process.
func startOllama() (*exec.Cmd, error) {
	cmd := exec.Command("ollama", "serve")

	// Set process attributes based on platform
	if runtime.GOOS == "windows" {
		// On Windows, hide the window
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
	}
	// On Unix/Linux/macOS, no special attributes needed
	// The process will run in background by default when started with cmd.Start()

	// Redirect output to prevent cluttering console
	cmd.Stdout = os.Stderr // Log to stderr so it's visible but not mixed with normal output
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

// waitForOllama waits for Ollama to become ready, checking at regular intervals.
func waitForOllama(ctx context.Context, ollamaURL string) bool {
	ticker := time.NewTicker(OllamaCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if isRunning(ollamaURL) {
				return true
			}
		}
	}
}
