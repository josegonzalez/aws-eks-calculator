package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors — blue/cyan theme for a cloud cost calculator feel.
	PrimaryColor   = lipgloss.Color("#3B82F6")
	SecondaryColor = lipgloss.Color("#60A5FA")
	AccentColor    = lipgloss.Color("#06B6D4")
	SuccessColor   = lipgloss.Color("#10B981")
	WarningColor   = lipgloss.Color("#F59E0B")
	ErrorColor     = lipgloss.Color("#EF4444")
	MutedColor     = lipgloss.Color("#6B7280")
	TextColor      = lipgloss.Color("#E5E7EB")
	DimColor       = lipgloss.Color("#9CA3AF")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentColor)

	SubSectionStyle = lipgloss.NewStyle().
				Foreground(AccentColor)

	LabelStyle = lipgloss.NewStyle().
			Foreground(DimColor)

	ValueStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Bold(true)

	MoneyStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Bold(true)

	BigMoneyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor)

	MutedStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)

	BlurredInputStyle = lipgloss.NewStyle().
				Foreground(DimColor)

	SelectedPresetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(PrimaryColor).
				Bold(true).
				Padding(0, 1)

	NormalPresetStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2)

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(PrimaryColor).
			Padding(0, 1)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(DimColor).
				Padding(0, 1)
)
