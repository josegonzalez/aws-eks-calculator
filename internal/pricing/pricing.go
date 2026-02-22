package pricing

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
)

// Rates holds the hourly pricing rates fetched from AWS.
type Rates struct {
	ArgoCDBasePerHour    float64
	ArgoCDAppPerHour     float64
	ACKBasePerHour       float64
	ACKResourcePerHour   float64
	KroBasePerHour       float64
	KroRGDPerHour        float64
	FargateVCPUPerHour   float64
	FargateMemGBPerHour  float64
}

// ForCapability returns the base and resource hourly rates for the given capability.
func (r Rates) ForCapability(cap calculator.Capability) (base, resource float64) {
	switch cap {
	case calculator.CapabilityArgoCD:
		return r.ArgoCDBasePerHour, r.ArgoCDAppPerHour
	case calculator.CapabilityACK:
		return r.ACKBasePerHour, r.ACKResourcePerHour
	case calculator.CapabilityKro:
		return r.KroBasePerHour, r.KroRGDPerHour
	default:
		return 0, 0
	}
}

// HasAllCapabilityRates returns true if all capability rates are populated (> 0).
// This is used to detect stale cache entries that were written before new
// capability fields were added to the Rates struct.
func (r Rates) HasAllCapabilityRates() bool {
	return r.ArgoCDBasePerHour > 0 && r.ArgoCDAppPerHour > 0 &&
		r.ACKBasePerHour > 0 && r.ACKResourcePerHour > 0 &&
		r.KroBasePerHour > 0 && r.KroRGDPerHour > 0
}

// DefaultRates returns the hardcoded fallback rates.
func DefaultRates() Rates {
	return Rates{
		ArgoCDBasePerHour:   0.03,
		ArgoCDAppPerHour:    0.0015,
		ACKBasePerHour:      0.005,
		ACKResourcePerHour:  0.00005,
		KroBasePerHour:      0.005,
		KroRGDPerHour:       0.00005,
		FargateVCPUPerHour:  0.04048,
		FargateMemGBPerHour: 0.004446,
	}
}

// loadDefaultConfig and newPricingClient are package-level vars for testing.
var loadDefaultConfig = config.LoadDefaultConfig
var newPricingClient = func(cfg aws.Config) PricingAPI {
	return pricing.NewFromConfig(cfg)
}

// PricingAPI is the interface for the AWS Pricing GetProducts call.
type PricingAPI interface {
	GetProducts(ctx context.Context, params *pricing.GetProductsInput, optFns ...func(*pricing.Options)) (*pricing.GetProductsOutput, error)
}

// FetchRates checks the local cache first, then creates an AWS SDK client
// and fetches live pricing for the given region. On success the result is
// cached. Falls back to DefaultRates on any error.
func FetchRates(ctx context.Context, region string) (Rates, error) {
	cache := NewCache()
	if cached := cache.Load(region); cached != nil {
		if cached.HasAllCapabilityRates() {
			return *cached, nil
		}
	}

	cfg, err := loadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return DefaultRates(), nil
	}

	client := newPricingClient(cfg)
	rates, err := FetchRatesWithClient(ctx, client, region)
	if err != nil {
		return rates, nil
	}

	_ = cache.Save(region, rates)

	return rates, nil
}

// capabilitySuffixes defines the usage type suffixes for each EKS capability.
type capabilitySuffixes struct {
	baseSuffix     string
	resourceSuffix string
}

var allCapSuffixes = []struct {
	name     string
	suffixes capabilitySuffixes
}{
	{"ArgoCD", capabilitySuffixes{
		baseSuffix:     "AmazonEKSCapabilities-ArgoCD-Hours:perCapability",
		resourceSuffix: "AmazonEKSCapabilities-ArgoCD-CR-Hours:perCustomResource",
	}},
	{"ACK", capabilitySuffixes{
		baseSuffix:     "AmazonEKSCapabilities-ACK-Hours:perCapability",
		resourceSuffix: "AmazonEKSCapabilities-ACK-CR-Hours:perCustomResource",
	}},
	{"kro", capabilitySuffixes{
		baseSuffix:     "AmazonEKSCapabilities-KRO-Hours:perCapability",
		resourceSuffix: "AmazonEKSCapabilities-KRO-CR-Hours:perCustomResource",
	}},
}

// FetchRatesWithClient fetches live pricing using the provided client.
// EKS capabilities are fetched in a single pass to minimize API calls.
// Missing capability products are not treated as errors (defaults are used).
// Only actual API failures (network, auth) are returned as errors.
func FetchRatesWithClient(ctx context.Context, client PricingAPI, region string) (Rates, error) {
	rates := DefaultRates()

	found, err := fetchAllEKSCapabilities(ctx, client, region)
	if err != nil {
		return rates, fmt.Errorf("fetching EKS pricing: %w", err)
	}

	// Apply any rates that were found; missing ones keep defaults
	for _, cs := range allCapSuffixes {
		baseRate := found[cs.suffixes.baseSuffix]
		resRate := found[cs.suffixes.resourceSuffix]
		if baseRate > 0 && resRate > 0 {
			switch cs.name {
			case "ArgoCD":
				rates.ArgoCDBasePerHour = baseRate
				rates.ArgoCDAppPerHour = resRate
			case "ACK":
				rates.ACKBasePerHour = baseRate
				rates.ACKResourcePerHour = resRate
			case "kro":
				rates.KroBasePerHour = baseRate
				rates.KroRGDPerHour = resRate
			}
		}
	}

	vcpuRate, memRate, err := fetchFargate(ctx, client, region)
	if err == nil {
		rates.FargateVCPUPerHour = vcpuRate
		rates.FargateMemGBPerHour = memRate
	}

	return rates, nil
}

