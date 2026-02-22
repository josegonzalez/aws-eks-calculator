package main

import (
	"fmt"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	old := tuiRun
	defer func() { tuiRun = old }()
	tuiRun = func() error { return nil }

	if err := run(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunError(t *testing.T) {
	old := tuiRun
	defer func() { tuiRun = old }()
	tuiRun = func() error { return fmt.Errorf("test error") }

	err := run()
	if err == nil {
		t.Error("expected error")
	}
}

func TestMainSuccess(t *testing.T) {
	oldRun := tuiRun
	defer func() { tuiRun = oldRun }()
	tuiRun = func() error { return nil }

	// main() should return without calling osExit
	main()
}

func TestMainError(t *testing.T) {
	oldRun := tuiRun
	oldExit := osExit
	defer func() { tuiRun = oldRun; osExit = oldExit }()

	tuiRun = func() error { return fmt.Errorf("test error") }
	exitCode := -1
	osExit = func(code int) { exitCode = code }

	main()

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}
