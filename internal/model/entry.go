package model

type EntryType int

const (
	KeyValueEntry EntryType = iota
	CommentEntry
	BlankEntry
)

func (et EntryType) String() string {
	switch et {
	case KeyValueEntry:
		return "KeyValueEntry"
	case CommentEntry:
		return "CommentEntry"
	case BlankEntry:
		return "BlankEntry"
	default:
		return "Unknown"
	}
}

type Entry struct {
	Type     EntryType
	Key      string
	Value    string
	Comment  string
	Line     int
	Exported bool
	IsSecret bool
}

type EnvFile struct {
	Path         string
	Entries      []*Entry
	originalHash string // Hash of original file content for detecting changes
	isModified   bool   // Track if file has unsaved changes
}

// SetModified marks the file as having unsaved changes
func (ef *EnvFile) SetModified() {
	ef.isModified = true
}

// ClearModified clears the modified flag (call after saving)
func (ef *EnvFile) ClearModified() {
	ef.isModified = false
}

// IsModified returns true if the file has unsaved changes
func (ef *EnvFile) IsModified() bool {
	return ef.isModified
}

// Clone creates a deep copy of the EnvFile
func (ef *EnvFile) Clone() *EnvFile {
	clone := &EnvFile{
		Path:         ef.Path,
		originalHash: ef.originalHash,
		isModified:   ef.isModified,
		Entries:      make([]*Entry, len(ef.Entries)),
	}
	for i, entry := range ef.Entries {
		clone.Entries[i] = &Entry{
			Type:     entry.Type,
			Key:      entry.Key,
			Value:    entry.Value,
			Comment:  entry.Comment,
			Line:     entry.Line,
			Exported: entry.Exported,
			IsSecret: entry.IsSecret,
		}
	}
	return clone
}

func (e *Entry) String() string {
	switch e.Type {
	case KeyValueEntry:
		prefix := ""
		if e.Exported {
			prefix = "export "
		}

		suffix := ""
		if e.Comment != "" {
			suffix = " " + e.Comment
		}

		return prefix + e.Key + "=" + e.Value + suffix
	case CommentEntry:
		return e.Comment
	case BlankEntry:
		return ""
	}
	return ""
}

func (e *Entry) DisplayValue() string {
	if e.IsSecret {
		return "••••••••"
	}
	return e.Value
}

func (e *Entry) Category() string {
	if len(e.Key) == 0 {
		return "other"
	}

	if idx := findPrefix(e.Key, []string{"DB_", "DATABASE_"}); idx != -1 {
		return "database"
	}
	if idx := findPrefix(e.Key, []string{"AWS_", "S3_"}); idx != -1 {
		return "aws"
	}
	if idx := findPrefix(e.Key, []string{"API_", "HTTP_"}); idx != -1 {
		return "api"
	}
	if e.IsSecret {
		return "secret"
	}

	return "other"
}

func findPrefix(key string, prefixes []string) int {
	for i, prefix := range prefixes {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			return i
		}
	}
	return -1
}
