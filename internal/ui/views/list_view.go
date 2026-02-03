package views

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/envtui/envtui/internal/model"
	"github.com/envtui/envtui/internal/storage"
	"github.com/envtui/envtui/internal/ui/styles"
)

// Bulk delete message
type BulkDeleteMsg struct {
	Keys []string
}

// Copy entry message
type CopyEntryMsg struct {
	Entry       *model.Entry
	TargetIndex int
}

type SortMode int

const (
	SortModeAlphabetical SortMode = iota
	SortModeByCategory
	SortModeByValueLength
)

type ListView struct {
	entries         []*model.Entry
	filteredEntries []*model.Entry
	selected        int
	searchInput     textinput.Model
	searching       bool
	showSecrets     bool
	width           int
	height          int
	envFiles        []*model.EnvFile
	currentIndex    int
	showDiffs       bool
	selectedItems   map[string]bool // Track multi-selected items
	bulkMode        bool            // Whether in bulk selection mode
	sortMode        SortMode
	copyMode        bool // Whether in copy mode (selecting target file)
	copyTargetIndex int  // Target file index for copy operation
}

type keyMap struct {
	Up             key.Binding
	Down           key.Binding
	Search         key.Binding
	Toggle         key.Binding
	Diff           key.Binding
	Undo           key.Binding
	Redo           key.Binding
	ToggleSelect   key.Binding
	BulkDelete     key.Binding
	ClearSelection key.Binding
	Sort           key.Binding
	Copy           key.Binding
	Template       key.Binding
	Backup         key.Binding
	Quit           key.Binding
	Enter          key.Binding
	Escape         key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "move down"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "toggle secrets"),
	),
	Diff: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "compare files"),
	),
	Undo: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "undo"),
	),
	Redo: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "redo"),
	),
	ToggleSelect: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle select"),
	),
	BulkDelete: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "bulk delete"),
	),
	ClearSelection: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear selection"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort mode"),
	),
	Copy: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "copy to file"),
	),
	Template: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "templates"),
	),
	Backup: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "backups"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "edit"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func NewListView(entries []*model.Entry) ListView {
	ti := textinput.New()
	ti.Placeholder = "Search entries..."
	ti.CharLimit = 50

	lv := ListView{
		entries:         entries,
		filteredEntries: entries,
		searchInput:     ti,
		selectedItems:   make(map[string]bool),
	}

	return lv
}

func (lv ListView) Init() tea.Cmd {
	return nil
}

func (lv ListView) Update(msg tea.Msg) (ListView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		lv.width = msg.Width
		lv.height = msg.Height
		lv.searchInput.Width = msg.Width - 4

	case tea.KeyMsg:
		// Handle copy mode (file picker for copying entries)
		if lv.copyMode {
			switch msg.String() {
			case "esc", "q":
				lv.copyMode = false
				lv.copyTargetIndex = -1
				return lv, nil
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				idx := int(msg.String()[0] - '1')
				if idx < len(lv.envFiles) && idx != lv.currentIndex {
					lv.copyTargetIndex = idx
					return lv, nil
				}
			}
			return lv, nil
		}

		if lv.searching {
			switch {
			case key.Matches(msg, keys.Escape):
				lv.searching = false
				lv.searchInput.SetValue("")
				lv.filteredEntries = lv.entries
				return lv, nil
			case key.Matches(msg, keys.Enter):
				lv.searching = false
				return lv, nil
			default:
				lv.searchInput, cmd = lv.searchInput.Update(msg)
				lv.filterEntries(lv.searchInput.Value())
				lv.selected = 0
				return lv, cmd
			}
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return lv, tea.Quit
		case key.Matches(msg, keys.Up):
			if lv.selected > 0 {
				lv.selected--
			}
		case key.Matches(msg, keys.Down):
			if lv.selected < len(lv.filteredEntries)-1 {
				lv.selected++
			}
		case key.Matches(msg, keys.Search):
			lv.searching = true
			lv.searchInput.Focus()
			return lv, textinput.Blink
		case key.Matches(msg, keys.Toggle):
			lv.showSecrets = !lv.showSecrets
		case key.Matches(msg, keys.Diff):
			lv.ToggleDiffs()
		case key.Matches(msg, keys.ToggleSelect):
			// Toggle selection of current item
			if lv.selected >= 0 && lv.selected < len(lv.filteredEntries) {
				entry := lv.filteredEntries[lv.selected]
				if lv.selectedItems[entry.Key] {
					delete(lv.selectedItems, entry.Key)
				} else {
					lv.selectedItems[entry.Key] = true
				}
				lv.bulkMode = len(lv.selectedItems) > 0
			}
		case key.Matches(msg, keys.BulkDelete):
			var keys []string
			for k := range lv.selectedItems {
				keys = append(keys, k)
			}
			return lv, tea.Batch(cmd, func() tea.Msg {
				return BulkDeleteMsg{Keys: keys}
			})
		case key.Matches(msg, keys.ClearSelection):
			lv.selectedItems = make(map[string]bool)
			lv.bulkMode = false
		case key.Matches(msg, keys.Sort):
			lv.cycleSortMode()
		case key.Matches(msg, keys.Copy):
			// Debug: log the copy key detection
			if len(lv.envFiles) > 1 && lv.selected >= 0 && lv.selected < len(lv.filteredEntries) {
				lv.copyMode = true
				lv.copyTargetIndex = -1
				return lv, nil
			}
		}
	}

	return lv, cmd
}

