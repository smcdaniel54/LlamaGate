package memory

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAtomicJSON_ReadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := SystemMemory{
		UpdatedAt: time.Now(),
		Notes:     []string{"a", "b"},
	}
	err := WriteAtomicJSON(path, &data)
	require.NoError(t, err)
	var out SystemMemory
	err = ReadJSON(path, &out)
	require.NoError(t, err)
	assert.Equal(t, data.Notes, out.Notes)
}

func TestLimits_trimStrings(t *testing.T) {
	lim := Limits{PinnedMax: 2, RecentMax: 3, MaxEntryChars: 5}
	out := lim.trimStrings([]string{"hello", "world", "extra"}, 2, 5)
	assert.Equal(t, []string{"hello", "world"}, out)
	out = lim.trimStrings([]string{"toolong"}, 2, 3)
	assert.Equal(t, []string{"too"}, out)
}

func TestFileStore_GetSaveSystemMemory(t *testing.T) {
	dir := t.TempDir()
	limits := DefaultLimits()
	limits.NotesMax = 5
	store, err := NewFileStore(dir, limits)
	require.NoError(t, err)
	// Get when missing returns empty
	sys, err := store.GetSystemMemory()
	require.NoError(t, err)
	require.NotNil(t, sys)
	// Save then get
	sys.Notes = []string{"n1", "n2"}
	err = store.SaveSystemMemory(sys)
	require.NoError(t, err)
	sys2, err := store.GetSystemMemory()
	require.NoError(t, err)
	assert.Equal(t, []string{"n1", "n2"}, sys2.Notes)
	// Caps: save more than NotesMax; count should be trimmed to 5
	sys.Notes = make([]string, 10)
	for i := range sys.Notes {
		sys.Notes[i] = "x"
	}
	err = store.SaveSystemMemory(sys)
	require.NoError(t, err)
	sys3, err := store.GetSystemMemory()
	require.NoError(t, err)
	assert.LessOrEqual(t, len(sys3.Notes), 5)
	// MaxEntryChars cap: long string gets trimmed (use new dir to avoid overwriting)
	dir2 := t.TempDir()
	limits2 := DefaultLimits()
	limits2.NotesMax = 5
	limits2.MaxEntryChars = 5
	store2, err := NewFileStore(dir2, limits2)
	require.NoError(t, err)
	sys4 := &SystemMemory{Notes: []string{"longer than five chars"}}
	err = store2.SaveSystemMemory(sys4)
	require.NoError(t, err)
	sys5, err := store2.GetSystemMemory()
	require.NoError(t, err)
	require.Len(t, sys5.Notes, 1)
	assert.LessOrEqual(t, len(sys5.Notes[0]), 5)
}

func TestFileStore_GetSaveUserMemory(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir, DefaultLimits())
	require.NoError(t, err)
	u, err := store.GetUserMemory("alice")
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "alice", u.UserID)
	u.Pinned = []string{"p1"}
	err = store.SaveUserMemory("alice", u)
	require.NoError(t, err)
	u2, err := store.GetUserMemory("alice")
	require.NoError(t, err)
	assert.Equal(t, []string{"p1"}, u2.Pinned)
}

func TestFileStore_MemoryStatus(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir, DefaultLimits())
	require.NoError(t, err)
	err = store.SaveSystemMemory(&SystemMemory{Notes: []string{"x"}})
	require.NoError(t, err)
	_ = store.SaveUserMemory("u1", &UserMemory{UserID: "u1", Pinned: []string{"a"}})
	status, err := store.MemoryStatus()
	require.NoError(t, err)
	assert.NotNil(t, status.SystemUpdatedAt)
	assert.Equal(t, 1, status.UserCount)
	assert.Greater(t, status.TotalSizeBytes, int64(0))
}

func TestFileStore_ConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir, DefaultLimits())
	require.NoError(t, err)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			sys, _ := store.GetSystemMemory()
			if sys == nil {
				sys = &SystemMemory{}
			}
			sys.Notes = append(sys.Notes, "note")
			_ = store.SaveSystemMemory(sys)
		}(i)
	}
	wg.Wait()
	// Should have one system.json with no corruption (readable)
	sys, err := store.GetSystemMemory()
	require.NoError(t, err)
	require.NotNil(t, sys)
	// File must exist and be valid JSON
	path := filepath.Join(dir, systemFile)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}
