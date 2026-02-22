package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
)

// osCreateFile abstracts os.Create for testing. Returns an io.WriteCloser.
var osCreateFile = func(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

// Scenario pairs an input with its calculated breakdown.
type Scenario struct {
	Input     calculator.ScenarioInput
	Breakdown calculator.CostBreakdown
}

// ToCSV writes the scenarios to a CSV file at the given path.
func ToCSV(scenarios []Scenario, path string) (retErr error) {
	f, err := osCreateFile(path)
	if err != nil {
		return fmt.Errorf("creating csv file: %w", err)
	}
	defer func() {
		if cErr := f.Close(); cErr != nil && retErr == nil {
			retErr = cErr
		}
	}()

	return writeCSV(f, scenarios)
}

func writeCSV(w io.Writer, scenarios []Scenario) error {
	cw := csv.NewWriter(w)

	// csv.Writer buffers writes internally; errors surface via Flush/Error.
	cw.Write([]string{"scenario", "capability", "metric", "value"}) //nolint:errcheck // errors checked via cw.Error()

	for _, s := range scenarios {
		cap := s.Input.Capability.String()
		rows := [][]string{
			{s.Input.Name, cap, "clusters", fmt.Sprintf("%d", s.Input.NumClusters)},
			{s.Input.Name, cap, "resources_per_cluster", fmt.Sprintf("%d", s.Input.ResourcesPerCluster)},
			{s.Input.Name, cap, "total_resources", fmt.Sprintf("%d", s.Breakdown.TotalResources)},
			{s.Input.Name, cap, "hours_per_month", fmt.Sprintf("%.0f", s.Input.HoursPerMonth)},
			{s.Input.Name, cap, "base_monthly", fmt.Sprintf("%.2f", s.Breakdown.BaseCapabilityMonthly)},
			{s.Input.Name, cap, "per_resource_monthly", fmt.Sprintf("%.2f", s.Breakdown.PerResourceMonthly)},
			{s.Input.Name, cap, "capability_subtotal_monthly", fmt.Sprintf("%.2f", s.Breakdown.CapabilitySubtotalMonthly)},
			{s.Input.Name, cap, "total_monthly", fmt.Sprintf("%.2f", s.Breakdown.TotalMonthly)},
			{s.Input.Name, cap, "total_annual", fmt.Sprintf("%.2f", s.Breakdown.TotalAnnual)},
			{s.Input.Name, cap, "self_managed_monthly", fmt.Sprintf("%.2f", s.Breakdown.SelfManagedTotalMonthly)},
			{s.Input.Name, cap, "difference_monthly", fmt.Sprintf("%.2f", s.Breakdown.ManagedVsSelfManaged)},
		}
		for _, row := range rows {
			cw.Write(row) //nolint:errcheck // errors checked via cw.Error()
		}
	}

	cw.Flush()
	return cw.Error()
}
