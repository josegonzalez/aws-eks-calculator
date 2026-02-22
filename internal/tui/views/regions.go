package views

import (
	"fmt"
	"strings"

	"github.com/josegonzalez/aws-eks-calculator/internal/tui/styles"
)

// RenderRegions renders the region picker overlay.
func RenderRegions(regions []string, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Select Region"))
	b.WriteString("\n\n")

	for i, r := range regions {
		line := fmt.Sprintf("%-20s", r)
		if i == cursor {
			b.WriteString("  " + styles.SelectedPresetStyle.Render(line))
		} else {
			b.WriteString("  " + styles.NormalPresetStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return styles.BoxStyle.Render(b.String())
}
