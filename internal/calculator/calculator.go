package calculator

// Calculate computes the full cost breakdown for a given scenario.
//
// The calculation follows this logic:
//
//  1. Total resources = (clusters x resources_per_cluster) + appset_expansion
//     For ArgoCD, appset_expansion = app_templates x clusters_per_template.
//     For ACK and kro, appset_expansion is always 0.
//
//  2. Base capability = base_rate/hr x hours_per_month x num_clusters
//     This fee is charged per cluster that has the capability enabled.
//
//  3. Per-resource = resource_rate/hr x total_resources x hours_per_month
//     Each resource instance is billed individually.
//
//  4. Self-managed comparison estimates the compute cost of running the capability yourself:
//     compute_per_cluster = (vCPU x vCPU_rate + memory_GB x memory_rate)
//     self_managed_total = compute_per_cluster x hours x clusters
//     This excludes operational overhead (upgrades, monitoring, HA setup).
//
// EKS cluster costs are excluded — both managed and self-managed assume
// existing EKS clusters.
func Calculate(input ScenarioInput) CostBreakdown {
	directResources := input.NumClusters * input.ResourcesPerCluster

	// ApplicationSet expansion only applies to ArgoCD
	appsetResources := 0
	if input.Capability == CapabilityArgoCD {
		appsetResources = input.AppTemplates * input.ClustersPerTemplate
	}
	totalResources := directResources + appsetResources

	hours := input.HoursPerMonth
	if hours <= 0 {
		hours = DefaultHoursPerMonth
	}

	// Managed service costs
	baseMonthly := input.BasePerHour * hours * float64(input.NumClusters)
	resourceMonthly := input.ResourcePerHour * float64(totalResources) * hours
	capabilitySubtotal := baseMonthly + resourceMonthly

	totalMonthly := capabilitySubtotal
	totalAnnual := totalMonthly * 12

	// Self-managed comparison
	computePerCluster := input.SelfManagedVCPUPerCluster*input.SelfManagedVCPUCostPerHour +
		input.SelfManagedMemGBPerCluster*input.SelfManagedMemGBCostPerHour
	selfManagedCompute := computePerCluster * hours * float64(input.NumClusters)
	selfManagedAnnual := selfManagedCompute * 12

	return CostBreakdown{
		TotalResources:            totalResources,
		BaseCapabilityMonthly:     baseMonthly,
		PerResourceMonthly:        resourceMonthly,
		CapabilitySubtotalMonthly: capabilitySubtotal,
		TotalMonthly:              totalMonthly,
		TotalAnnual:               totalAnnual,
		SelfManagedComputeMonthly: selfManagedCompute,
		SelfManagedTotalMonthly:   selfManagedCompute,
		SelfManagedTotalAnnual:    selfManagedAnnual,
		ManagedVsSelfManaged:      totalMonthly - selfManagedCompute,
	}
}
