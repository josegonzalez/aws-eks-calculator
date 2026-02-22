package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
)

func makeTestInputs(n int) []textinput.Model {
	values9 := []string{"3", "10", "730", "0", "0", "1.0", "2.0", "0.0405", "0.0044"}
	values7 := []string{"3", "10", "730", "1.0", "2.0", "0.0405", "0.0044"}
	var values []string
	if n == 9 {
		values = values9
	} else {
		values = values7
	}
	inputs := make([]textinput.Model, n)
	for i, v := range values {
		inputs[i] = textinput.New()
		inputs[i].SetValue(v)
	}
	return inputs
}

func TestRenderCalculatorArgoCD(t *testing.T) {
	inputs := makeTestInputs(9)
	input := calculator.ScenarioInput{
		Capability:                  calculator.CapabilityArgoCD,
		NumClusters:                 3,
		ResourcesPerCluster:         10,
		HoursPerMonth:               730,
		SelfManagedVCPUPerCluster:   1.0,
		SelfManagedMemGBPerCluster:  2.0,
		SelfManagedVCPUCostPerHour:  0.04048,
		SelfManagedMemGBCostPerHour: 0.004446,
	}
	breakdown := calculator.Calculate(input)

	output := RenderCalculator(calculator.CapabilityArgoCD, inputs, 0, input, breakdown, 120, 40)

	if !strings.Contains(output, "EKS-MANAGED COSTS") {
		t.Error("missing input panel header")
	}
	if !strings.Contains(output, "Deployment Config") {
		t.Error("missing deployment config sub-section")
	}
	if !strings.Contains(output, "EKS-MANAGED COST BREAKDOWN") {
		t.Error("missing breakdown panel header")
	}
	if !strings.Contains(output, "ApplicationSets") {
		t.Error("missing ApplicationSets sub-section for ArgoCD")
	}
	if !strings.Contains(output, "SELF-MANAGED COSTS") {
		t.Error("missing self-managed section")
	}
	if !strings.Contains(output, "Per-application") {
		t.Error("missing Per-application label for ArgoCD")
	}
	if !strings.Contains(output, "DIFFERENCE") {
		t.Error("missing difference section")
	}
	if !strings.Contains(output, "/yr") {
		t.Error("missing annual difference")
	}
}

func TestRenderCalculatorACK(t *testing.T) {
	inputs := makeTestInputs(7)
	input := calculator.ScenarioInput{
		Capability:          calculator.CapabilityACK,
		NumClusters:         3,
		ResourcesPerCluster: 10,
		HoursPerMonth:       730,
	}
	breakdown := calculator.Calculate(input)

	output := RenderCalculator(calculator.CapabilityACK, inputs, 0, input, breakdown, 120, 40)

	if strings.Contains(output, "ApplicationSets") {
		t.Error("ACK should NOT have ApplicationSets section")
	}
	if !strings.Contains(output, "Per-resource") {
		t.Error("missing Per-resource label for ACK")
	}
	if !strings.Contains(output, "Total resources:") {
		t.Error("missing Total resources label for ACK")
	}
}

func TestRenderCalculatorKro(t *testing.T) {
	inputs := makeTestInputs(7)
	input := calculator.ScenarioInput{
		Capability:          calculator.CapabilityKro,
		NumClusters:         3,
		ResourcesPerCluster: 10,
		HoursPerMonth:       730,
	}
	breakdown := calculator.Calculate(input)

	output := RenderCalculator(calculator.CapabilityKro, inputs, 0, input, breakdown, 120, 40)

	if strings.Contains(output, "ApplicationSets") {
		t.Error("kro should NOT have ApplicationSets section")
	}
	if !strings.Contains(output, "Per-RGD") {
		t.Error("missing Per-RGD label for kro")
	}
	if !strings.Contains(output, "Total RGDs:") {
		t.Error("missing Total RGDs label for kro")
	}
}

