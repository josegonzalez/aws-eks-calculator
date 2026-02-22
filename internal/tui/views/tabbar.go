package views

import (
	"strings"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
	"github.com/josegonzalez/aws-eks-calculator/internal/tui/styles"
)

// RenderTabBar renders a horizontal tab bar showing the active capability.
func RenderTabBar(active calculator.Capability) string {
	var b strings.Builder

	for i, cap := range calculator.AllCapabilities {
		label := cap.String()
		if cap == active {
			b.WriteString(styles.ActiveTabStyle.Render(label))
		} else {
			b.WriteString(styles.InactiveTabStyle.Render(label))
		}
		if i < len(calculator.AllCapabilities)-1 {
			b.WriteString("  ")
		}
	}

	return b.String()
}
