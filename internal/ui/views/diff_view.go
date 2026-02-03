package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/envtui/envtui/internal/model"
	"github.com/envtui/envtui/internal/ui/styles"
)

// DiffView displays unsaved changes in an env file
type DiffView struct {
	currentState  *model.EnvFile
	originalState *model.EnvFile
	width         int
	height        int
}

// DiffEntry represents a single difference between current and original
type DiffEntry struct {
	Key      string
	OldValue string
	NewValue string
	Type     DiffType
}

type DiffType int

const (
	DiffAdded DiffType = iota
	DiffModified
	DiffDeleted
)

// NewDiffView creates a new diff view comparing current and original states
func NewDiffView(current, original *model.EnvFile) DiffView {
	return DiffView{
		currentState:  current,
		originalState: original,
	}
}

// SetSize sets the dimensions of the diff view
func (dv *DiffView) SetSize(width, height int) {
	dv.width = width
	dv.height = height
}

// ComputeDifferences calculates the differences between current and original
func (dv DiffView) ComputeDifferences() []DiffEntry {
	var diffs []DiffEntry

	if dv.originalState == nil || dv.currentState == nil {
		return diffs
	}

	// Build maps for easier comparison
	originalEntries := make(map[string]string)
	for _, entry := range dv.originalState.Entries {
		if entry.Type == model.KeyValueEntry {
			originalEntries[entry.Key] = entry.Value
		}
	}

	currentEntries := make(map[string]string)
	for _, entry := range dv.currentState.Entries {
		if entry.Type == model.KeyValueEntry {
			currentEntries[entry.Key] = entry.Value
		}
	}

	// Check for added and modified entries
	for key, newValue := range currentEntries {
		if oldValue, exists := originalEntries[key]; !exists {
			diffs = append(diffs, DiffEntry{
				Key:      key,
				NewValue: newValue,
				Type:     DiffAdded,
			})
		} else if oldValue != newValue {
			diffs = append(diffs, DiffEntry{
				Key:      key,
				OldValue: oldValue,
				NewValue: newValue,
				Type:     DiffModified,
			})
		}
	}

	// Check for deleted entries
	for key, oldValue := range originalEntries {
		if _, exists := currentEntries[key]; !exists {
			diffs = append(diffs, DiffEntry{
				Key:      key,
				OldValue: oldValue,
				Type:     DiffDeleted,
			})
		}
	}

	return diffs
}

// View renders the diff view
func (dv DiffView) View() string {
	if dv.width == 0 {
		return "Loading..."
	}

	diffs := dv.ComputeDifferences()

	if len(diffs) == 0 {
		return lipgloss.NewStyle().
			Width(dv.width).
			Height(dv.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No unsaved changes - file is up to date")
	}

	var sections []string

	// Title
	title := styles.TitleStyle.Render(fmt.Sprintf("Unsaved Changes - %d differences", len(diffs)))
	sections = append(sections, title)

	// Subtitle with file info
	subtitle := styles.SubtitleStyle.Render(fmt.Sprintf("📁 %s", dv.currentState.Path))
	sections = append(sections, subtitle)

	// Diff entries
	listHeight := dv.height - 8
	var items []string

	for _, diff := range diffs {
		items = append(items, dv.renderDiffEntry(diff))
	}

	list := strings.Join(items, "\n")
	listBox := styles.BorderStyle.Width(dv.width - 4).Height(listHeight).Render(list)
	sections = append(sections, listBox)

	// Help
	help := dv.renderHelp()
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (dv DiffView) renderDiffEntry(diff DiffEntry) string {
	var prefix string
	var color lipgloss.Color

	switch diff.Type {
	case DiffAdded:
		prefix = "+"
		color = lipgloss.Color("#22C55E") // Green
	case DiffModified:
		prefix = "~"
		color = lipgloss.Color("#F59E0B") // Yellow/Orange
	case DiffDeleted:
		prefix = "-"
		color = lipgloss.Color("#EF4444") // Red
	}

	style := lipgloss.NewStyle().
		Foreground(color).
		Width(dv.width - 6)

	keyStr := styles.KeyStyle.Render(diff.Key)

	switch diff.Type {
	case DiffAdded:
		return style.Render(fmt.Sprintf("%s %s = %s", prefix, keyStr, diff.NewValue))
	case DiffModified:
		return style.Render(fmt.Sprintf("%s %s: %s → %s", prefix, keyStr, diff.OldValue, diff.NewValue))
	case DiffDeleted:
		return style.Render(fmt.Sprintf("%s %s = %s", prefix, keyStr, diff.OldValue))
	}

	return ""
}

func (dv DiffView) renderHelp() string {
	helpItems := []string{
		styles.HelpKeyStyle.Render("Esc") + " " + styles.HelpDescStyle.Render("close diff view"),
		styles.HelpKeyStyle.Render("q") + " " + styles.HelpDescStyle.Render("quit"),
	}

	return strings.Join(helpItems, styles.HelpSeparatorStyle.Render(" • "))
}
