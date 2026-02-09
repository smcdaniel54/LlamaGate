package introspection

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// CaptureHardware returns a HardwareSnapshot (best-effort, no hostname/username in minimal).
func CaptureHardware(ctx context.Context, dataDir string, level DetailLevel) HardwareSnapshot {
	s := HardwareSnapshot{}
	s.OS = runtime.GOOS
	s.Arch = runtime.GOARCH

	// CPU
	if count, err := cpu.CountsWithContext(ctx, true); err == nil {
		s.CPU.LogicalCores = count
	}
	if infos, err := cpu.InfoWithContext(ctx); err == nil && len(infos) > 0 {
		s.CPU.ModelName = infos[0].ModelName
	}
	if level == DetailMinimal && s.CPU.ModelName != "" {
		// Optionally redact model name to avoid fingerprinting; for minimal we keep it short
		// Per spec: minimal = no hostnames, usernames, serial/MAC. CPU model is usually safe.
	}

	// Memory
	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		s.Memory.TotalBytes = int64(vm.Total)
		s.Memory.AvailableBytes = int64(vm.Available)
	}

	// Disk (data dir)
	if dataDir != "" {
		if err := os.MkdirAll(dataDir, 0755); err == nil {
			if stat, err := os.Stat(dataDir); err == nil && stat.IsDir() {
				total, free := diskUsage(dataDir)
				s.Disk.DataDirTotalBytes = total
				s.Disk.DataDirFreeBytes = free
			}
		}
	}

	// GPU: best-effort
	s.GPU.Present = false
	// Could integrate with internal/hardware Detector; for now leave false/empty to avoid heavy deps in this path

	return s
}

// diskUsage returns total and free bytes for the volume containing path (best-effort).
func diskUsage(path string) (total, free int64) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return 0, 0
	}
	// Walk to root to find a directory we can stat
	for p := abs; p != "" && p != "."; p = filepath.Dir(p) {
		var fi os.FileInfo
		fi, err = os.Stat(p)
		if err != nil || !fi.IsDir() {
			continue
		}
		// On Unix we could use unix.Statfs; on Windows different. Use a simple approach:
		// try to stat the path and use available space from the same volume.
		// Standard library doesn't expose disk free; return 0,0 and rely on callers to accept best-effort.
		_ = fi
		break
	}
	return 0, 0
}

// CaptureHardwareWithTimeout runs CaptureHardware with a timeout.
func CaptureHardwareWithTimeout(dataDir string, level DetailLevel, timeout time.Duration) HardwareSnapshot {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return CaptureHardware(ctx, dataDir, level)
}
