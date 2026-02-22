# Pricing Cache

The calculator caches AWS pricing rates to a temporary file so that subsequent launches and region switches don't require a fresh API call every time.

## How it works

1. When pricing is requested for a region, the cache is checked first.
2. If a valid (non-expired) cache file exists, its rates are returned immediately — no API call is made.
3. If the cache is missing, expired, or corrupt, the AWS Pricing API is called.
4. On a successful API response the rates are written to the cache for next time.

Cache writes are best-effort; failures are silently ignored so the calculator still works even if the temp directory is unwritable.

## File location

Cache files are stored under the OS temporary directory:

```
<os.TempDir()>/aws-eks-calculator/rates-<region>.json
```

For example on Linux/macOS:

```
/tmp/aws-eks-calculator/rates-us-east-1.json
```

## File format

Each file is a JSON object:

```json
{
  "rates": {
    "ArgoCDBasePerHour": 0.03,
    "ArgoCDAppPerHour": 0.0015,
    "FargateVCPUPerHour": 0.04048,
    "FargateMemGBPerHour": 0.004446
  },
  "fetched_at": "2026-02-19T12:00:00Z"
}
```

## Background warming

After the first successful pricing fetch, the calculator spawns a background task that sequentially fetches and caches rates for every other region. This means switching regions later is typically instant (served from cache) rather than requiring a live API call. The warming runs once per session and does not block the UI.

## Expiry

Cache entries expire after **24 hours** (measured from `fetched_at`). After expiry the file is ignored and a new API call is made. The stale file is overwritten on the next successful fetch.

## Clearing the cache

Delete the cache directory to force a fresh fetch on the next launch:

```sh
rm -rf "${TMPDIR:-/tmp}/aws-eks-calculator"
```
