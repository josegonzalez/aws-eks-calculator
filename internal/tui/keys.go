package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all keybindings for the application.
type KeyMap struct {
	NextInput key.Binding
	PrevInput key.Binding
	PrevTab key.Binding
	NextTab key.Binding
	Export    key.Binding
	Help      key.Binding
	Back      key.Binding
	Select    key.Binding
	Quit      key.Binding
	Up        key.Binding
	Down      key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		NextInput: key.NewBinding(
			key.WithKeys("tab", "down"),
			key.WithHelp("↓/tab", "next field"),
		),
		PrevInput: key.NewBinding(
			key.WithKeys("shift+tab", "up"),
			key.WithHelp("↑/shift+tab", "prev field"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "prev capability"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next capability"),
		),
		Export: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
	}
}

// ShortHelp returns keybindings to show in the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextInput, k.PrevTab, k.NextTab, k.Export, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextInput, k.PrevInput},
		{k.PrevTab, k.NextTab},
		{k.Export, k.Help, k.Quit},
	}
}
