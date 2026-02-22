package views

import (
	"strings"
	"testing"
)

func TestRenderCapabilitySelector(t *testing.T) {
	output := RenderCapabilitySelector(0)

	if !strings.Contains(output, "Select EKS Capability") {
		t.Error("missing title")
	}
	if !strings.Contains(output, "ArgoCD") {
		t.Error("missing ArgoCD")
	}
	if !strings.Contains(output, "ACK") {
		t.Error("missing ACK")
	}
	if !strings.Contains(output, "kro") {
		t.Error("missing kro")
	}
	if strings.Contains(output, "navigate") {
		t.Error("hints should not be inside the selector box")
	}
}

func TestRenderCapabilitySelectorCursorMid(t *testing.T) {
	output := RenderCapabilitySelector(1)

	// Should still contain all capabilities
	if !strings.Contains(output, "ArgoCD") {
		t.Error("missing ArgoCD")
	}
	if !strings.Contains(output, "ACK") {
		t.Error("missing ACK")
	}
	if !strings.Contains(output, "kro") {
		t.Error("missing kro")
	}
}
