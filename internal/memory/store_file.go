package memory

import (
	"os"
	"path/filepath"
	"time"
)

const (
	systemFile  = "system.json"
	usersDir    = "users"
	sessionsDir = "sessions"
)

// FileStore implements Store with JSON files under a directory.
type FileStore struct {
	dir    string
	limits Limits
	fl     *FileLock
}

// NewFileStore creates a FileStore with base directory and limits.
func NewFileStore(dir string, limits Limits) (*FileStore, error) {
	if limits.PinnedMax <= 0 {
		limits.PinnedMax = DefaultPinnedMax
	}
	if limits.RecentMax <= 0 {
		limits.RecentMax = DefaultRecentMax
	}
	if limits.NotesMax <= 0 {
		limits.NotesMax = DefaultNotesMax
	}
	if limits.MaxEntryChars <= 0 {
		limits.MaxEntryChars = DefaultMaxEntryChars
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	if _, err := EnsureDir(dir, usersDir); err != nil {
		return nil, err
	}
	if _, err := EnsureDir(dir, sessionsDir); err != nil {
		return nil, err
	}
	return &FileStore{dir: dir, limits: limits, fl: NewFileLock()}, nil
}

// GetSystemMemory reads system.json.
func (s *FileStore) GetSystemMemory() (*SystemMemory, error) {
	path := filepath.Join(s.dir, systemFile)
	unlock := s.fl.Lock("system")
	defer unlock()
	var m SystemMemory
	if err := ReadJSON(path, &m); err != nil {
		if os.IsNotExist(err) {
			return &SystemMemory{UpdatedAt: time.Now()}, nil
		}
		return nil, err
	}
	// Apply caps on read
	m.Notes = s.limits.trimStrings(m.Notes, s.limits.NotesMax, s.limits.MaxEntryChars)
	return &m, nil
}

// SaveSystemMemory writes system.json with caps applied.
func (s *FileStore) SaveSystemMemory(m *SystemMemory) error {
	if m == nil {
		return nil
	}
	m.UpdatedAt = time.Now()
	m.Notes = s.limits.trimStrings(m.Notes, s.limits.NotesMax, s.limits.MaxEntryChars)
	path := filepath.Join(s.dir, systemFile)
	unlock := s.fl.Lock("system")
	defer unlock()
	return WriteAtomicJSON(path, m)
}

// GetUserMemory reads users/<userID>.json.
func (s *FileStore) GetUserMemory(userID string) (*UserMemory, error) {
	if userID == "" {
		return nil, nil
	}
	safeName := filepath.Base(userID)
	if safeName == "" || safeName == "." || safeName == ".." {
		return nil, nil
	}
	path := filepath.Join(s.dir, usersDir, safeName+".json")
	unlock := s.fl.Lock("user:" + userID)
	defer unlock()
	var m UserMemory
	if err := ReadJSON(path, &m); err != nil {
		if os.IsNotExist(err) {
			return &UserMemory{UserID: userID, UpdatedAt: time.Now()}, nil
		}
		return nil, err
	}
	m.UserID = userID
	m.Pinned = s.limits.trimStrings(m.Pinned, s.limits.PinnedMax, s.limits.MaxEntryChars)
	m.Recent = s.limits.trimStrings(m.Recent, s.limits.RecentMax, s.limits.MaxEntryChars)
	return &m, nil
}

// SaveUserMemory writes users/<userID>.json with caps applied.
func (s *FileStore) SaveUserMemory(userID string, m *UserMemory) error {
	if userID == "" || m == nil {
		return nil
	}
	safeName := filepath.Base(userID)
	if safeName == "" || safeName == "." || safeName == ".." {
		return nil
	}
	m.UserID = userID
	m.UpdatedAt = time.Now()
	m.Pinned = s.limits.trimStrings(m.Pinned, s.limits.PinnedMax, s.limits.MaxEntryChars)
	m.Recent = s.limits.trimStrings(m.Recent, s.limits.RecentMax, s.limits.MaxEntryChars)
	path := filepath.Join(s.dir, usersDir, safeName+".json")
	unlock := s.fl.Lock("user:" + userID)
	defer unlock()
	return WriteAtomicJSON(path, m)
}

// MemoryStatus returns counts and last updated times (no full dumps).
func (s *FileStore) MemoryStatus() (*MemoryStatus, error) {
	status := &MemoryStatus{}
	systemPath := filepath.Join(s.dir, systemFile)
	if fi, err := os.Stat(systemPath); err == nil {
		t := fi.ModTime()
		status.SystemUpdatedAt = &t
		status.TotalSizeBytes += fi.Size()
	}
	usersPath := filepath.Join(s.dir, usersDir)
	entries, _ := os.ReadDir(usersPath)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		status.UserCount++
		if fi, err := e.Info(); err == nil {
			status.TotalSizeBytes += fi.Size()
		}
	}
	sessionsPath := filepath.Join(s.dir, sessionsDir)
	sessionEntries, _ := os.ReadDir(sessionsPath)
	for _, e := range sessionEntries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		status.SessionCount++
		if fi, err := e.Info(); err == nil {
			status.TotalSizeBytes += fi.Size()
		}
	}
	return status, nil
}
