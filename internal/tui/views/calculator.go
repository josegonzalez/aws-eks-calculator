package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
	"github.com/josegonzalez/aws-eks-calculator/internal/tui/styles"
)

// InputField defines a label/hint pair for an input field.
type InputField struct {
	Label string
	Hint  string
}

// InputFieldsForCapability returns the input field definitions for a capability.
// ArgoCD has 9 inputs (including AppTemplates/ClustersPerTemplate),
// ACK and kro have 7 inputs.
func InputFieldsForCapability(cap calculator.Capability) []InputField {
	base := []InputField{
		{"Clusters", "Number of EKS clusters with the capability enabled. Each cluster incurs a base fee."},
	}

	switch cap {
	case calculator.CapabilityArgoCD:
		base = append(base,
			InputField{"Apps/cluster", "ArgoCD Applications deployed per cluster. Each app is billed separately."},
		)
	case calculator.CapabilityACK:
		base = append(base,
			InputField{"Resources/cluster", "Managed AWS resources per cluster. Each resource is billed separately."},
		)
	case calculator.CapabilityKro:
		base = append(base,
			InputField{"RGDs/cluster", "ResourceGraphDefinition instances per cluster. Each RGD is billed separately."},
		)
	}

	base = append(base,
		InputField{"Hours/month", "Billing hours per month. AWS default is 730 (365.25 days x 24h / 12)."},
	)

	if cap == calculator.CapabilityArgoCD {
		base = append(base,
			InputField{"App templates", "Number of ApplicationSet templates. Each generates one Application per target cluster."},
			InputField{"Clusters/tmpl", "Target clusters per ApplicationSet template. Total generated apps = templates x this value."},
		)
	}

	base = append(base,
		InputField{"vCPU/cluster", "vCPU allocated to pods per cluster when self-managing."},
		InputField{"Memory GB", "Memory (GB) allocated to pods per cluster when self-managing."},
		InputField{"vCPU $/hr", "Fargate vCPU cost per hour. Fetched from AWS Pricing API; override for custom pricing."},
		InputField{"Mem GB $/hr", "Fargate memory cost per GB-hour. Fetched from AWS Pricing API; override for custom pricing."},
	)

	return base
}

// InputLabelsForCapability returns just the labels for a capability.
func InputLabelsForCapability(cap calculator.Capability) []string {
	fields := InputFieldsForCapability(cap)
	labels := make([]string, len(fields))
	for i, f := range fields {
		labels[i] = f.Label
	}
	return labels
}

// InputHintsForCapability returns just the hints for a capability.
func InputHintsForCapability(cap calculator.Capability) []string {
	fields := InputFieldsForCapability(cap)
	hints := make([]string, len(fields))
	for i, f := range fields {
		hints[i] = f.Hint
	}
	return hints
}

// ResourceLabel returns the per-resource label for the given capability.
func ResourceLabel(cap calculator.Capability) string {
	switch cap {
	case calculator.CapabilityArgoCD:
		return "Per-application"
	case calculator.CapabilityACK:
		return "Per-resource"
	case calculator.CapabilityKro:
		return "Per-RGD"
	default:
		return "Per-resource"
	}
}

// TotalResourcesLabel returns the total resources label for the given capability.
func TotalResourcesLabel(cap calculator.Capability) string {
	switch cap {
	case calculator.CapabilityArgoCD:
		return "Total apps:"
	case calculator.CapabilityACK:
		return "Total resources:"
	case calculator.CapabilityKro:
		return "Total RGDs:"
	default:
		return "Total resources:"
	}
}

