package introspection

import (
	"runtime"
	"time"
)

// BuildInfo can be set via ldflags at build time.
var (
	BuildAppVersion = "0.11.0"
	BuildGitCommit  = ""
	BuildDate       = ""
)

// UptimeStart is set at process start for uptime calculation.
var UptimeStart = time.Now()

// CaptureRuntime returns a RuntimeSnapshot (no I/O, deterministic except uptime).
func CaptureRuntime() RuntimeSnapshot {
	return RuntimeSnapshot{
		AppVersion:  BuildAppVersion,
		GitCommit:   BuildGitCommit,
		BuildDate:   BuildDate,
		GoVersion:   runtime.Version(),
		UptimeSecs:  time.Since(UptimeStart).Seconds(),
		ProcessRSS:  captureProcessRSS(),
		CurrentTime: time.Now(),
	}
}

// captureProcessRSS returns process RSS in bytes (best-effort).
func captureProcessRSS() int64 {
	// runtime does not expose RSS; on many platforms we'd need syscall or exec.
	// Return 0 for now; can be filled via build-tag implementations (e.g. read /proc/self/status on Linux).
	return 0
}
