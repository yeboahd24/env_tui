package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/envtui/envtui/internal/storage"
	"github.com/envtui/envtui/internal/ui/styles"
)

// BackupViewMode represents the mode of the backup view
type BackupViewMode int

const (
	BackupViewModeList BackupViewMode = iota
	BackupViewModeConfirmRestore
	BackupViewModeConfirmDelete
)

// BackupView displays and manages backup files
type BackupView struct {
	backups      []storage.BackupInfo
	selected     int
	filePath     string
	mode         BackupViewMode
	width        int
	height       int
	message      string
	messageTimer time.Time
}

// NewBackupView creates a new backup view
func NewBackupView(filePath string, backups []storage.BackupInfo) BackupView {
	return BackupView{
		backups:  backups,
		filePath: filePath,
		selected: 0,
		mode:     BackupViewModeList,
	}
}

// SetSize sets the dimensions of the view
func (bv *BackupView) SetSize(width, height int) {
	bv.width = width
	bv.height = height
}

// Init initializes the view
func (bv BackupView) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (bv BackupView) Update(msg tea.Msg) (BackupView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch bv.mode {
		case BackupViewModeConfirmRestore:
			switch msg.String() {
			case "y", "Y":
				return bv, bv.confirmRestore()
			case "n", "N", "esc":
				bv.mode = BackupViewModeList
				return bv, nil
			}
		case BackupViewModeConfirmDelete:
			switch msg.String() {
			case "y", "Y":
				return bv, bv.confirmDelete()
			case "n", "N", "esc":
				bv.mode = BackupViewModeList
				return bv, nil
			}
		default:
			switch msg.String() {
			case "q", "esc":
				return bv, tea.Quit
			case "up", "k":
				if bv.selected > 0 {
					bv.selected--
				}
			case "down", "j":
				if bv.selected < len(bv.backups)-1 {
					bv.selected++
				}
			case "r":
				if len(bv.backups) > 0 {
					bv.mode = BackupViewModeConfirmRestore
				}
			case "d":
				if len(bv.backups) > 0 {
					bv.mode = BackupViewModeConfirmDelete
				}
			}
		}
	}
	return bv, nil
}

func (bv BackupView) confirmRestore() tea.Cmd {
	if bv.selected >= 0 && bv.selected < len(bv.backups) {
		backup := bv.backups[bv.selected]
		err := storage.RestoreBackup(backup.Path, bv.filePath)
		if err != nil {
			bv.message = fmt.Sprintf("Error restoring: %v", err)
		} else {
			bv.message = "Backup restored successfully!"
			bv.mode = BackupViewModeList
		}
		bv.messageTimer = time.Now()
	}
	return nil
}

func (bv BackupView) confirmDelete() tea.Cmd {
	if bv.selected >= 0 && bv.selected < len(bv.backups) {
		backup := bv.backups[bv.selected]
		err := storage.DeleteBackup(backup.Path)
		if err != nil {
			bv.message = fmt.Sprintf("Error deleting: %v", err)
		} else {
			bv.message = "Backup deleted successfully!"
			// Remove from list
			bv.backups = append(bv.backups[:bv.selected], bv.backups[bv.selected+1:]...)
			if bv.selected >= len(bv.backups) && bv.selected > 0 {
				bv.selected--
			}
			bv.mode = BackupViewModeList
		}
		bv.messageTimer = time.Now()
	}
	return nil
}

// View renders the backup view
func (bv BackupView) View() string {
	if bv.width == 0 {
		return "Loading..."
	}

	var sections []string

	// Title
	title := styles.TitleStyle.Render("Backup Manager")
	sections = append(sections, title)

	// File info
	subtitle := styles.SubtitleStyle.Render(fmt.Sprintf("📁 %s", bv.filePath))
	sections = append(sections, subtitle)

	// Message area
	if bv.message != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22C55E")).
			Padding(1, 1)
		sections = append(sections, msgStyle.Render(bv.message))
	}

	// Backup list or confirmation dialog
	switch bv.mode {
	case BackupViewModeConfirmRestore:
		sections = append(sections, bv.renderConfirmDialog("restore"))
	case BackupViewModeConfirmDelete:
		sections = append(sections, bv.renderConfirmDialog("delete"))
	default:
		sections = append(sections, bv.renderBackupList())
	}

	// Help
	help := bv.renderHelp()
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (bv BackupView) renderBackupList() string {
	if len(bv.backups) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Padding(2, 2).
			Render("No backups found for this file.")
	}

	listHeight := bv.height - 12
	var items []string

	start := 0
	if bv.selected > listHeight/2 {
		start = bv.selected - listHeight/2
	}
	end := start + listHeight
	if end > len(bv.backups) {
		end = len(bv.backups)
	}

	for i := start; i < end; i++ {
		backup := bv.backups[i]
		item := bv.renderBackupItem(backup, i == bv.selected)
		items = append(items, item)
	}

	list := strings.Join(items, "\n")
	return styles.BorderStyle.Width(bv.width - 4).Height(listHeight).Render(list)
}

func (bv BackupView) renderBackupItem(backup storage.BackupInfo, selected bool) string {
	style := styles.ListItemStyle
	if selected {
		style = styles.SelectedItemStyle
	}

	timeStr := backup.Timestamp.Format("Jan 02 15:04:05")
	sizeStr := formatBytes(backup.Size)

	content := fmt.Sprintf("%s (%s)", timeStr, sizeStr)
	return style.Width(bv.width - 6).Render(content)
}

func (bv BackupView) renderConfirmDialog(action string) string {
	if bv.selected >= len(bv.backups) {
		return ""
	}

	backup := bv.backups[bv.selected]

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F59E0B")).
		Padding(2, 4).
		Width(bv.width - 8)

	timeStr := backup.Timestamp.Format("Jan 02 15:04:05")
	content := fmt.Sprintf("Are you sure you want to %s the backup from %s?\n\n[y/N]", action, timeStr)

	return dialogStyle.Render(content)
}

func (bv BackupView) renderHelp() string {
	helpItems := []string{
		styles.HelpKeyStyle.Render("↑/k") + " " + styles.HelpDescStyle.Render("up"),
		styles.HelpKeyStyle.Render("↓/j") + " " + styles.HelpDescStyle.Render("down"),
		styles.HelpKeyStyle.Render("r") + " " + styles.HelpDescStyle.Render("restore"),
		styles.HelpKeyStyle.Render("d") + " " + styles.HelpDescStyle.Render("delete"),
		styles.HelpKeyStyle.Render("Esc/q") + " " + styles.HelpDescStyle.Render("close"),
	}

	return strings.Join(helpItems, styles.HelpSeparatorStyle.Render(" • "))
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// GetSelectedBackup returns the currently selected backup
func (bv BackupView) GetSelectedBackup() *storage.BackupInfo {
	if bv.selected >= 0 && bv.selected < len(bv.backups) {
		return &bv.backups[bv.selected]
	}
	return nil
}
