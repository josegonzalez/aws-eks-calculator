package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
)

func testScenario() Scenario {
	input := calculator.ScenarioInput{
		Name:                "Test",
		Capability:          calculator.CapabilityArgoCD,
		NumClusters:         1,
		ResourcesPerCluster: 5,
		HoursPerMonth:       730,
	}
	return Scenario{
		Input:     input,
		Breakdown: calculator.Calculate(input),
	}
}

func TestToCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	scenarios := []Scenario{testScenario()}
	if err := ToCSV(scenarios, path); err != nil {
		t.Fatalf("ToCSV: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading csv: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "scenario,capability,metric,value") {
		t.Error("missing CSV header")
	}
	if !strings.Contains(content, "Test,ArgoCD,clusters,1") {
		t.Error("missing clusters row")
	}
	if !strings.Contains(content, "Test,ArgoCD,resources_per_cluster,5") {
		t.Error("missing resources_per_cluster row")
	}
	if !strings.Contains(content, "Test,ArgoCD,total_resources,5") {
		t.Error("missing total_resources row")
	}
	if !strings.Contains(content, "Test,ArgoCD,total_monthly,") {
		t.Error("missing total_monthly row")
	}
	if !strings.Contains(content, "Test,ArgoCD,base_monthly,") {
		t.Error("missing base_monthly row")
	}
}

func TestToCSVMultipleScenarios(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.csv")

	s1 := testScenario()
	s2 := testScenario()
	s2.Input.Name = "Second"
	s2.Input.NumClusters = 3

	if err := ToCSV([]Scenario{s1, s2}, path); err != nil {
		t.Fatalf("ToCSV: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading csv: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Test,") {
		t.Error("missing first scenario")
	}
	if !strings.Contains(content, "Second,") {
		t.Error("missing second scenario")
	}
}

func TestToCSVInvalidPath(t *testing.T) {
	err := ToCSV([]Scenario{testScenario()}, "/nonexistent/dir/file.csv")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestToCSVACK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ack.csv")

	input := calculator.ScenarioInput{
		Name:                "ACK Test",
		Capability:          calculator.CapabilityACK,
		NumClusters:         2,
		ResourcesPerCluster: 10,
		HoursPerMonth:       730,
	}
	s := Scenario{Input: input, Breakdown: calculator.Calculate(input)}

	if err := ToCSV([]Scenario{s}, path); err != nil {
		t.Fatalf("ToCSV: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading csv: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "ACK Test,ACK,") {
		t.Error("missing ACK capability in CSV")
	}
}
