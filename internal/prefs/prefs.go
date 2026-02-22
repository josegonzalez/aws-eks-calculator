package prefs

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configSubdir = "aws-eks-calculator"
const configFile = "prefs.json"

var jsonMarshal = json.Marshal

// Prefs holds user preferences that persist across sessions.
type Prefs struct {
	Region string `json:"region,omitempty"`
}

// overrideDir allows tests to redirect prefs to a temporary directory.
var overrideDir string

// SetDir overrides the config directory. Pass "" to reset to the default.
// Intended for testing.
func SetDir(dir string) {
	overrideDir = dir
}

// store abstracts the config directory for testing.
type store struct {
	dir string
}

func defaultDir() string {
	if overrideDir != "" {
		return overrideDir
	}
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, configSubdir)
}

func (s *store) path() string {
	return filepath.Join(s.dir, configFile)
}

func (s *store) load() Prefs {
	data, err := os.ReadFile(s.path())
	if err != nil {
		return Prefs{}
	}

	var p Prefs
	if err := json.Unmarshal(data, &p); err != nil {
		return Prefs{}
	}

	return p
}

func (s *store) save(p Prefs) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}

	data, err := jsonMarshal(p)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path(), data, 0o600)
}

// Load reads saved preferences from the user config directory.
// Returns zero-value Prefs on any error.
func Load() Prefs {
	d := defaultDir()
	if d == "" {
		return Prefs{}
	}
	return (&store{dir: d}).load()
}

// Save writes preferences to the user config directory.
func Save(p Prefs) error {
	d := defaultDir()
	if d == "" {
		return nil
	}
	return (&store{dir: d}).save(p)
}
