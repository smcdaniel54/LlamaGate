//go:build windows

package ollama

import (
	"os"
	"os/exec"
	"syscall"
)

// startOllama starts the Ollama server process (Windows: hide console window).
func startOllama() (*exec.Cmd, error) {
	cmd := exec.Command("ollama", "serve")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
