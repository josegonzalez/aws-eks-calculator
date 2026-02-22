package calculator

import "testing"

func TestDefaultInput(t *testing.T) {
	d := DefaultInput(CapabilityArgoCD)

	if d.Name != "Custom" {
		t.Errorf("Name: got %q, want %q", d.Name, "Custom")
	}
	if d.Capability != CapabilityArgoCD {
		t.Errorf("Capability: got %d, want %d", d.Capability, CapabilityArgoCD)
	}
	if d.NumClusters != 1 {
		t.Errorf("NumClusters: got %d, want 1", d.NumClusters)
	}
	if d.ResourcesPerCluster != 5 {
		t.Errorf("ResourcesPerCluster: got %d, want 5", d.ResourcesPerCluster)
	}
	if d.HoursPerMonth != DefaultHoursPerMonth {
		t.Errorf("HoursPerMonth: got %f, want %f", d.HoursPerMonth, DefaultHoursPerMonth)
	}
	if d.SelfManagedVCPUPerCluster != 1.0 {
		t.Errorf("SelfManagedVCPUPerCluster: got %f, want 1.0", d.SelfManagedVCPUPerCluster)
	}
	if d.SelfManagedMemGBPerCluster != 2.0 {
		t.Errorf("SelfManagedMemGBPerCluster: got %f, want 2.0", d.SelfManagedMemGBPerCluster)
	}
	if d.SelfManagedVCPUCostPerHour != 0.04048 {
		t.Errorf("SelfManagedVCPUCostPerHour: got %f, want 0.04048", d.SelfManagedVCPUCostPerHour)
	}
	if d.SelfManagedMemGBCostPerHour != 0.004446 {
		t.Errorf("SelfManagedMemGBCostPerHour: got %f, want 0.004446", d.SelfManagedMemGBCostPerHour)
	}
	if d.Region != "us-east-1" {
		t.Errorf("Region: got %q, want %q", d.Region, "us-east-1")
	}
	if d.BasePerHour != 0 {
		t.Errorf("BasePerHour: got %f, want 0 (populated at runtime)", d.BasePerHour)
	}
	if d.ResourcePerHour != 0 {
		t.Errorf("ResourcePerHour: got %f, want 0 (populated at runtime)", d.ResourcePerHour)
	}
}

func TestDefaultInputACK(t *testing.T) {
	d := DefaultInput(CapabilityACK)
	if d.Capability != CapabilityACK {
		t.Errorf("Capability: got %d, want %d", d.Capability, CapabilityACK)
	}
}

func TestDefaultInputKro(t *testing.T) {
	d := DefaultInput(CapabilityKro)
	if d.Capability != CapabilityKro {
		t.Errorf("Capability: got %d, want %d", d.Capability, CapabilityKro)
	}
}

func TestCapabilityString(t *testing.T) {
	tests := []struct {
		cap  Capability
		want string
	}{
		{CapabilityArgoCD, "ArgoCD"},
		{CapabilityACK, "ACK"},
		{CapabilityKro, "kro"},
		{Capability(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.cap.String(); got != tt.want {
			t.Errorf("Capability(%d).String(): got %q, want %q", tt.cap, got, tt.want)
		}
	}
}

func TestAllCapabilities(t *testing.T) {
	if len(AllCapabilities) != 3 {
		t.Errorf("expected 3 capabilities, got %d", len(AllCapabilities))
	}
	if AllCapabilities[0] != CapabilityArgoCD {
		t.Error("first capability should be ArgoCD")
	}
	if AllCapabilities[1] != CapabilityACK {
		t.Error("second capability should be ACK")
	}
	if AllCapabilities[2] != CapabilityKro {
		t.Error("third capability should be kro")
	}
}
