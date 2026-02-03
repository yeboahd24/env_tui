package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Primary   = lipgloss.Color("#7C3AED")
	Secondary = lipgloss.Color("#10B981")
	Danger    = lipgloss.Color("#EF4444")
	Warning   = lipgloss.Color("#F59E0B")
	Info      = lipgloss.Color("#3B82F6")

	// Category colors
	DatabaseColor = lipgloss.Color("#3B82F6")
	AWSColor      = lipgloss.Color("#FF9500")
	APIColor      = lipgloss.Color("#10B981")
	SecretColor   = lipgloss.Color("#EF4444")
	OtherColor    = lipgloss.Color("#6B7280")
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(0, 1)

	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(0, 1)
)

// List styles
var (
	ListItemStyle = lipgloss.NewStyle().
			Padding(0, 2)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(Primary).
				Padding(0, 2)

	KeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB"))

	SecretValueStyle = lipgloss.NewStyle().
				Foreground(SecretColor)

	CommentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)
)

// Help styles
var (
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	HelpSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4B5563"))
)

func CategoryColor(category string) lipgloss.Color {
	switch category {
	case "database":
		return DatabaseColor
	case "aws":
		return AWSColor
	case "api":
		return APIColor
	case "secret":
		return SecretColor
	default:
		return OtherColor
	}
}
