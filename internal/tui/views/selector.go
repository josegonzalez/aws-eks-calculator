package views

import (
	"fmt"
	"strings"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
	"github.com/josegonzalez/aws-eks-calculator/internal/tui/styles"
)

// capabilityInfo holds display metadata for a capability.
type capabilityInfo struct {
	Cap         calculator.Capability
	Description string
}

var capabilityInfos = []capabilityInfo{
	{calculator.CapabilityArgoCD, "GitOps continuous delivery — per Application/hr"},
	{calculator.CapabilityACK, "AWS Controllers for Kubernetes — per managed AWS resource/hr"},
	{calculator.CapabilityKro, "Kube Resource Orchestrator — per RGD instance/hr"},
}

// RenderCapabilitySelector renders the capability picker overlay.
func RenderCapabilitySelector(cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Select EKS Capability"))
	b.WriteString("\n\n")

	for i, info := range capabilityInfos {
		line := fmt.Sprintf("%-10s %s", info.Cap.String(), info.Description)
		if i == cursor {
			b.WriteString("  " + styles.SelectedPresetStyle.Render(line))
		} else {
			b.WriteString("  " + styles.NormalPresetStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return styles.BoxStyle.Render(b.String())
}
