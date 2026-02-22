package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors — blue/cyan theme for a cloud cost calculator feel.
	primaryColor   = lipgloss.Color("#3B82F6")
	accentColor    = lipgloss.Color("#06B6D4")
	successColor   = lipgloss.Color("#10B981")
	warningColor   = lipgloss.Color("#F59E0B")
	errorColor     = lipgloss.Color("#EF4444")
	mutedColor     = lipgloss.Color("#6B7280")
	textColor      = lipgloss.Color("#E5E7EB")
	dimColor       = lipgloss.Color("#9CA3AF")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	SubSectionStyle = lipgloss.NewStyle().
				Foreground(accentColor)

	LabelStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	ValueStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true)

	MoneyStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true)

	BigMoneyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	MutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	BlurredInputStyle = lipgloss.NewStyle().
				Foreground(dimColor)

	SelectedPresetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(primaryColor).
				Bold(true).
				Padding(0, 1)

	NormalPresetStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 1)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(dimColor).
				Padding(0, 1)
)
