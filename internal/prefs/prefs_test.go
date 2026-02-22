package prefs

import (
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