// fetchAllEKSCapabilities fetches all EKS capability rates in a single
// paginated query. It returns a map from suffix to rate for each product
// found. This avoids making 3 separate paginated queries (one per capability).
func fetchAllEKSCapabilities(ctx context.Context, client PricingAPI, region string) (map[string]float64, error) {
	// Build the set of suffixes we're looking for
	allSuffixes := make(map[string]bool)
	for _, cs := range allCapSuffixes {
		allSuffixes[cs.suffixes.baseSuffix] = false
		allSuffixes[cs.suffixes.resourceSuffix] = false
	}

	found := make(map[string]float64)
	var nextToken *string

	for {
		input := &pricing.GetProductsInput{
			ServiceCode: aws.String("AmazonEKS"),
			Filters: []types.Filter{
				{
					Type:  types.FilterTypeTermMatch,
					Field: aws.String("regionCode"),
					Value: aws.String(region),
				},
			},
			MaxResults: aws.Int32(100),
			NextToken:  nextToken,
		}

		output, err := client.GetProducts(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, priceJSON := range output.PriceList {
			var doc productDoc
			if err := json.Unmarshal([]byte(priceJSON), &doc); err != nil {
				continue
			}

			usageType := doc.Product.Attributes["usagetype"]
			for suffix, matched := range allSuffixes {
				if !matched && strings.HasSuffix(usageType, suffix) {
					if rate, err := extractRateFromDoc(doc); err == nil {
						found[suffix] = rate
						allSuffixes[suffix] = true
					}
				}
			}

			// Early return if all 6 rates found
			if len(found) == len(allSuffixes) {
				return found, nil
			}
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return found, nil
}

func fetchFargate(ctx context.Context, client PricingAPI, region string) (vcpuRate, memRate float64, err error) {
	vcpuInput := &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonECS"),
		Filters: []types.Filter{
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("regionCode"),
				Value: aws.String(region),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("productFamily"),
				Value: aws.String("Compute"),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("cputype"),
				Value: aws.String("perCPU"),
			},
		},
		MaxResults: aws.Int32(1),
	}

	vcpuRate, err = fetchSingleRate(ctx, client, vcpuInput)
	if err != nil {
		return 0, 0, fmt.Errorf("fargate vCPU: %w", err)
	}

	memInput := &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonECS"),
		Filters: []types.Filter{
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("regionCode"),
				Value: aws.String(region),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("productFamily"),
				Value: aws.String("Compute"),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("memorytype"),
				Value: aws.String("perGB"),
			},
		},
		MaxResults: aws.Int32(1),
	}

	memRate, err = fetchSingleRate(ctx, client, memInput)
	if err != nil {
		return 0, 0, fmt.Errorf("fargate memory: %w", err)
	}

	return vcpuRate, memRate, nil
}

// productDoc represents the JSON structure returned by the Pricing API.
type productDoc struct {
	Product struct {
		Attributes map[string]string `json:"attributes"`
	} `json:"product"`
	Terms struct {
		OnDemand map[string]struct {
			PriceDimensions map[string]struct {
				PricePerUnit map[string]string `json:"pricePerUnit"`
				Unit         string            `json:"unit"`
			} `json:"priceDimensions"`
		} `json:"OnDemand"`
	} `json:"terms"`
}

func extractRateFromDoc(doc productDoc) (float64, error) {
	for _, offer := range doc.Terms.OnDemand {
		for _, dim := range offer.PriceDimensions {
			usdStr, ok := dim.PricePerUnit["USD"]
			if !ok {
				continue
			}

			rate, err := strconv.ParseFloat(usdStr, 64)
			if err != nil {
				return 0, fmt.Errorf("parsing USD rate %q: %w", usdStr, err)
			}

			// Fargate pricing is per-second; convert to per-hour
			if strings.EqualFold(dim.Unit, "Second") || strings.EqualFold(dim.Unit, "Seconds") {
				rate *= 3600
			}

			return rate, nil
		}
	}

	return 0, fmt.Errorf("no OnDemand pricing found")
}

func fetchSingleRate(ctx context.Context, client PricingAPI, input *pricing.GetProductsInput) (float64, error) {
	output, err := client.GetProducts(ctx, input)
	if err != nil {
		return 0, err
	}

	if len(output.PriceList) == 0 {
		return 0, fmt.Errorf("no products found")
	}

	return parseRate(output.PriceList[0])
}

func parseRate(priceJSON string) (float64, error) {
	var doc productDoc
	if err := json.Unmarshal([]byte(priceJSON), &doc); err != nil {
		return 0, fmt.Errorf("parsing price JSON: %w", err)
	}

	return extractRateFromDoc(doc)
}
