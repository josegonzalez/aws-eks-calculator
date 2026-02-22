package calculator

import (
	"math"
	"testing"
)

const tolerance = 0.01

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestCalculateBasic(t *testing.T) {
	// 1 cluster, 1 app, 730 hours
	// Base: 0.02771 * 730 * 1 = 20.2283
	// Apps: 0.00136 * 1 * 730 = 0.9928
	// Total: 21.2211
	input := ScenarioInput{
		Capability:                  CapabilityArgoCD,
		NumClusters:                 1,
		ResourcesPerCluster:         1,
		HoursPerMonth:               730,
		BasePerHour:                 0.02771,
		ResourcePerHour:             0.00136,
		SelfManagedVCPUPerCluster:   0.5,
		SelfManagedMemGBPerCluster:  1.0,
		SelfManagedVCPUCostPerHour:  0.04048,
		SelfManagedMemGBCostPerHour: 0.004446,
	}

	result := Calculate(input)

	if result.TotalResources != 1 {
		t.Errorf("TotalResources: got %d, want 1", result.TotalResources)
	}
	if !almostEqual(result.BaseCapabilityMonthly, 20.23) {
		t.Errorf("BaseCapabilityMonthly: got %.2f, want 20.23", result.BaseCapabilityMonthly)
	}
	if !almostEqual(result.PerResourceMonthly, 0.99) {
		t.Errorf("PerResourceMonthly: got %.2f, want 0.99", result.PerResourceMonthly)
	}
	if !almostEqual(result.TotalMonthly, 21.22) {
		t.Errorf("TotalMonthly: got %.2f, want 21.22", result.TotalMonthly)
	}
	if !almostEqual(result.TotalAnnual, result.TotalMonthly*12) {
		t.Errorf("TotalAnnual: got %.2f, want %.2f", result.TotalAnnual, result.TotalMonthly*12)
	}
}

func TestCalculateZeroClusters(t *testing.T) {
	input := ScenarioInput{
		Capability:          CapabilityArgoCD,
		NumClusters:         0,
		ResourcesPerCluster: 10,
		HoursPerMonth:       730,
		BasePerHour:         0.02771,
		ResourcePerHour:     0.00136,
	}

	result := Calculate(input)

	if result.TotalResources != 0 {
		t.Errorf("TotalResources: got %d, want 0", result.TotalResources)
	}
	if result.TotalMonthly != 0 {
		t.Errorf("TotalMonthly: got %.2f, want 0", result.TotalMonthly)
	}
}

func TestCalculateApplicationSets(t *testing.T) {
	// 2 clusters, 5 apps/cluster = 10 direct apps
	// 3 templates x 2 clusters/template = 6 appset apps
	// Total: 16 apps
	input := ScenarioInput{
		Capability:          CapabilityArgoCD,
		NumClusters:         2,
		ResourcesPerCluster: 5,
		HoursPerMonth:       730,
		BasePerHour:         0.02771,
		ResourcePerHour:     0.00136,
		AppTemplates:        3,
		ClustersPerTemplate: 2,
	}

	result := Calculate(input)

	if result.TotalResources != 16 {
		t.Errorf("TotalResources: got %d, want 16", result.TotalResources)
	}

	// Apps: 0.00136 * 16 * 730 = 15.8848
	if !almostEqual(result.PerResourceMonthly, 15.88) {
		t.Errorf("PerResourceMonthly: got %.2f, want 15.88", result.PerResourceMonthly)
	}
}

func TestCalculateApplicationSetsIgnoredForACK(t *testing.T) {
	input := ScenarioInput{
		Capability:          CapabilityACK,
		NumClusters:         2,
		ResourcesPerCluster: 5,
		HoursPerMonth:       730,
		BasePerHour:         0.02771,
		ResourcePerHour:     0.00136,
		AppTemplates:        3,
		ClustersPerTemplate: 2,
	}

	result := Calculate(input)

	// ACK should ignore AppTemplates/ClustersPerTemplate
	if result.TotalResources != 10 {
		t.Errorf("TotalResources: got %d, want 10 (ACK ignores appsets)", result.TotalResources)
	}
}

func TestCalculateApplicationSetsIgnoredForKro(t *testing.T) {
	input := ScenarioInput{
		Capability:          CapabilityKro,
		NumClusters:         2,
		ResourcesPerCluster: 5,
		HoursPerMonth:       730,
		BasePerHour:         0.02771,
		ResourcePerHour:     0.00136,
		AppTemplates:        3,
		ClustersPerTemplate: 2,
	}

	result := Calculate(input)

	// kro should ignore AppTemplates/ClustersPerTemplate
	if result.TotalResources != 10 {
		t.Errorf("TotalResources: got %d, want 10 (kro ignores appsets)", result.TotalResources)
	}
}