func (lv *ListView) filterEntries(query string) {
	if query == "" {
		lv.filteredEntries = lv.entries
		return
	}

	query = strings.ToLower(query)
	var filtered []*model.Entry

	for _, entry := range lv.entries {
		if strings.Contains(strings.ToLower(entry.Key), query) ||
			strings.Contains(strings.ToLower(entry.Value), query) {
			filtered = append(filtered, entry)
		}
	}

	lv.filteredEntries = filtered
}

func (lv ListView) View() string {
	return lv.ViewWithFiles(nil, 0, nil)
}

// ViewWithFiles renders the list view with file tabs
func (lv *ListView) ViewWithFiles(envFiles []*model.EnvFile, currentIndex int, gitInfos []storage.FileGitInfo) string {
	if lv.width == 0 {
		return "Loading..."
	}

	// Store files for comparison rendering
	lv.SetFiles(envFiles, currentIndex)

	var sections []string

	// Title with file tabs if multiple files
	var header string
	if len(envFiles) > 1 {
		// Show file tabs with label
		var tabs []string

		// Add "FILES:" label
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Bold(true).
			Padding(0, 1).
			MarginRight(1)
		tabs = append(tabs, labelStyle.Render("FILES:"))

		for i, ef := range envFiles {
			tabName := filepath.Base(ef.Path)
			entryCount := len(ef.FilterEntries(""))

			// Add git status icon if available
			gitIndicator := ""
			if i < len(gitInfos) && gitInfos[i].Status != storage.GitStatusNone {
				gitIndicator = storage.FormatGitStatusForTab(gitInfos[i].Status)
			}

			if i == currentIndex {
				// Active tab - bright purple with glow effect
				activeTabStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("#9333EA")).
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true).
					Padding(0, 2).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#C084FC")).
					MarginRight(1)
				tabs = append(tabs, activeTabStyle.
					Render(fmt.Sprintf(" ▶ %d:%s%s (%d) ", i+1, tabName, gitIndicator, entryCount)))
			} else {
				// Inactive tab - darker but still visible
				inactiveTabStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("#374151")).
					Foreground(lipgloss.Color("#9CA3AF")).
					Padding(0, 2).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#4B5563")).
					MarginRight(1)
				tabs = append(tabs, inactiveTabStyle.
					Render(fmt.Sprintf(" %d:%s%s (%d) ", i+1, tabName, gitIndicator, entryCount)))
			}
		}
		tabsRow := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

		// File indicator showing current file info
		currentFile := envFiles[currentIndex]
		fileInfo := fmt.Sprintf("📁 %s (%d entries)", filepath.Base(currentFile.Path), len(currentFile.FilterEntries("")))

		// Add git branch info if available
		if currentIndex < len(gitInfos) && gitInfos[currentIndex].Branch != "" {
			fileInfo += fmt.Sprintf(" (git: %s)", gitInfos[currentIndex].Branch)
		}

		title := styles.TitleStyle.Render("EnvTUI")
		subtitle := styles.SubtitleStyle.Render(fileInfo)
		header = lipgloss.JoinVertical(lipgloss.Left, title, tabsRow, subtitle)
	} else {
		title := styles.TitleStyle.Render("EnvTUI")
		subtitle := styles.SubtitleStyle.Render(fmt.Sprintf("%d entries", len(lv.entries)))

		// Add git status for single file
		if len(gitInfos) > 0 && gitInfos[0].Status != storage.GitStatusNone {
			subtitle = styles.SubtitleStyle.Render(fmt.Sprintf("%d entries %s", len(lv.entries), storage.FormatGitStatusForTab(gitInfos[0].Status)))
		}

		header = lipgloss.JoinHorizontal(lipgloss.Left, title, subtitle)
	}
	sections = append(sections, header)

	// Copy mode banner
	if lv.copyMode {
		copyBanner := lipgloss.NewStyle().
			Background(lipgloss.Color("#F59E0B")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 2).
			Width(lv.width - 4).
			Render(" 📋 COPY MODE: Select target file (1-9) or Esc to cancel ")
		sections = append(sections, copyBanner)
	}

	// Search input
	if lv.searching {
		searchBox := styles.BorderStyle.Render(lv.searchInput.View())
		sections = append(sections, searchBox)
	}

	// Entries list - calculate available height
	// Account for: header (3 rows) + help (5 rows) + padding (2) = 10 minimum
	listHeight := lv.height - 10
	if lv.searching {
		listHeight -= 3
	}
	// Adjust for tabs if shown (tabs take 2 extra rows)
	if len(envFiles) > 1 {
		listHeight -= 3
	}
	// Adjust for copy mode banner
	if lv.copyMode {
		listHeight -= 1
	}
	// Ensure minimum height
	if listHeight < 5 {
		listHeight = 5
	}

	var items []string
	start := max(0, lv.selected-listHeight/2)
	end := min(len(lv.filteredEntries), start+listHeight)

	for i := start; i < end; i++ {
		entry := lv.filteredEntries[i]
		item := lv.renderEntry(entry, i == lv.selected)
		items = append(items, item)
	}

	list := strings.Join(items, "\n")
	listBox := styles.BorderStyle.Width(lv.width - 4).Height(listHeight).Render(list)
	sections = append(sections, listBox)

	// Help
	help := lv.renderHelpWithFiles(len(envFiles) > 1)
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (lv ListView) renderEntry(entry *model.Entry, selected bool) string {
	style := styles.ListItemStyle
	if selected {
		style = styles.SelectedItemStyle
	}

	// Checkmark for selected items in bulk mode
	checkmark := "  "
	if lv.selectedItems[entry.Key] {
		checkmark = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22C55E")).
			Render("✓ ")
	}

	// Category indicator
	categoryColor := styles.CategoryColor(entry.Category())
	indicator := lipgloss.NewStyle().Foreground(categoryColor).Render("●")

	// Key with diff indicator
	keyStr := styles.KeyStyle.Render(entry.Key)

	// Check for differences with other files
	diffIndicator := ""
	if len(lv.envFiles) > 1 && lv.showDiffs {
		diffIndicator = lv.getDiffIndicator(entry)
	}

	// Value
	value := entry.Value
	if entry.IsSecret && !lv.showSecrets {
		value = entry.DisplayValue()
	}
	valueStr := styles.ValueStyle.Render(value)

	content := fmt.Sprintf("%s%s %s%s = %s", checkmark, indicator, keyStr, diffIndicator, valueStr)
	return style.Width(lv.width - 6).Render(content)
}

