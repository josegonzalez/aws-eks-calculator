package views

import (
	"strings"
	"testing"
)

func TestRenderHelp(t *testing.T) {
	output := RenderHelp()

	if !strings.Contains(output, "Keyboard Shortcuts") {
		t.Error("missing title")
	}
	if !strings.Contains(output, "Navigate between input fields") {
		t.Error("missing navigation help")
	}
	if !strings.Contains(output, "Previous / next capability") {
		t.Error("missing capability switching help")
	}
	if !strings.Contains(output, "Quit") {
		t.Error("missing quit help")
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"abc", 5, "abc  "},
		{"abc", 3, "abc"},
		{"abc", 2, "abc"},
		{"", 3, "   "},
	}
	for _, tt := range tests {
		got := padRight(tt.input, tt.n)
		if got != tt.want {
			t.Errorf("padRight(%q, %d): got %q, want %q", tt.input, tt.n, got, tt.want)
		}
	}
}
