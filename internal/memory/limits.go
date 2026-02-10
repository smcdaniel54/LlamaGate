package memory

// Default limits for memory entries (can be overridden by config).
const (
	DefaultPinnedMax     = 50
	DefaultRecentMax     = 100
	DefaultNotesMax      = 100
	DefaultMaxEntryChars = 500
)

// Limits holds caps for memory store.
type Limits struct {
	PinnedMax     int
	RecentMax     int
	NotesMax      int
	MaxEntryChars int
}

// DefaultLimits returns the default Limits.
func DefaultLimits() Limits {
	return Limits{
		PinnedMax:     DefaultPinnedMax,
		RecentMax:     DefaultRecentMax,
		NotesMax:      DefaultNotesMax,
		MaxEntryChars: DefaultMaxEntryChars,
	}
}

// Apply caps slice length to maxLen and trims each string to maxChars.
func (l Limits) trimStrings(slice []string, maxLen int, maxChars int) []string {
	if maxLen <= 0 {
		maxLen = l.NotesMax
	}
	if maxChars <= 0 {
		maxChars = l.MaxEntryChars
	}
	out := make([]string, 0, len(slice))
	for i, s := range slice {
		if i >= maxLen {
			break
		}
		if len(s) > maxChars {
			s = s[:maxChars]
		}
		out = append(out, s)
	}
	return out
}