func (lv ListView) getDiffIndicator(entry *model.Entry) string {
	if len(lv.envFiles) <= 1 {
		return ""
	}

	currentFile := lv.envFiles[lv.currentIndex]
	currentEntry := currentFile.GetEntry(entry.Key)
	if currentEntry == nil {
		return ""
	}

	// Check against all other files
	hasDiff := false
	diffFiles := []string{}
	for i, ef := range lv.envFiles {
		if i == lv.currentIndex {
			continue
		}
		otherEntry := ef.GetEntry(entry.Key)
		if otherEntry == nil {
			hasDiff = true
			diffFiles = append(diffFiles, filepath.Base(ef.Path))
		} else if otherEntry.Value != currentEntry.Value {
			hasDiff = true
			diffFiles = append(diffFiles, filepath.Base(ef.Path))
		}
	}

	if hasDiff {
		if len(diffFiles) == 1 {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")).
				Render(" ⚠" + diffFiles[0])
		}
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Render(" ⚠" + fmt.Sprintf("%d files", len(diffFiles)))
	}
	return ""
}

func (lv ListView) renderHelp() string {
	return lv.renderHelpWithFiles(false)
}

func (lv ListView) renderHelpWithFiles(showFileShortcuts bool) string {
	if lv.searching {
		return styles.HelpDescStyle.Render("Press Enter to confirm search, Esc to cancel")
	}

	// Show copy mode help if active
	if lv.copyMode {
		helpItems := []string{
			styles.HelpKeyStyle.Render("1-9") + " " + styles.HelpDescStyle.Render("select target file"),
			styles.HelpKeyStyle.Render("Esc") + " " + styles.HelpDescStyle.Render("cancel"),
		}
		return strings.Join(helpItems, styles.HelpSeparatorStyle.Render(" • "))
	}

	// Build help in organized rows
	var rows []string
	separator := styles.HelpSeparatorStyle.Render(" • ")

	// Row 1: Navigation
	navItems := []string{
		styles.HelpKeyStyle.Render("↑/k") + " " + styles.HelpDescStyle.Render("up"),
		styles.HelpKeyStyle.Render("↓/j") + " " + styles.HelpDescStyle.Render("down"),
		styles.HelpKeyStyle.Render("/") + " " + styles.HelpDescStyle.Render("search"),
	}
	rows = append(rows, strings.Join(navItems, separator))

	// Row 2: CRUD Operations
	crudItems := []string{
		styles.HelpKeyStyle.Render("a") + " " + styles.HelpDescStyle.Render("add"),
		styles.HelpKeyStyle.Render("e") + " " + styles.HelpDescStyle.Render("edit"),
		styles.HelpKeyStyle.Render("d") + " " + styles.HelpDescStyle.Render("delete"),
		styles.HelpKeyStyle.Render("x") + " " + styles.HelpDescStyle.Render("secrets"),
	}
	// Add file-specific operations if multiple files
	if showFileShortcuts {
		crudItems = append(crudItems, styles.HelpKeyStyle.Render("y")+" "+styles.HelpDescStyle.Render("copy"))
	}
	rows = append(rows, strings.Join(crudItems, separator))

	// Row 3: History & Comparison
	historyItems := []string{
		styles.HelpKeyStyle.Render("u") + " " + styles.HelpDescStyle.Render("undo"),
		styles.HelpKeyStyle.Render("r") + " " + styles.HelpDescStyle.Render("redo"),
		styles.HelpKeyStyle.Render("v") + " " + styles.HelpDescStyle.Render("diff"),
		styles.HelpKeyStyle.Render("s") + " " + styles.HelpDescStyle.Render("sort"),
	}
	if showFileShortcuts {
		historyItems = append(historyItems, styles.HelpKeyStyle.Render("c")+" "+styles.HelpDescStyle.Render("compare"))
		historyItems = append(historyItems, styles.HelpKeyStyle.Render("1-9")+" "+styles.HelpDescStyle.Render("files"))
	}
	rows = append(rows, strings.Join(historyItems, separator))

	// Row 4: Copy Mode (only when active)
	if lv.copyMode {
		copyItems := []string{
			styles.HelpKeyStyle.Render("1-9") + " " + styles.HelpDescStyle.Render("select target file"),
			styles.HelpKeyStyle.Render("Esc") + " " + styles.HelpDescStyle.Render("cancel copy"),
		}
		rows = append(rows, strings.Join(copyItems, separator))
	}

	// Row 5: Bulk Selection (only when active)
	if lv.bulkMode {
		bulkItems := []string{
			styles.HelpKeyStyle.Render("space") + " " + styles.HelpDescStyle.Render("select"),
			styles.HelpKeyStyle.Render("D") + " " + styles.HelpDescStyle.Render("bulk del ("+fmt.Sprintf("%d", len(lv.selectedItems))+")"),
			styles.HelpKeyStyle.Render("Esc") + " " + styles.HelpDescStyle.Render("clear"),
		}
		rows = append(rows, strings.Join(bulkItems, separator))
	}

	// Row 5: Utilities & Quit
	utilItems := []string{
		styles.HelpKeyStyle.Render("t") + " " + styles.HelpDescStyle.Render("templates"),
		styles.HelpKeyStyle.Render("b") + " " + styles.HelpDescStyle.Render("backups"),
		styles.HelpKeyStyle.Render("q") + " " + styles.HelpDescStyle.Render("quit"),
	}
	rows = append(rows, strings.Join(utilItems, separator))

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (lv ListView) GetSelected() *model.Entry {
	if lv.selected >= 0 && lv.selected < len(lv.filteredEntries) {
		return lv.filteredEntries[lv.selected]
	}
	return nil
}

func (lv ListView) Width() int {
	return lv.width
}

func (lv ListView) Height() int {
	return lv.height
}

func (lv *ListView) SetSize(width, height int) {
	lv.width = width
	lv.height = height
	lv.searchInput.Width = width - 4
}

func (lv *ListView) SetFiles(envFiles []*model.EnvFile, currentIndex int) {
	lv.envFiles = envFiles
	lv.currentIndex = currentIndex
}

func (lv *ListView) ToggleDiffs() {
	lv.showDiffs = !lv.showDiffs
}

func (lv *ListView) cycleSortMode() {
	lv.sortMode = (lv.sortMode + 1) % 3
	lv.applySort()
}

func (lv *ListView) applySort() {
	switch lv.sortMode {
	case SortModeAlphabetical:
		sort.Slice(lv.filteredEntries, func(i, j int) bool {
			return lv.filteredEntries[i].Key < lv.filteredEntries[j].Key
		})
	case SortModeByCategory:
		sort.Slice(lv.filteredEntries, func(i, j int) bool {
			catI := lv.filteredEntries[i].Category()
			catJ := lv.filteredEntries[j].Category()
			if catI != catJ {
				return catI < catJ
			}
			return lv.filteredEntries[i].Key < lv.filteredEntries[j].Key
		})
	case SortModeByValueLength:
		sort.Slice(lv.filteredEntries, func(i, j int) bool {
			return len(lv.filteredEntries[i].Value) > len(lv.filteredEntries[j].Value)
		})
	}
}

func (lv ListView) GetSortModeName() string {
	switch lv.sortMode {
	case SortModeAlphabetical:
		return "alphabetical"
	case SortModeByCategory:
		return "by category"
	case SortModeByValueLength:
		return "by value length"
	}
	return ""
}

func (lv ListView) GetSelectedItems() []string {
	var keys []string
	for k := range lv.selectedItems {
		keys = append(keys, k)
	}
	return keys
}

func (lv *ListView) ClearSelection() {
	lv.selectedItems = make(map[string]bool)
	lv.bulkMode = false
}

func (lv ListView) IsCopyMode() bool {
	return lv.copyMode
}

func (lv *ListView) SetCopyMode(enabled bool) {
	lv.copyMode = enabled
	if !enabled {
		lv.copyTargetIndex = -1
	}
}

func (lv ListView) GetCopyTargetIndex() int {
	return lv.copyTargetIndex
}

func (lv *ListView) SetCopyTargetIndex(idx int) {
	lv.copyTargetIndex = idx
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
