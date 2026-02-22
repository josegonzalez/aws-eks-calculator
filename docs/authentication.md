# AWS Authentication

The calculator fetches live pricing from the AWS Pricing API on startup. AWS credentials are optional — without them, hardcoded default rates are used.

## Required Permission

The only IAM permission needed is `pricing:GetProducts`. A minimal policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "pricing:GetProducts",
      "Resource": "*"
    }
  ]
}
```

Any standard AWS credential method works (environment variables, `~/.aws/credentials`, IAM roles, SSO, etc.). The calculator uses the [AWS SDK default credential chain](https://docs.aws.amazon.com/sdkref/latest/guide/standardized-credentials.html).

## Services Queried

Two AWS services are queried via the Pricing API:

| Service Code | Purpose |
|---|---|
| `AmazonEKS` | EKS capability rates (ArgoCD, ACK, kro) |
| `AmazonECS` | Fargate compute rates (vCPU and memory) for self-managed comparison |

The Pricing API endpoint is always `us-east-1` regardless of which pricing region you select. The region selection controls which region's prices are returned, not which API endpoint is called.

## Fallback Behavior

When credentials are missing or a fetch fails, the calculator falls back to hardcoded default rates (based on `us-east-1` pricing). Each service is fetched independently — if EKS capability rates succeed but Fargate rates fail, the Fargate rates use defaults while the capability rates use live data.

See [calculations.md](calculations.md) for the default rate values.

## TUI Warnings

When a pricing fetch fails, the calculator shows a warning in the status bar:

- **First failure** (no rates previously loaded): "Using default rates"
- **Subsequent failure** (rates were loaded for another region): "Using previously fetched rates"
