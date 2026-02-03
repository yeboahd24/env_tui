package views

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/envtui/envtui/internal/model"
	"github.com/envtui/envtui/internal/ui/styles"
)

func logDebug(msg string) {
	f, _ := os.OpenFile("/tmp/editview_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05"), msg)
		f.Close()
	}
}

type EditMode int

const (
	EditModeAdd EditMode = iota
	EditModeEdit
)

type Template struct {
	Name        string
	Key         string
	Value       string
	Description string
}

var QuickTemplates = []Template{
	{Name: "DATABASE_URL", Key: "DATABASE_URL", Value: "postgresql://user:password@localhost:5432/dbname", Description: "Database connection URL"},
	{Name: "API_KEY", Key: "API_KEY", Value: "your-api-key-here", Description: "Generic API key"},
	{Name: "DEBUG", Key: "DEBUG", Value: "true", Description: "Debug mode flag"},
	{Name: "PORT", Key: "PORT", Value: "3000", Description: "Server port"},
	{Name: "NODE_ENV", Key: "NODE_ENV", Value: "development", Description: "Node environment"},
	{Name: "JWT_SECRET", Key: "JWT_SECRET", Value: "your-jwt-secret-here", Description: "JWT signing secret"},
	{Name: "REDIS_URL", Key: "REDIS_URL", Value: "redis://localhost:6379", Description: "Redis connection URL"},
	{Name: "AWS_ACCESS_KEY", Key: "AWS_ACCESS_KEY_ID", Value: "your-access-key", Description: "AWS access key ID"},
	{Name: "AWS_SECRET", Key: "AWS_SECRET_ACCESS_KEY", Value: "your-secret-key", Description: "AWS secret key"},
	{Name: "S3_BUCKET", Key: "S3_BUCKET_NAME", Value: "my-bucket", Description: "S3 bucket name"},
}

type EditView struct {
	mode          EditMode
	keyInput      textinput.Model
	valueInput    textinput.Model
	focused       int
	entry         *model.Entry
	width         int
	height        int
	showTemplates bool
	templateIndex int
}

func NewEditView(mode EditMode, entry *model.Entry, width int) EditView {
	keyInput := textinput.New()
	keyInput.Placeholder = "Type key name here..."
	keyInput.CharLimit = 100
	keyInput.Focus()
	keyInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	// Use bright cyan for high visibility
	keyInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	keyInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	if width > 0 {
		keyInput.Width = width - 10
	}

	valueInput := textinput.New()
	valueInput.Placeholder = "Type value here..."
	valueInput.CharLimit = 500
	valueInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	// Use bright cyan for high visibility
	valueInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	valueInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	if width > 0 {
		valueInput.Width = width - 10
	}

	// Set values for both modes
	if entry != nil && mode == EditModeEdit {
		keyInput.SetValue(entry.Key)
		valueInput.SetValue(entry.Value)
	} else {
		keyInput.SetValue("")
		valueInput.SetValue("")
	}

	// Process a dummy message to activate the input
	keyInput.Update(tea.KeyMsg{})

	return EditView{
		mode:       mode,
		keyInput:   keyInput,
		valueInput: valueInput,
		focused:    0,
		entry:      entry,
		width:      width,
	}
}

func (ev EditView) Init() tea.Cmd {
	// Focus the key input and return blink
	return tea.Batch(
		ev.keyInput.Focus(),
		textinput.Blink,
	)
}

func (ev EditView) Update(msg tea.Msg) (EditView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ev.width = msg.Width
		ev.height = msg.Height
		ev.keyInput.Width = msg.Width - 10
		ev.valueInput.Width = msg.Width - 10
		return ev, nil

	case tea.KeyMsg:
		// Handle template mode
		if ev.showTemplates {
			switch msg.String() {
			case "esc", "q":
				ev.showTemplates = false
				return ev, nil
			case "up", "k":
				if ev.templateIndex > 0 {
					ev.templateIndex--
				}
				return ev, nil
			case "down", "j":
				if ev.templateIndex < len(QuickTemplates)-1 {
					ev.templateIndex++
				}
				return ev, nil
			case "enter":
				// Apply selected template
				template := QuickTemplates[ev.templateIndex]
				ev.keyInput.SetValue(template.Key)
				ev.valueInput.SetValue(template.Value)
				ev.showTemplates = false
				// Keep focus on value field so user can see both fields populated
				ev.focused = 1
				ev.keyInput.Blur()
				ev.valueInput.Focus()
				return ev, nil
			}
			return ev, nil
		}

		switch msg.String() {
		case "enter", "esc":
			return ev, nil
		case "t":
			// Show template picker
			ev.showTemplates = true
			ev.templateIndex = 0
			return ev, nil
		case "tab", "shift+tab", "down":
			// Don't allow switching to value field if key is empty
			if ev.focused == 0 && ev.keyInput.Value() == "" {
				// Stay on key field, show error state
				return ev, nil
			}
			if ev.focused == 0 {
				ev.focused = 1
				ev.keyInput.Blur()
				ev.valueInput.Focus()
				return ev, textinput.Blink
			} else {
				ev.focused = 0
				ev.valueInput.Blur()
				ev.keyInput.Focus()
				return ev, textinput.Blink
			}
		case "up":
			if ev.focused == 1 {
				ev.focused = 0
				ev.valueInput.Blur()
				ev.keyInput.Focus()
				return ev, textinput.Blink
			}
		}
	}

	// Always update the focused input
	if ev.focused == 0 {
		ev.keyInput, cmd = ev.keyInput.Update(msg)
	} else {
		ev.valueInput, cmd = ev.valueInput.Update(msg)
	}

	return ev, cmd
}

