package views

import (
	"strings"

	"github.com/josegonzalez/aws-eks-calculator/internal/tui/styles"
)

// RenderHelp renders the help overlay with all keybindings.
func RenderHelp() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	bindings := []struct{ key, desc string }{
		{"↑/↓ / tab / shift+tab", "Navigate between input fields"},
		{"[ / ]", "Previous / next capability"},
		{"r", "Open region picker"},
		{"e", "Export current scenario to CSV"},
		{"?", "Toggle this help overlay"},
		{"esc", "Close overlay / go back"},
		{"q / ctrl+c", "Quit"},
	}

	for _, bind := range bindings {
		b.WriteString("  ")
		b.WriteString(styles.FocusedInputStyle.Render(padRight(bind.key, 24)))
		b.WriteString(styles.LabelStyle.Render(bind.desc))
		b.WriteString("\n")
	}

	return styles.BoxStyle.Render(b.String())
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
