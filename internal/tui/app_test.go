package tui

import (
	"fmt"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type mockProgram struct {
	err error
}

func (m *mockProgram) Run() (tea.Model, error) {
	return nil, m.err
}

func withMockProgram(err error) func() {
	old := createProgram
	createProgram = func(model tea.Model, opts ...tea.ProgramOption) programRunner {
		return &mockProgram{err: err}
	}
	return func() { createProgram = old }
}

func TestRunWithIO(t *testing.T) {
	defer withMockProgram(nil)()
	err := RunWithIO(strings.NewReader(""), io.Discard)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunWithIOError(t *testing.T) {
	defer withMockProgram(fmt.Errorf("test error"))()
	err := RunWithIO(strings.NewReader(""), io.Discard)
	if err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), "running TUI") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestRunWithIOAltScreen(t *testing.T) {
	defer withMockProgram(nil)()
	// out == nil triggers the alt screen branch
	err := RunWithIO(strings.NewReader(""), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunWithIONilInput(t *testing.T) {
	defer withMockProgram(nil)()
	// in == nil means no WithInput added
	err := RunWithIO(nil, io.Discard)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun(t *testing.T) {
	defer withMockProgram(nil)()
	err := Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunWithOutput(t *testing.T) {
	defer withMockProgram(nil)()
	err := RunWithOutput(io.Discard)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