func (ev EditView) View() string {
	// Show template picker if active
	if ev.showTemplates {
		return ev.renderTemplatePicker()
	}

	title := "Add Entry"
	if ev.mode == EditModeEdit {
		title = "Edit Entry"
	}

	// Check if key is empty and we're in add mode
	keyIsEmpty := ev.keyInput.Value() == ""
	showKeyError := keyIsEmpty && ev.mode == EditModeAdd && ev.focused == 0

	// Active field styling
	activeLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Padding(0, 1)

	inactiveLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Padding(0, 1)

	errorLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true).
		Padding(0, 1)

	activeIndicator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Render(" ▶ ")

	inactiveIndicator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#374151")).
		Render("   ")

	titleStyle := styles.TitleStyle.Render(title)

	// Build key field - use plain rendering without border styling interference
	var keyLabel, keyBox string
	if ev.focused == 0 {
		if showKeyError {
			keyLabel = errorLabelStyle.Render("⚠ ENTER KEY NAME FIRST") + activeIndicator
		} else {
			keyLabel = activeLabelStyle.Render("STEP 1: Enter Key Name") + activeIndicator
		}
		// Simple border without extra styling
		keyBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Render(ev.keyInput.View())
	} else {
		keyLabel = inactiveLabelStyle.Render("Key: "+ev.keyInput.Value()) + inactiveIndicator
		keyBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Render(ev.keyInput.View())
	}

	// Build value field
	var valueLabel, valueBox string
	if ev.focused == 1 {
		valueLabel = activeLabelStyle.Render("STEP 2: Enter Value") + activeIndicator
		valueBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Render(ev.valueInput.View())
	} else {
		valueLabel = inactiveLabelStyle.Render("Value") + inactiveIndicator
		valueBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Render(ev.valueInput.View())
	}

	// Help text with clearer instructions
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Padding(1, 1)

	help := helpStyle.Render("Tab: next field (key required)  •  t: templates  •  Enter: save  •  Esc: cancel")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle,
		"",
		keyLabel,
		keyBox,
		"",
		valueLabel,
		valueBox,
		"",
		help,
	)
}

func (ev EditView) renderTemplatePicker() string {
	// Create a prominent banner for template mode
	bannerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#8B5CF6")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 2).
		Width(ev.width - 4).
		Render(" 📋 TEMPLATE MODE: Press t again to toggle ")

	titleStyle := styles.TitleStyle.Render("Quick Templates - Select a template")

	// Show preview of what will be filled
	selectedTemplate := QuickTemplates[ev.templateIndex]
	previewStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#10B981")).
		Padding(1, 2).
		Width(ev.width - 4).
		Render(fmt.Sprintf("Preview: %s=%s", selectedTemplate.Key, selectedTemplate.Value))

	var items []string
	for i, template := range QuickTemplates {
		style := lipgloss.NewStyle().Padding(0, 2)
		if i == ev.templateIndex {
			style = style.
				Background(lipgloss.Color("#7C3AED")).
				Foreground(lipgloss.Color("#FFFFFF"))
		}

		nameStyle := lipgloss.NewStyle().Bold(true)
		if i == ev.templateIndex {
			nameStyle = nameStyle.Foreground(lipgloss.Color("#FFFFFF"))
		} else {
			nameStyle = nameStyle.Foreground(lipgloss.Color("#7C3AED"))
		}

		item := style.Render(
			nameStyle.Render(template.Name) + " - " + template.Description,
		)
		items = append(items, item)
	}

	list := lipgloss.JoinVertical(lipgloss.Left, items...)
	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Render(list)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Padding(1, 1)

	help := helpStyle.Render("↑/↓ or k/j: navigate  •  Enter: apply template  •  Esc: cancel")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		bannerStyle,
		"",
		titleStyle,
		previewStyle,
		"",
		listBox,
		"",
		help,
	)
}

func (ev EditView) GetKey() string {
	return ev.keyInput.Value()
}

func (ev EditView) GetValue() string {
	return ev.valueInput.Value()
}

func (ev EditView) GetMode() EditMode {
	return ev.mode
}
