# Cost Calculations

This document explains how the AWS EKS Capabilities cost calculator derives its estimates.

## Supported Capabilities

The calculator supports three EKS capabilities, all sharing the same billing model (base per cluster/hr + per-resource/hr):

| Capability | Base Unit | Resource Unit |
|------------|-----------|---------------|
| ArgoCD | per cluster/hr | per Application/hr |
| ACK | per cluster/hr | per managed AWS resource/hr |
| kro | per cluster/hr | per RGD instance/hr |

## Pricing Source

Pricing is fetched dynamically from the [AWS Pricing API](https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/price-changes.html) (`GetProducts`) on startup. The region is configurable (default: `us-east-1`) and changing it in the TUI triggers a re-fetch. Each capability is fetched independently; if one fails, defaults are used for that capability. If AWS credentials or network are unavailable, the calculator falls back to hardcoded defaults.

| Component | Default Rate |
|---|---|
| ArgoCD base capability | $0.03/hr |
| ArgoCD per-application | $0.0015/hr |
| ACK base capability | $0.005/hr |
| ACK per-resource | $0.00005/hr |
| kro base capability | $0.005/hr |
| kro per-RGD | $0.00005/hr |

## Hours Per Month

The default is **730 hours/month**, the standard AWS billing assumption derived from `365.25 days * 24 hours / 12 months = 730.5`, rounded to 730. Users can override this for partial-month estimates.

## Total Resources

Resources are the billable unit for the per-resource fee:

```
total_resources = (num_clusters * resources_per_cluster) + appset_expansion
```

- For **ArgoCD**, `appset_expansion = app_templates * clusters_per_template` (ApplicationSets generate one Application per target cluster)
- For **ACK** and **kro**, `appset_expansion = 0` (no ApplicationSet concept)

## Managed Service Costs

### Base Capability Fee

Charged per cluster that has the capability enabled:

```
base_capability_monthly = base_rate/hr * hours_per_month * num_clusters
```

Example: 3 clusters at 730 hours = `0.03 * 730 * 3 = $65.70/mo`

### Per-Resource Fee

Charged per resource instance across all clusters:

```
per_resource_monthly = resource_rate/hr * total_resources * hours_per_month
```

Example: 30 total resources at 730 hours = `0.0015 * 30 * 730 = $32.85/mo`

### Capability Subtotal

```
capability_subtotal = base_capability_monthly + per_resource_monthly
```

## Total Cost

The calculator assumes you already have EKS clusters running, so EKS control plane fees are not included in the cost breakdown.

```
total_monthly = capability_subtotal
total_annual  = total_monthly * 12
```

## Self-Managed Comparison

This estimates the compute cost of running the capability yourself on EKS, rather than using the managed service. It helps answer: "Is the managed fee worth it compared to running my own?"

### Compute Cost

The self-managed estimate uses the vCPU and memory resources needed to run the capability pods on each cluster:

```
compute_per_cluster = (vcpu_per_cluster * vcpu_cost_per_hour) + (memory_gb_per_cluster * memory_gb_cost_per_hour)
self_managed_compute_monthly = compute_per_cluster * hours_per_month * num_clusters
```

Default rates use EKS Fargate pricing in us-east-1 (Linux/X86):

| Resource | Default | Rate |
|---|---|---|
| vCPU per cluster | 1.0 | $0.04048/hr (Fargate: $0.000011244/vCPU/s) |
| Memory GB per cluster | 2.0 | $0.004446/hr (Fargate: $0.000001235/GB/s) |

### Managed vs Self-Managed Difference

```
difference = total_monthly - self_managed_total_monthly
```

- **Positive**: AWS managed costs more than self-managed compute
- **Negative**: AWS managed costs less

### Caveats

The self-managed comparison **only accounts for compute costs**. It does **not** include:

- Engineer time for installation, upgrades, and maintenance
- High-availability configuration (multiple replicas, pod disruption budgets)
- Monitoring and alerting setup
- Security patching and CVE response
- Backup and disaster recovery
- Network costs for cross-cluster sync

The managed service handles all of the above, so the actual cost advantage of managed capabilities is larger than the raw compute difference suggests.

## ArgoCD ApplicationSets

ArgoCD has an additional concept: **ApplicationSets**. An ApplicationSet template generates one Application per target cluster, so `app_templates * clusters_per_template` additional billable Applications are created. ACK and kro do not have this concept.

## Worked Example

**Scenario**: ArgoCD, 3 clusters, 10 apps per cluster, 730 hours/month

| Line Item | Calculation | Cost |
|---|---|---|
| Base capability | $0.03 x 730 x 3 | $65.70/mo |
| Per-application | $0.0015 x 30 x 730 | $32.85/mo |
| **Monthly total** | | **$98.55/mo** |
| **Annual total** | $98.55 x 12 | **$1,182.60/yr** |
| Self-managed compute | (1.0 x $0.04048 + 2.0 x $0.004446) x 730 x 3 | $108.12/mo |
| Difference | $98.55 - $108.12 | -$9.57/mo |
