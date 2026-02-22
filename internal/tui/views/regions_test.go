package views

import (
	"strings"
	"testing"
)

func TestRenderRegions(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	output := RenderRegions(regions, 0)

	if !strings.Contains(output, "Select Region") {
		t.Error("missing title")
	}
	for _, r := range regions {
		if !strings.Contains(output, r) {
			t.Errorf("missing region %q", r)
		}
	}
	if strings.Contains(output, "navigate") {
		t.Error("hints should not be inside the regions box")
	}
}

func TestRenderRegionsCursorMid(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	output := RenderRegions(regions, 1)

	// Should still contain all regions
	for _, r := range regions {
		if !strings.Contains(output, r) {
			t.Errorf("missing region %q", r)
		}
	}
}
