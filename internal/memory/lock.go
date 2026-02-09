package memory

import (
	"os"
	"path/filepath"
	"sync"
)

// FileLock provides a per-key mutex for file writes (cross-platform).
type FileLock struct {
	mu   sync.Mutex
	keys map[string]*sync.Mutex
}

// NewFileLock creates a new FileLock.
func NewFileLock() *FileLock {
	return &FileLock{keys: make(map[string]*sync.Mutex)}
}

// Lock acquires the lock for the given key (e.g. file path or "system", "user:id").
func (f *FileLock) Lock(key string) func() {
	f.mu.Lock()
	m, ok := f.keys[key]
	if !ok {
		m = &sync.Mutex{}
		f.keys[key] = m
	}
	f.mu.Unlock()
	m.Lock()
	return func() { m.Unlock() }
}

// EnsureDir creates dir and returns its path.
func EnsureDir(baseDir string, subdir string) (string, error) {
	p := filepath.Join(baseDir, subdir)
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", err
	}
	return p, nil
}
