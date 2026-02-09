package memory

// Store is the interface for the file-based memory store.
type Store interface {
	GetSystemMemory() (*SystemMemory, error)
	SaveSystemMemory(m *SystemMemory) error
	GetUserMemory(userID string) (*UserMemory, error)
	SaveUserMemory(userID string, m *UserMemory) error
	MemoryStatus() (*MemoryStatus, error)
}