func TestRenderCalculatorNarrowWidth(t *testing.T) {
	inputs := makeTestInputs(9)
	input := calculator.ScenarioInput{
		Capability:          calculator.CapabilityArgoCD,
		NumClusters:         1,
		ResourcesPerCluster: 5,
		HoursPerMonth:       730,
	}
	breakdown := calculator.Calculate(input)

	// Width too narrow for right panel
	output := RenderCalculator(calculator.CapabilityArgoCD, inputs, 0, input, breakdown, 50, 40)
	if output == "" {
		t.Error("should still render with narrow width")
	}
}

func TestRenderBreakdownDiffZero(t *testing.T) {
	inputs := makeTestInputs(9)
	input := calculator.ScenarioInput{Capability: calculator.CapabilityArgoCD, HoursPerMonth: 730, NumClusters: 1}
	breakdown := calculator.CostBreakdown{
		TotalMonthly:            100,
		SelfManagedTotalMonthly: 100,
		ManagedVsSelfManaged:    0,
	}

	output := RenderCalculator(calculator.CapabilityArgoCD, inputs, 0, input, breakdown, 120, 40)
	if !strings.Contains(output, "same cost") {
		t.Error("should show 'same cost' when difference is 0")
	}
}

func TestRenderBreakdownManagedCheaper(t *testing.T) {
	inputs := makeTestInputs(9)
	input := calculator.ScenarioInput{Capability: calculator.CapabilityArgoCD, HoursPerMonth: 730, NumClusters: 1}
	breakdown := calculator.CostBreakdown{
		TotalMonthly:            80,
		SelfManagedTotalMonthly: 100,
		ManagedVsSelfManaged:    -20,
	}

	output := RenderCalculator(calculator.CapabilityArgoCD, inputs, 0, input, breakdown, 120, 40)
	if !strings.Contains(output, "AWS managed saves") {
		t.Error("should show saves message when managed is cheaper")
	}
}

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "$0.00"},
		{1.5, "$1.50"},
		{99.99, "$99.99"},
		{1000, "$1,000.00"},
		{12345.67, "$12,345.67"},
		{1234567.89, "$1,234,567.89"},
	}
	for _, tt := range tests {
		got := formatMoney(tt.input)
		if got != tt.want {
			t.Errorf("formatMoney(%f): got %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatMoneyWithSign(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "$0.00"},
		{10.5, "+$10.50"},
		{-5.25, "-$5.25"},
	}
	for _, tt := range tests {
		got := formatMoneyWithSign(tt.input)
		if got != tt.want {
			t.Errorf("formatMoneyWithSign(%f): got %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInputFieldsForCapability(t *testing.T) {
	argoCDFields := InputFieldsForCapability(calculator.CapabilityArgoCD)
	if len(argoCDFields) != 9 {
		t.Errorf("ArgoCD: expected 9 fields, got %d", len(argoCDFields))
	}

	ackFields := InputFieldsForCapability(calculator.CapabilityACK)
	if len(ackFields) != 7 {
		t.Errorf("ACK: expected 7 fields, got %d", len(ackFields))
	}

	kroFields := InputFieldsForCapability(calculator.CapabilityKro)
	if len(kroFields) != 7 {
		t.Errorf("kro: expected 7 fields, got %d", len(kroFields))
	}
}

func TestResourceLabel(t *testing.T) {
	if ResourceLabel(calculator.CapabilityArgoCD) != "Per-application" {
		t.Error("wrong ArgoCD label")
	}
	if ResourceLabel(calculator.CapabilityACK) != "Per-resource" {
		t.Error("wrong ACK label")
	}
	if ResourceLabel(calculator.CapabilityKro) != "Per-RGD" {
		t.Error("wrong kro label")
	}
}

func TestTotalResourcesLabel(t *testing.T) {
	if TotalResourcesLabel(calculator.CapabilityArgoCD) != "Total apps:" {
		t.Error("wrong ArgoCD total label")
	}
	if TotalResourcesLabel(calculator.CapabilityACK) != "Total resources:" {
		t.Error("wrong ACK total label")
	}
	if TotalResourcesLabel(calculator.CapabilityKro) != "Total RGDs:" {
		t.Error("wrong kro total label")
	}
}
