package calculator

// Capability represents an EKS capability type.
type Capability int

const (
	CapabilityArgoCD Capability = iota
	CapabilityACK
	CapabilityKro
)

// String returns the display name for the capability.
func (c Capability) String() string {
	switch c {
	case CapabilityArgoCD:
		return "ArgoCD"
	case CapabilityACK:
		return "ACK"
	case CapabilityKro:
		return "kro"
	default:
		return "Unknown"
	}
}

// AllCapabilities returns all supported capabilities.
var AllCapabilities = []Capability{CapabilityArgoCD, CapabilityACK, CapabilityKro}

// ScenarioInput holds all user-configurable inputs for a cost scenario.
type ScenarioInput struct {
	Name        string
	Capability  Capability
	NumClusters int
	// ResourcesPerCluster is the number of billable resources per cluster
	// (Applications for ArgoCD, managed AWS resources for ACK, RGD instances for kro).
	ResourcesPerCluster int
	HoursPerMonth       float64

	// AWS region code for pricing lookup (default: "us-east-1").
	Region string

	// Capability rates (fetched from AWS Pricing API or hardcoded defaults).
	BasePerHour     float64
	ResourcePerHour float64

	// ApplicationSet expansion (ArgoCD-only): each template generates one Application per target cluster.
	AppTemplates        int
	ClustersPerTemplate int

	// Self-managed comparison inputs.
	// Default resources based on ArgoCD recommended requests for core components
	// (server: 125m/128Mi, repo-server: 250m/256Mi, application-controller: 250m/1Gi).
	// Default rates use EKS Fargate pricing for us-east-1 Linux/X86:
	// vCPU: $0.000011244/sec = $0.04048/hr, GB: $0.000001235/sec = $0.004446/hr
	SelfManagedVCPUPerCluster   float64
	SelfManagedMemGBPerCluster  float64
	SelfManagedVCPUCostPerHour  float64
	SelfManagedMemGBCostPerHour float64
}

// DefaultInput returns a ScenarioInput with sensible defaults for the given capability.
func DefaultInput(cap Capability) ScenarioInput {
	return ScenarioInput{
		Name:                        "Custom",
		Capability:                  cap,
		NumClusters:                 1,
		ResourcesPerCluster:         5,
		HoursPerMonth:               DefaultHoursPerMonth,
		Region:                      "us-east-1",
		SelfManagedVCPUPerCluster:   1.0,
		SelfManagedMemGBPerCluster:  2.0,
		SelfManagedVCPUCostPerHour:  0.04048,
		SelfManagedMemGBCostPerHour: 0.004446,
	}
}

// CostBreakdown holds the calculated cost breakdown for a scenario.
type CostBreakdown struct {
	TotalResources int

	// Capability managed service costs.
	BaseCapabilityMonthly    float64
	PerResourceMonthly       float64
	CapabilitySubtotalMonthly float64

	// Totals (managed only, assumes existing EKS clusters).
	TotalMonthly float64
	TotalAnnual  float64

	// Self-managed comparison.
	SelfManagedComputeMonthly float64 // compute cost for pods
	SelfManagedTotalMonthly   float64 // compute only (assumes existing EKS clusters)
	SelfManagedTotalAnnual    float64
	ManagedVsSelfManaged      float64 // positive means managed costs more
}
