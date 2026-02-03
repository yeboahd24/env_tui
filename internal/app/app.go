package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/envtui/envtui/internal/model"
	"github.com/envtui/envtui/internal/parser"
	"github.com/envtui/envtui/internal/storage"
	"github.com/envtui/envtui/internal/ui/views"
)

func logDebug(msg string) {
	f, _ := os.OpenFile("/tmp/envtui_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05"), msg)
		f.Close()
	}
}

type ViewMode int

const (
	ViewModeList ViewMode = iota
	ViewModeEdit
	ViewModeAdd
	ViewModeDiff
	ViewModeBackup
)

type Model struct {
	envFiles         []*model.EnvFile
	originalStates   []*model.EnvFile // Original states for diff view
	currentFileIndex int
	listView         views.ListView
	editView         views.EditView
	diffView         views.DiffView
	backupView       views.BackupView
	viewMode         ViewMode
	err              error
	validationIssues []model.ValidationIssue
	changeStack      *model.ChangeStack
}

// New creates a model with a single file (backward compatibility)
func New(filePath string) Model {
	return NewMultiFile([]string{filePath})
}

// NewMultiFile creates a model with multiple files
func NewMultiFile(filePaths []string) Model {
	if len(filePaths) == 0 {
		return Model{err: fmt.Errorf("no files provided")}
	}

	var envFiles []*model.EnvFile
	var originalStates []*model.EnvFile
	var firstErr error

	for _, path := range filePaths {
		envFile, err := storage.ReadFile(path)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		envFiles = append(envFiles, envFile)
		// Store original state for diff view
		originalStates = append(originalStates, envFile.Clone())
	}

	if len(envFiles) == 0 {
		return Model{err: firstErr}
	}

	// Load the first file
	currentFile := envFiles[0]
	issues := currentFile.Validate()

	// Create list view and set files for copy operations
	listView := views.NewListView(currentFile.FilterEntries(""))
	listView.SetFiles(envFiles, 0)

	return Model{
		envFiles:         envFiles,
		originalStates:   originalStates,
		currentFileIndex: 0,
		listView:         listView,
		viewMode:         ViewModeList,
		validationIssues: issues,
		changeStack:      model.NewChangeStack(100), // Track up to 100 changes
	}
}

// GetCurrentEnvFile returns the currently active env file
func (m Model) GetCurrentEnvFile() *model.EnvFile {
	if m.currentFileIndex >= 0 && m.currentFileIndex < len(m.envFiles) {
		return m.envFiles[m.currentFileIndex]
	}
	return nil
}

// GetOriginalState returns the original state of the current file
func (m Model) GetOriginalState() *model.EnvFile {
	if m.currentFileIndex >= 0 && m.currentFileIndex < len(m.originalStates) {
		return m.originalStates[m.currentFileIndex]
	}
	return nil
}

// ShowDiffView shows the diff view comparing current state to original
func (m *Model) ShowDiffView() {
	current := m.GetCurrentEnvFile()
	original := m.GetOriginalState()
	if current != nil && original != nil {
		m.diffView = views.NewDiffView(current, original)
		m.diffView.SetSize(m.listView.Width(), m.listView.Height())
		m.viewMode = ViewModeDiff
	}
}

// GetCurrentFileName returns the filename of the current env file
func (m Model) GetCurrentFileName() string {
	if envFile := m.GetCurrentEnvFile(); envFile != nil {
		return filepath.Base(envFile.Path)
	}
	return ""
}

// SwitchToFile switches to the env file at the given index
func (m *Model) SwitchToFile(index int) {
	// Preserve dimensions when switching files
	oldWidth := m.listView.Width()
	oldHeight := m.listView.Height()
	m.currentFileIndex = index
	m.listView = views.NewListView(m.GetCurrentEnvFile().FilterEntries(""))
	m.listView.SetSize(oldWidth, oldHeight)
	// Set files for copy operations
	m.listView.SetFiles(m.envFiles, index)
}