func TestCalculateMultipleClusters(t *testing.T) {
	// 3 clusters, 10 apps/cluster
	// Base: 0.02771 * 730 * 3 = 60.6849
	// Apps: 0.00136 * 30 * 730 = 29.784
	input := ScenarioInput{
		Capability:          CapabilityArgoCD,
		NumClusters:         3,
		ResourcesPerCluster: 10,
		HoursPerMonth:       730,
		BasePerHour:         0.02771,
		ResourcePerHour:     0.00136,
	}

	result := Calculate(input)

	if result.TotalResources != 30 {
		t.Errorf("TotalResources: got %d, want 30", result.TotalResources)
	}
	if !almostEqual(result.BaseCapabilityMonthly, 60.68) {
		t.Errorf("BaseCapabilityMonthly: got %.2f, want 60.68", result.BaseCapabilityMonthly)
	}
	if !almostEqual(result.PerResourceMonthly, 29.78) {
		t.Errorf("PerResourceMonthly: got %.2f, want 29.78", result.PerResourceMonthly)
	}
}

func TestCalculateSelfManaged(t *testing.T) {
	// 1 cluster, 0.5 vCPU at $0.04048/hr + 1.0 GB at $0.004446/hr (Fargate us-east-1)
	// Compute: (0.5 * 0.04048 + 1.0 * 0.004446) * 730 * 1 = 0.024686 * 730 = 18.02
	input := ScenarioInput{
		Capability:                  CapabilityArgoCD,
		NumClusters:                 1,
		ResourcesPerCluster:         5,
		HoursPerMonth:               730,
		BasePerHour:                 0.02771,
		ResourcePerHour:             0.00136,
		SelfManagedVCPUPerCluster:   0.5,
		SelfManagedMemGBPerCluster:  1.0,
		SelfManagedVCPUCostPerHour:  0.04048,
		SelfManagedMemGBCostPerHour: 0.004446,
	}

	result := Calculate(input)

	if !almostEqual(result.SelfManagedComputeMonthly, 18.02) {
		t.Errorf("SelfManagedComputeMonthly: got %.2f, want 18.02", result.SelfManagedComputeMonthly)
	}
	if !almostEqual(result.SelfManagedTotalMonthly, 18.02) {
		t.Errorf("SelfManagedTotalMonthly: got %.2f, want 18.02", result.SelfManagedTotalMonthly)
	}

	// Managed vs self-managed: positive = managed costs more
	if result.ManagedVsSelfManaged <= 0 {
		t.Error("expected managed to cost more than self-managed compute alone")
	}
}

func TestCalculateDefaultHoursWhenZero(t *testing.T) {
	input := ScenarioInput{
		Capability:          CapabilityArgoCD,
		NumClusters:         1,
		ResourcesPerCluster: 1,
		HoursPerMonth:       0, // should default to 730
		BasePerHour:         0.02771,
		ResourcePerHour:     0.00136,
	}

	result := Calculate(input)

	// Should use default hours
	expected := Calculate(ScenarioInput{
		Capability:          CapabilityArgoCD,
		NumClusters:         1,
		ResourcesPerCluster: 1,
		HoursPerMonth:       DefaultHoursPerMonth,
		BasePerHour:         0.02771,
		ResourcePerHour:     0.00136,
	})

	if !almostEqual(result.TotalMonthly, expected.TotalMonthly) {
		t.Errorf("TotalMonthly with 0 hours: got %.2f, want %.2f", result.TotalMonthly, expected.TotalMonthly)
	}
}

func TestCalculateZeroRates(t *testing.T) {
	input := ScenarioInput{
		Capability:          CapabilityArgoCD,
		NumClusters:         1,
		ResourcesPerCluster: 5,
		HoursPerMonth:       730,
	}

	result := Calculate(input)

	if result.BaseCapabilityMonthly != 0 {
		t.Errorf("BaseCapabilityMonthly: got %.2f, want 0", result.BaseCapabilityMonthly)
	}
	if result.PerResourceMonthly != 0 {
		t.Errorf("PerResourceMonthly: got %.2f, want 0", result.PerResourceMonthly)
	}
	if result.TotalMonthly != 0 {
		t.Errorf("TotalMonthly: got %.2f, want 0", result.TotalMonthly)
	}
}