// RenderCalculator renders the main calculator view with inputs on the left
// and cost breakdown on the right.
func RenderCalculator(cap calculator.Capability, inputs []textinput.Model, focusIndex int, input calculator.ScenarioInput, breakdown calculator.CostBreakdown, width, height int) string {
	leftWidth := 32
	rightWidth := width - leftWidth - 5
	if rightWidth < 40 {
		rightWidth = 40
	}

	left := renderInputPanel(cap, inputs, focusIndex, breakdown, leftWidth, input.Region)
	right := renderBreakdownPanel(cap, input, breakdown, rightWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

func renderInputPanel(cap calculator.Capability, inputs []textinput.Model, focusIndex int, breakdown calculator.CostBreakdown, width int, region string) string {
	var b strings.Builder
	labels := InputLabelsForCapability(cap)

	b.WriteString(styles.SectionStyle.Render("EKS-MANAGED COSTS"))
	b.WriteString("\n\n")

	b.WriteString(styles.SubSectionStyle.Render("  Deployment Config"))
	b.WriteString("\n")

	// Main inputs: clusters, resources/cluster, hours (indices 0-2)
	for i := 0; i < 3 && i < len(inputs); i++ {
		renderInput(&b, labels[i], inputs[i], i == focusIndex)
	}
	b.WriteString("\n")

	inputIdx := 3

	// ApplicationSets section (ArgoCD only)
	if cap == calculator.CapabilityArgoCD {
		b.WriteString(styles.SubSectionStyle.Render("  ApplicationSets"))
		b.WriteString("\n")
		for i := 3; i < 5 && i < len(inputs); i++ {
			renderInput(&b, labels[i], inputs[i], i == focusIndex)
		}
		inputIdx = 5
	}

	// Total resources summary
	fmt.Fprintf(&b, "  %s %s\n\n",
		styles.LabelStyle.Render(TotalResourcesLabel(cap)),
		styles.ValueStyle.Render(fmt.Sprintf("%d", breakdown.TotalResources)),
	)

	// Self-managed section
	b.WriteString(styles.SectionStyle.Render("SELF-MANAGED COSTS"))
	b.WriteString("\n\n")
	for i := inputIdx; i < len(inputs); i++ {
		renderInput(&b, labels[i], inputs[i], i == focusIndex)
	}

	// Pricing region label
	b.WriteString("\n")
	b.WriteString(styles.SectionStyle.Render("PRICING REGION"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "  %s  %s\n",
		styles.LabelStyle.Render("Region:"),
		styles.ValueStyle.Render(region+"  ")+styles.MutedStyle.Render("(r to change)"),
	)

	return b.String()
}

func renderInput(b *strings.Builder, label string, input textinput.Model, focused bool) {
	style := styles.BlurredInputStyle
	if focused {
		style = styles.FocusedInputStyle
	}
	fmt.Fprintf(b, "  %s %s\n",
		style.Render(fmt.Sprintf("%-17s", label+":")),
		input.View(),
	)
}

func renderBreakdownPanel(cap calculator.Capability, input calculator.ScenarioInput, breakdown calculator.CostBreakdown, width int) string {
	var b strings.Builder

	b.WriteString(styles.SectionStyle.Render("EKS-MANAGED COST BREAKDOWN"))
	b.WriteString("\n\n")

	// Base capability
	fmt.Fprintf(&b, "  %s  %s\n",
		styles.LabelStyle.Render("Base capability"),
		styles.MoneyStyle.Render(formatMoney(breakdown.BaseCapabilityMonthly)+"/mo"),
	)
	fmt.Fprintf(&b, "  %s\n",
		styles.MutedStyle.Render(fmt.Sprintf("$%.6f/hr x %.0fh x %d clusters",
			input.BasePerHour,
			input.HoursPerMonth,
			input.NumClusters)),
	)

	// Per-resource
	resLabel := ResourceLabel(cap)
	fmt.Fprintf(&b, "  %s  %s\n",
		styles.LabelStyle.Render(resLabel),
		styles.MoneyStyle.Render(formatMoney(breakdown.PerResourceMonthly)+"/mo"),
	)
	fmt.Fprintf(&b, "  %s\n",
		styles.MutedStyle.Render(fmt.Sprintf("$%.6f/hr x %d x %.0fh",
			input.ResourcePerHour,
			breakdown.TotalResources,
			input.HoursPerMonth)),
	)

	// Totals
	b.WriteString(styles.LabelStyle.Render(strings.Repeat("─", 36)))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s  %s\n",
		styles.LabelStyle.Render("MONTHLY TOTAL  "),
		styles.BigMoneyStyle.Render(formatMoney(breakdown.TotalMonthly)),
	)
	fmt.Fprintf(&b, "  %s  %s\n\n",
		styles.LabelStyle.Render("ANNUAL TOTAL   "),
		styles.BigMoneyStyle.Render(formatMoney(breakdown.TotalAnnual)),
	)

	// Self-managed comparison
	b.WriteString(styles.SectionStyle.Render("SELF-MANAGED COST BREAKDOWN"))
	b.WriteString("\n\n")

	fmt.Fprintf(&b, "  %s  %s\n",
		styles.LabelStyle.Render("Compute        "),
		styles.MoneyStyle.Render(formatMoney(breakdown.SelfManagedComputeMonthly)+"/mo"),
	)
	fmt.Fprintf(&b, "  %s\n",
		styles.MutedStyle.Render(fmt.Sprintf("(%.1f vCPU x $%.6f + %.1fGB x $%.6f)/hr",
			input.SelfManagedVCPUPerCluster, input.SelfManagedVCPUCostPerHour,
			input.SelfManagedMemGBPerCluster, input.SelfManagedMemGBCostPerHour)),
	)
	fmt.Fprintf(&b, "  %s\n",
		styles.MutedStyle.Render(fmt.Sprintf("x %.0fh x %d clusters",
			input.HoursPerMonth, input.NumClusters)),
	)

	b.WriteString(styles.LabelStyle.Render(strings.Repeat("─", 36)))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s  %s\n",
		styles.LabelStyle.Render("MONTHLY TOTAL  "),
		styles.BigMoneyStyle.Render(formatMoney(breakdown.SelfManagedTotalMonthly)),
	)
	fmt.Fprintf(&b, "  %s  %s\n\n",
		styles.LabelStyle.Render("ANNUAL TOTAL   "),
		styles.BigMoneyStyle.Render(formatMoney(breakdown.SelfManagedTotalAnnual)),
	)

	// Difference
	b.WriteString(styles.SectionStyle.Render("DIFFERENCE"))
	b.WriteString("\n\n")

	diff := breakdown.ManagedVsSelfManaged
	diffStyle := styles.SuccessStyle
	diffLabel := "(AWS managed saves)"
	if diff > 0 {
		diffStyle = styles.ErrorStyle
		diffLabel = "(AWS managed costs more)"
	} else if diff == 0 {
		diffStyle = styles.MutedStyle
		diffLabel = "(same cost)"
	}
	fmt.Fprintf(&b, "  %s  %s\n",
		styles.LabelStyle.Render("Monthly        "),
		diffStyle.Render(formatMoneyWithSign(diff)+"/mo"),
	)
	fmt.Fprintf(&b, "  %s  %s\n",
		styles.LabelStyle.Render("Annual         "),
		diffStyle.Render(formatMoneyWithSign(diff*12)+"/yr"),
	)
	fmt.Fprintf(&b, "  %s\n",
		styles.MutedStyle.Render(diffLabel),
	)

	return b.String()
}

func formatMoney(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	// Add comma separators for thousands
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	if len(intPart) > 3 {
		var result []byte
		for i, c := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		intPart = string(result)
	}
	return "$" + intPart + "." + parts[1]
}

func formatMoneyWithSign(v float64) string {
	if v > 0 {
		return fmt.Sprintf("+$%.2f", v)
	} else if v < 0 {
		return fmt.Sprintf("-$%.2f", -v)
	}
	return "$0.00"
}
