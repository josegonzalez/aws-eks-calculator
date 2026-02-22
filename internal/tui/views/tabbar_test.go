package views

import (
	"strings"
	"testing"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
)

func TestRenderTabBar(t *testing.T) {
	output := RenderTabBar(calculator.CapabilityArgoCD)

	if !strings.Contains(output, "ArgoCD") {
		t.Error("missing ArgoCD tab")
	}
	if !strings.Contains(output, "ACK") {
		t.Error("missing ACK tab")
	}
	if !strings.Contains(output, "kro") {
		t.Error("missing kro tab")
	}
}

func TestRenderTabBarACK(t *testing.T) {
	output := RenderTabBar(calculator.CapabilityACK)

	if !strings.Contains(output, "ACK") {
		t.Error("missing ACK tab")
	}
}
