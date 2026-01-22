package hardware

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// Specs represents detected hardware specifications
type Specs struct {
	CPUCores        int    `json:"cpu_cores"`
	CPUModel        string  `json:"cpu_model"`
	TotalRAMGB      int     `json:"total_ram_gb"`
	GPUDetected     bool    `json:"gpu_detected"`
	GPUName         string  `json:"gpu_name"`
	GPUVRAMGB       int     `json:"gpu_vram_gb"`
	DetectionMethod string  `json:"detection_method"`
}

// Detector handles hardware detection
type Detector struct {
	// No fields needed - all detection is done via Go libraries and system calls
}

// NewDetector creates a new hardware detector
func NewDetector() *Detector {
	return &Detector{}
}

// Detect detects hardware specifications using Go libraries and system calls
func (d *Detector) Detect(ctx context.Context) (*Specs, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	specs := &Specs{
		GPUDetected:     false,
		GPUName:         "",
		GPUVRAMGB:       0,
		DetectionMethod: "none",
	}

	// CPU Detection using gopsutil
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil || len(cpuInfo) == 0 {
		return nil, fmt.Errorf("failed to detect CPU: %w", err)
	}
	
	// Get logical CPU count
	cpuCount, err := cpu.CountsWithContext(ctx, true) // true = logical cores
	if err != nil {
		cpuCount = len(cpuInfo) // Fallback
	}
	
	specs.CPUCores = cpuCount
	if len(cpuInfo) > 0 {
		specs.CPUModel = cpuInfo[0].ModelName
	}

	// RAM Detection using gopsutil
	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect RAM: %w", err)
	}
	specs.TotalRAMGB = int(vmStat.Total / (1024 * 1024 * 1024)) // Convert bytes to GB

	// GPU Detection - try multiple methods
	d.detectGPU(ctx, specs)

	// Validate specs
	if specs.CPUCores <= 0 {
		return nil, fmt.Errorf("invalid CPU cores: %d", specs.CPUCores)
	}
	if specs.TotalRAMGB <= 0 {
		return nil, fmt.Errorf("invalid RAM: %d GB", specs.TotalRAMGB)
	}

	return specs, nil
}

// detectGPU attempts to detect GPU and VRAM using various methods
func (d *Detector) detectGPU(ctx context.Context, specs *Specs) {
	// Method 1: Try nvidia-smi first (most accurate for NVIDIA GPUs)
	if d.detectNVIDIAGPU(ctx, specs) {
		return
	}

	// Method 2: Platform-specific detection
	switch runtime.GOOS {
	case "windows":
		d.detectGPUWindows(ctx, specs)
	case "linux":
		d.detectGPULinux(ctx, specs)
	case "darwin":
		d.detectGPUMacOS(ctx, specs)
	}
}

// detectNVIDIAGPU tries to detect NVIDIA GPU using nvidia-smi
func (d *Detector) detectNVIDIAGPU(ctx context.Context, specs *Specs) bool {
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return false
	}

	// Parse first GPU
	parts := strings.Split(lines[0], ",")
	if len(parts) >= 2 {
		specs.GPUDetected = true
		specs.GPUName = strings.Trim(strings.TrimSpace(parts[0]), "\"")
		
		var vramMB int
		if _, err := fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &vramMB); err == nil {
			specs.GPUVRAMGB = vramMB / 1024
		}
		specs.DetectionMethod = "nvidia-smi"
		return true
	}

	return false
}

// detectGPUWindows detects GPU on Windows using WMI (via PowerShell as fallback)
func (d *Detector) detectGPUWindows(ctx context.Context, specs *Specs) {
	// Try PowerShell to query WMI
	psScript := `Get-CimInstance Win32_VideoController | Where-Object { $_.AdapterRAM -and $_.AdapterRAM -gt 0 } | Select-Object -First 1 | ConvertTo-Json -Compress`
	cmd := exec.CommandContext(ctx, "powershell.exe", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	// Simple JSON parsing (we only need Name and AdapterRAM)
	outputStr := string(output)
	if strings.Contains(outputStr, "Name") {
		specs.GPUDetected = true
		
		// Extract GPU name
		if nameIdx := strings.Index(outputStr, `"Name"`); nameIdx > 0 {
			nameStart := strings.Index(outputStr[nameIdx:], ":")
			if nameStart > 0 {
				namePart := outputStr[nameIdx+nameStart+1:]
				namePart = strings.Trim(namePart, `" ,}`)
				if nameEnd := strings.Index(namePart, `"`); nameEnd > 0 {
					specs.GPUName = namePart[:nameEnd]
				} else {
					specs.GPUName = strings.TrimSpace(namePart)
				}
			}
		}

		// Extract AdapterRAM (in bytes)
		if ramIdx := strings.Index(outputStr, `"AdapterRAM"`); ramIdx > 0 {
			ramStart := strings.Index(outputStr[ramIdx:], ":")
			if ramStart > 0 {
				var ramBytes int64
				ramPart := strings.TrimSpace(outputStr[ramIdx+ramStart+1:])
				if _, err := fmt.Sscanf(ramPart, "%d", &ramBytes); err == nil && ramBytes > 0 {
					vramGB := int(ramBytes / (1024 * 1024 * 1024))
					// Cap at reasonable maximum (some systems report shared memory)
					if vramGB <= 128 {
						specs.GPUVRAMGB = vramGB
					}
				}
			}
		}

		if specs.DetectionMethod == "none" {
			specs.DetectionMethod = "wmi"
		}
	}
}

// detectGPULinux detects GPU on Linux
func (d *Detector) detectGPULinux(ctx context.Context, specs *Specs) {
	// Try lspci for GPU name
	cmd := exec.CommandContext(ctx, "lspci")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "vga") || strings.Contains(lineLower, "3d") || strings.Contains(lineLower, "display") {
			specs.GPUDetected = true
			// Extract GPU name (everything after the device class)
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				specs.GPUName = strings.TrimSpace(parts[len(parts)-1])
			}
			specs.DetectionMethod = "lspci"
			// lspci doesn't provide VRAM info, so we leave it at 0
			break
		}
	}
}

// detectGPUMacOS detects GPU on macOS
func (d *Detector) detectGPUMacOS(ctx context.Context, specs *Specs) {
	// Use system_profiler
	cmd := exec.CommandContext(ctx, "system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Chipset Model:") {
		specs.GPUDetected = true
		
		// Extract GPU name
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Chipset Model:") {
				parts := strings.Split(line, "Chipset Model:")
				if len(parts) >= 2 {
					specs.GPUName = strings.TrimSpace(parts[1])
				}
				break
			}
		}

		// Try to extract VRAM
		for _, line := range lines {
			if strings.Contains(line, "VRAM") {
				var vramMB int
				if _, err := fmt.Sscanf(line, "%*s VRAM: %d", &vramMB); err == nil && vramMB > 0 {
					specs.GPUVRAMGB = vramMB / 1024
				}
				break
			}
		}

		specs.DetectionMethod = "system_profiler"
	}
}
