package model

// ChangeType represents the type of change made
type ChangeType int

const (
	ChangeTypeAdd ChangeType = iota
	ChangeTypeUpdate
	ChangeTypeDelete
)

// Change represents a single change to an env file
type Change struct {
	Type     ChangeType
	FilePath string
	Entry    *Entry
	OldValue string // For updates: the previous value
}

// ChangeStack tracks changes for undo/redo functionality
type ChangeStack struct {
	changes []Change
	current int // Index of current position in stack (-1 means no changes)
	maxSize int // Maximum number of changes to track
}

// NewChangeStack creates a new change stack with a max size
func NewChangeStack(maxSize int) *ChangeStack {
	return &ChangeStack{
		changes: make([]Change, 0),
		current: -1,
		maxSize: maxSize,
	}
}

// Push adds a new change to the stack
func (cs *ChangeStack) Push(change Change) {
	// Remove any redo history
	if cs.current < len(cs.changes)-1 {
		cs.changes = cs.changes[:cs.current+1]
	}

	// Add new change
	cs.changes = append(cs.changes, change)
	cs.current++

	// Trim if exceeds max size
	if len(cs.changes) > cs.maxSize {
		cs.changes = cs.changes[1:]
		cs.current--
	}
}

// Undo reverts the last change and returns it
func (cs *ChangeStack) Undo() (*Change, bool) {
	if cs.current < 0 {
		return nil, false
	}

	change := cs.changes[cs.current]
	cs.current--
	return &change, true
}

// Redo re-applies the last undone change
func (cs *ChangeStack) Redo() (*Change, bool) {
	if cs.current >= len(cs.changes)-1 {
		return nil, false
	}

	cs.current++
	return &cs.changes[cs.current], true
}

// CanUndo returns true if there's something to undo
func (cs *ChangeStack) CanUndo() bool {
	return cs.current >= 0
}

// CanRedo returns true if there's something to redo
func (cs *ChangeStack) CanRedo() bool {
	return cs.current < len(cs.changes)-1
}

// Clear removes all changes
func (cs *ChangeStack) Clear() {
	cs.changes = cs.changes[:0]
	cs.current = -1
}

// GetHistory returns the current change history for display
func (cs *ChangeStack) GetHistory() []Change {
	result := make([]Change, len(cs.changes))
	copy(result, cs.changes)
	return result
}

// GetCurrentPosition returns the current position in history
func (cs *ChangeStack) GetCurrentPosition() int {
	return cs.current
}
