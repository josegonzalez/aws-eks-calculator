package tui

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)


type programRunner interface {
	Run() (tea.Model, error)
}

var createProgram = func(model tea.Model, opts ...tea.ProgramOption) programRunner {
	return tea.NewProgram(model, opts...)
}

// Run starts the TUI application.
func Run() error {
	return RunWithIO(nil, nil)
}

// RunWithOutput starts the TUI with a custom output writer.
// Pass nil for default (stdout with alt screen).
func RunWithOutput(out io.Writer) error {
	return RunWithIO(nil, out)
}

// RunWithIO starts the TUI with custom input and output.
// Pass nil for defaults (stdin/stdout with alt screen).
func RunWithIO(in io.Reader, out io.Writer) error {
	model := NewModel()

	var opts []tea.ProgramOption
	if out != nil {
		opts = append(opts, tea.WithOutput(out))
	} else {
		opts = append(opts, tea.WithAltScreen())
	}
	if in != nil {
		opts = append(opts, tea.WithInput(in))
	}

	p := createProgram(model, opts...)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}
	return nil
}
