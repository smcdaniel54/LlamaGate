//go:build !windows

package ollama

import (
	"os"
	"os/exec"
)

// startOllama starts the Ollama server process (Unix: no special attributes).
func startOllama() (*exec.Cmd, error) {
	cmd := exec.Command("ollama", "serve")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
