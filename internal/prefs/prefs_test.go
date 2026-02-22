package prefs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *store {
	t.Helper()
	return &store{dir: t.TempDir()}
}

func TestSaveAndLoad(t *testing.T) {
	s := newTestStore(t)
	p := Prefs{Region: "eu-west-1"}

	if err := s.save(p); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got := s.load()
	if got.Region != "eu-west-1" {
		t.Errorf("expected region eu-west-1, got %q", got.Region)
	}
}

func TestLoadMissing(t *testing.T) {
	s := newTestStore(t)
	got := s.load()
	if got.Region != "" {
		t.Errorf("expected empty region, got %q", got.Region)
	}
}

func TestLoadCorrupt(t *testing.T) {
	s := newTestStore(t)

	if err := os.WriteFile(s.path(), []byte("{invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := s.load()
	if got.Region != "" {
		t.Errorf("expected empty region for corrupt file, got %q", got.Region)
	}
}

func TestSaveCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "prefs")
	s := &store{dir: dir}

	if err := s.save(Prefs{Region: "ap-southeast-1"}); err != nil {
		t.Fatalf("save should create intermediate dirs: %v", err)
	}

	got := s.load()
	if got.Region != "ap-southeast-1" {
		t.Errorf("expected ap-southeast-1, got %q", got.Region)
	}
}

func TestSaveOverwrites(t *testing.T) {
	s := newTestStore(t)

	if err := s.save(Prefs{Region: "us-east-1"}); err != nil {
		t.Fatal(err)
	}
	if err := s.save(Prefs{Region: "eu-central-1"}); err != nil {
		t.Fatal(err)
	}

	got := s.load()
	if got.Region != "eu-central-1" {
		t.Errorf("expected eu-central-1, got %q", got.Region)
	}
}

func TestSetDir(t *testing.T) {
	dir := t.TempDir()
	SetDir(dir)
	t.Cleanup(func() { SetDir("") })

	if err := Save(Prefs{Region: "ap-south-1"}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got := Load()
	if got.Region != "ap-south-1" {
		t.Errorf("expected ap-south-1, got %q", got.Region)
	}
}

func TestLoadAndSaveViaPublicAPI(t *testing.T) {
	SetDir(t.TempDir())
	t.Cleanup(func() { SetDir("") })

	if err := Save(Prefs{Region: "eu-west-1"}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got := Load()
	if got.Region != "eu-west-1" {
		t.Errorf("expected eu-west-1, got %q", got.Region)
	}
}

func TestLoadReturnsZeroWhenDirEmpty(t *testing.T) {
	// defaultDir returns "" when overrideDir is "" and UserConfigDir fails.
	// We can't easily make UserConfigDir fail, but we can test the overrideDir
	// path by verifying that Load returns zero prefs from a nonexistent dir.
	SetDir(filepath.Join(t.TempDir(), "nonexistent"))
	t.Cleanup(func() { SetDir("") })

	got := Load()
	if got.Region != "" {
		t.Errorf("expected empty region, got %q", got.Region)
	}
}

// unsetHOME clears the HOME env var so os.UserConfigDir() fails on darwin/linux.
// It returns a cleanup function that restores the original value.
func unsetHOME(t *testing.T) {
	t.Helper()
	orig := os.Getenv("HOME")
	t.Setenv("HOME", "")
	if err := os.Unsetenv("HOME"); err != nil {
		t.Fatalf("failed to unset HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("HOME", orig); err != nil {
			t.Errorf("failed to restore HOME: %v", err)
		}
	})
}

func TestLoadReturnsZeroWhenDefaultDirEmpty(t *testing.T) {
	// Ensure overrideDir is empty so defaultDir falls through to UserConfigDir
	SetDir("")
	t.Cleanup(func() { SetDir("") })

	unsetHOME(t)

	got := Load()
	if got.Region != "" {
		t.Errorf("expected empty region when defaultDir is empty, got %q", got.Region)
	}
}

func TestSaveReturnsNilWhenDefaultDirEmpty(t *testing.T) {
	SetDir("")
	t.Cleanup(func() { SetDir("") })

	unsetHOME(t)

	err := Save(Prefs{Region: "us-west-2"})
	if err != nil {
		t.Errorf("expected nil error when defaultDir is empty, got %v", err)
	}
}

func TestSaveMkdirAllError(t *testing.T) {
	// Point dir under a file (not a directory) to trigger MkdirAll error.
	tmp := t.TempDir()
	blockingFile := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blockingFile, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	s := &store{dir: filepath.Join(blockingFile, "subdir")}
	if err := s.save(Prefs{Region: "us-west-2"}); err == nil {
		t.Error("expected error when dir is under a file")
	}
}

func TestDefaultDirUsesUserConfigDir(t *testing.T) {
	// Ensure overrideDir is cleared so we exercise the real os.UserConfigDir path.
	SetDir("")
	t.Cleanup(func() { SetDir("") })

	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME not set, cannot test UserConfigDir path")
	}

	got := Load()
	// Just verify it doesn't panic and returns a valid Prefs (even if empty).
	_ = got
}

func TestSaveJsonMarshalError(t *testing.T) {
	orig := jsonMarshal
	defer func() { jsonMarshal = orig }()

	jsonMarshal = func(v any) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	}

	s := newTestStore(t)
	err := s.save(Prefs{Region: "us-east-1"})
	if err == nil {
		t.Error("expected error from json marshal")
	}
	if err.Error() != "marshal error" {
		t.Errorf("unexpected error: %v", err)
	}
}