// TrackChange records a change for undo/redo
func (m *Model) TrackChange(changeType model.ChangeType, entry *model.Entry, oldValue string) {
	if m.changeStack == nil {
		return
	}

	envFile := m.GetCurrentEnvFile()
	if envFile == nil {
		return
	}

	change := model.Change{
		Type:     changeType,
		FilePath: envFile.Path,
		Entry: &model.Entry{
			Type:     entry.Type,
			Key:      entry.Key,
			Value:    entry.Value,
			Comment:  entry.Comment,
			Line:     entry.Line,
			Exported: entry.Exported,
			IsSecret: entry.IsSecret,
		},
		OldValue: oldValue,
	}

	m.changeStack.Push(change)
	logDebug(fmt.Sprintf("Tracked change: %v for key %s", changeType, entry.Key))
}

// Undo reverts the last change
func (m *Model) Undo() bool {
	if m.changeStack == nil || !m.changeStack.CanUndo() {
		return false
	}

	change, ok := m.changeStack.Undo()
	if !ok {
		return false
	}

	envFile := m.GetCurrentEnvFile()
	if envFile == nil {
		return false
	}

	switch change.Type {
	case model.ChangeTypeAdd:
		// Undo add = delete the entry
		envFile.DeleteEntry(change.Entry.Key)
		logDebug(fmt.Sprintf("Undo add: deleted %s", change.Entry.Key))
	case model.ChangeTypeUpdate:
		// Undo update = restore old value
		envFile.UpdateEntry(change.Entry.Key, change.OldValue)
		logDebug(fmt.Sprintf("Undo update: restored %s to %s", change.Entry.Key, change.OldValue))
	case model.ChangeTypeDelete:
		// Undo delete = re-add the entry
		envFile.AddEntry(&model.Entry{
			Type:     change.Entry.Type,
			Key:      change.Entry.Key,
			Value:    change.Entry.Value,
			Comment:  change.Entry.Comment,
			Line:     change.Entry.Line,
			Exported: change.Entry.Exported,
			IsSecret: change.Entry.IsSecret,
		})
		logDebug(fmt.Sprintf("Undo delete: restored %s", change.Entry.Key))
	}

	// Save the file
	if err := storage.WriteFile(envFile); err != nil {
		m.err = err
		return false
	}

	// Refresh the list view
	oldWidth := m.listView.Width()
	oldHeight := m.listView.Height()
	m.listView = views.NewListView(envFile.FilterEntries(""))
	m.listView.SetSize(oldWidth, oldHeight)
	m.validationIssues = envFile.Validate()

	return true
}

// Redo re-applies the last undone change
func (m *Model) Redo() bool {
	if m.changeStack == nil || !m.changeStack.CanRedo() {
		return false
	}

	change, ok := m.changeStack.Redo()
	if !ok {
		return false
	}

	envFile := m.GetCurrentEnvFile()
	if envFile == nil {
		return false
	}

	switch change.Type {
	case model.ChangeTypeAdd:
		// Redo add = add the entry back
		envFile.AddEntry(&model.Entry{
			Type:     change.Entry.Type,
			Key:      change.Entry.Key,
			Value:    change.Entry.Value,
			Comment:  change.Entry.Comment,
			Line:     change.Entry.Line,
			Exported: change.Entry.Exported,
			IsSecret: change.Entry.IsSecret,
		})
		logDebug(fmt.Sprintf("Redo add: restored %s", change.Entry.Key))
	case model.ChangeTypeUpdate:
		// Redo update = apply the new value
		envFile.UpdateEntry(change.Entry.Key, change.Entry.Value)
		logDebug(fmt.Sprintf("Redo update: set %s to %s", change.Entry.Key, change.Entry.Value))
	case model.ChangeTypeDelete:
		// Redo delete = delete the entry
		envFile.DeleteEntry(change.Entry.Key)
		logDebug(fmt.Sprintf("Redo delete: removed %s", change.Entry.Key))
	}

	// Save the file
	if err := storage.WriteFile(envFile); err != nil {
		m.err = err
		return false
	}

	// Refresh the list view
	oldWidth := m.listView.Width()
	oldHeight := m.listView.Height()
	m.listView = views.NewListView(envFile.FilterEntries(""))
	m.listView.SetSize(oldWidth, oldHeight)
	m.validationIssues = envFile.Validate()

	return true
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case views.BulkDeleteMsg:
		// Handle bulk delete
		envFile := m.GetCurrentEnvFile()
		if envFile != nil && len(msg.Keys) > 0 {
			for _, key := range msg.Keys {
				entry := envFile.GetEntry(key)
				if entry != nil {
					m.TrackChange(model.ChangeTypeDelete, entry, "")
					envFile.DeleteEntry(key)
				}
			}
			if err := storage.WriteFile(envFile); err != nil {
				m.err = err
				return m, nil
			}
			oldWidth := m.listView.Width()
			oldHeight := m.listView.Height()
			m.listView = views.NewListView(envFile.FilterEntries(""))
			m.listView.SetSize(oldWidth, oldHeight)
			m.validationIssues = envFile.Validate()
		}
		return m, nil
	case views.CopyEntryMsg:
		// Handle copy entry to another file
		if msg.TargetIndex >= 0 && msg.TargetIndex < len(m.envFiles) && msg.Entry != nil {
			targetFile := m.envFiles[msg.TargetIndex]
			// Check if entry already exists
			existing := targetFile.GetEntry(msg.Entry.Key)
			if existing == nil {
				newEntry := &model.Entry{
					Type:     model.KeyValueEntry,
					Key:      msg.Entry.Key,
					Value:    msg.Entry.Value,
					IsSecret: msg.Entry.IsSecret,
				}
				targetFile.AddEntry(newEntry)
				if err := storage.WriteFile(targetFile); err != nil {
					m.err = err
				}
			}
			m.listView.SetCopyMode(false)
		}
		return m, nil
	case tea.KeyMsg:
		keyStr := msg.String()
		logDebug(fmt.Sprintf("Key pressed: '%s' (Type: %v, Runes: %v)", msg.String(), msg.Type, msg.Runes))
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// File switching with number keys (only when NOT in copy mode)
		if m.viewMode == ViewModeList && !m.listView.IsCopyMode() {
			switch keyStr {
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				idx := int(keyStr[0] - '1') // Convert '1' to 0, '2' to 1, etc.
				if idx < len(m.envFiles) {
					logDebug(fmt.Sprintf("Switching to file %d: %s", idx+1, m.envFiles[idx].Path))
					m.SwitchToFile(idx)
					return m, nil
				}
			}
		}

		switch m.viewMode {
		case ViewModeList:
			return m.handleListKeys(msg)
		case ViewModeEdit, ViewModeAdd:
			// Handle enter/esc at app level first
			keyStr := msg.String()
			logDebug(fmt.Sprintf("Checking key: '%s'", keyStr))
			if keyStr == "enter" || keyStr == "esc" {
				logDebug("Key is enter or esc, calling handleEditKeys")
				return m.handleEditKeys(msg)
			}
			// Pass other keys to edit view
			logDebug("Passing key to editView")
			var cmd tea.Cmd
			m.editView, cmd = m.editView.Update(msg)
			logDebug(fmt.Sprintf("After editView.Update: key='%s' value='%s'", m.editView.GetKey(), m.editView.GetValue()))
			return m, cmd
		case ViewModeDiff:
			// Handle esc/q to return to list view
			if keyStr == "esc" || keyStr == "q" {
				logDebug("Leaving diff view, returning to list")
				m.viewMode = ViewModeList
				return m, nil
			}
		case ViewModeBackup:
			// Handle esc/q to return to list view
			if keyStr == "esc" || keyStr == "q" {
				logDebug("Leaving backup view, returning to list")
				// Reload the file in case a backup was restored
				if envFile := m.GetCurrentEnvFile(); envFile != nil {
					oldWidth := m.listView.Width()
					oldHeight := m.listView.Height()
					m.envFiles[m.currentFileIndex], _ = storage.ReadFile(envFile.Path)
					m.listView = views.NewListView(m.envFiles[m.currentFileIndex].FilterEntries(""))
					m.listView.SetSize(oldWidth, oldHeight)
				}
				m.viewMode = ViewModeList
				return m, nil
			}
			// Pass other keys to backup view
			var cmd tea.Cmd
			m.backupView, cmd = m.backupView.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		if m.err == nil {
			var cmd tea.Cmd
			switch m.viewMode {
			case ViewModeList:
				m.listView, cmd = m.listView.Update(msg)
			case ViewModeEdit, ViewModeAdd:
				m.editView, cmd = m.editView.Update(msg)
			case ViewModeDiff:
				m.diffView.SetSize(msg.Width, msg.Height)
			case ViewModeBackup:
				m.backupView.SetSize(msg.Width, msg.Height)
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()
	logDebug(fmt.Sprintf("handleListKeys: key='%s'", keyStr))

	// Handle copy mode file selection
	if m.listView.IsCopyMode() {
		switch keyStr {
		case "esc", "q":
			m.listView.SetCopyMode(false)
			return m, nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(keyStr[0] - '1')
			if idx < len(m.envFiles) && idx != m.currentFileIndex {
				// Copy the selected entry to the target file
				selected := m.listView.GetSelected()
				if selected != nil {
					targetFile := m.envFiles[idx]
					existing := targetFile.GetEntry(selected.Key)
					if existing == nil {
						newEntry := &model.Entry{
							Type:     model.KeyValueEntry,
							Key:      selected.Key,
							Value:    selected.Value,
							IsSecret: selected.IsSecret,
						}
						targetFile.AddEntry(newEntry)
						if err := storage.WriteFile(targetFile); err != nil {
							m.err = err
						}
					}
				}
				m.listView.SetCopyMode(false)
				return m, nil
			}
		}
		// In copy mode, only allow the above keys
		return m, nil
	}

	switch keyStr {
	case "q":
		logDebug("'q' pressed - quitting")
		return m, tea.Quit
	case "a":
		logDebug("'a' pressed - switching to add mode")
		m.viewMode = ViewModeAdd
		m.editView = views.NewEditView(views.EditModeAdd, nil, m.listView.Width())
		return m, m.editView.Init()
	case "e":
		logDebug("'e' pressed - switching to edit mode")
		// Get selected entry and edit
		if selected := m.listView.GetSelected(); selected != nil {
			m.viewMode = ViewModeEdit
			m.editView = views.NewEditView(views.EditModeEdit, selected, m.listView.Width())
			return m, m.editView.Init()
		}
	case "d":
		logDebug("'d' pressed - deleting entry")
		// Delete selected entry
		envFile := m.GetCurrentEnvFile()
		if selected := m.listView.GetSelected(); selected != nil && envFile != nil {
			// Track the delete for undo
			m.TrackChange(model.ChangeTypeDelete, selected, "")
			envFile.DeleteEntry(selected.Key)
			if err := storage.WriteFile(envFile); err != nil {
				m.err = err
				return m, nil
			}
			// Preserve dimensions when recreating list view
			oldWidth := m.listView.Width()
			oldHeight := m.listView.Height()
			m.listView = views.NewListView(envFile.FilterEntries(""))
			m.listView.SetSize(oldWidth, oldHeight)
			m.validationIssues = envFile.Validate()
		}
		return m, nil
	case "u":
		logDebug("'u' pressed - undoing")
		if m.Undo() {
			logDebug("Undo successful")
		} else {
			logDebug("Nothing to undo")
		}
		return m, nil
	case "r":
		logDebug("'r' pressed - redoing")
		if m.Redo() {
			logDebug("Redo successful")
		} else {
			logDebug("Nothing to redo")
		}
		return m, nil
	case "v":
		logDebug("'v' pressed - showing diff view")
		m.ShowDiffView()
		return m, nil
	case "b":
		logDebug("'b' pressed - showing backup view")
		envFile := m.GetCurrentEnvFile()
		if envFile != nil {
			backups, err := storage.ListBackups(envFile.Path)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.backupView = views.NewBackupView(envFile.Path, backups)
			m.backupView.SetSize(m.listView.Width(), m.listView.Height())
			m.viewMode = ViewModeBackup
		}
		return m, nil
	default:
		logDebug(fmt.Sprintf("Passing key '%s' to listView", keyStr))
		var cmd tea.Cmd
		m.listView, cmd = m.listView.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()
	logDebug(fmt.Sprintf("handleEditKeys: key='%s'", keyStr))
	envFile := m.GetCurrentEnvFile()
	if envFile == nil {
		return m, nil
	}

	switch keyStr {
	case "esc":
		logDebug("ESC pressed - returning to list")
		m.viewMode = ViewModeList
		return m, nil
	case "enter":
		key := m.editView.GetKey()
		value := m.editView.GetValue()
		logDebug(fmt.Sprintf("ENTER pressed - key='%s' value='%s' editMode=%d", key, value, m.editView.GetMode()))

		if key == "" {
			logDebug("Empty key, canceling")
			m.viewMode = ViewModeList
			return m, nil
		}

		// Check the edit view mode before changing viewMode
		if m.editView.GetMode() == views.EditModeAdd {
			logDebug(fmt.Sprintf("Adding new entry: Key='%s' Value='%s'", key, value))
			entry := &model.Entry{
				Type:     model.KeyValueEntry,
				Key:      key,
				Value:    value,
				IsSecret: parser.IsSecretKey(key),
			}
			logDebug(fmt.Sprintf("Entry String() output: '%s'", entry.String()))
			envFile.AddEntry(entry)
			// Track the add for undo
			m.TrackChange(model.ChangeTypeAdd, entry, "")
		} else {
			logDebug("Updating existing entry")
			// Get old value before updating for undo tracking
			oldEntry := envFile.GetEntry(key)
			oldValue := ""
			if oldEntry != nil {
				oldValue = oldEntry.Value
			}
			envFile.UpdateEntry(key, value)
			// Track the update for undo
			updatedEntry := &model.Entry{
				Type:     model.KeyValueEntry,
				Key:      key,
				Value:    value,
				IsSecret: parser.IsSecretKey(key),
			}
			m.TrackChange(model.ChangeTypeUpdate, updatedEntry, oldValue)
		}

		logDebug(fmt.Sprintf("Saving file with %d entries", len(envFile.Entries)))
		if err := storage.WriteFile(envFile); err != nil {
			logDebug(fmt.Sprintf("Save error: %v", err))
			m.err = err
			m.viewMode = ViewModeList
			return m, nil
		}
		logDebug("File saved successfully")

		m.viewMode = ViewModeList

		// Preserve the width/height from the old list view
		oldWidth := m.listView.Width()
		oldHeight := m.listView.Height()

		m.listView = views.NewListView(envFile.FilterEntries(""))

		// Restore dimensions if we had them
		if oldWidth > 0 && oldHeight > 0 {
			m.listView.SetSize(oldWidth, oldHeight)
		}

		m.validationIssues = envFile.Validate()
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit", m.err)
	}

	envFile := m.GetCurrentEnvFile()
	if envFile == nil || len(envFile.Entries) == 0 {
		fileName := m.GetCurrentFileName()
		if fileName == "" {
			fileName = "No file"
		}
		return fmt.Sprintf("[%s] No entries found\n\nPress a to add, q to quit", fileName)
	}

	switch m.viewMode {
	case ViewModeList:
		// Collect git info for all files
		var gitInfos []storage.FileGitInfo
		for _, ef := range m.envFiles {
			gitInfos = append(gitInfos, storage.GetFileGitInfo(ef.Path))
		}
		return m.listView.ViewWithFiles(m.envFiles, m.currentFileIndex, gitInfos)
	case ViewModeEdit, ViewModeAdd:
		return m.editView.View()
	case ViewModeDiff:
		return m.diffView.View()
	case ViewModeBackup:
		return m.backupView.View()
	}

	return ""
}
