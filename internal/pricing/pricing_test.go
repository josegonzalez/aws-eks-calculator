package pricing

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/pricing"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
)

type mockPricingAPI struct {
	responses map[string]*pricing.GetProductsOutput
	err       error
}

func (m *mockPricingAPI) GetProducts(ctx context.Context, params *pricing.GetProductsInput, optFns ...func(*pricing.Options)) (*pricing.GetProductsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := *params.ServiceCode
	for _, f := range params.Filters {
		key += ":" + *f.Field + "=" + *f.Value
	}

	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}

	return &pricing.GetProductsOutput{PriceList: []string{}}, nil
}

func eksProductJSON(usagetype, rate string) string {
	return fmt.Sprintf(`{
		"product": {
			"attributes": {
				"usagetype": "%s"
			}
		},
		"terms": {
			"OnDemand": {
				"offer1": {
					"priceDimensions": {
						"dim1": {
							"pricePerUnit": {"USD": "%s"},
							"unit": "Hour"
						}
					}
				}
			}
		}
	}`, usagetype, rate)
}

func fargateJSON(rate, unit string) string {
	return fmt.Sprintf(`{
		"product": {"attributes": {}},
		"terms": {
			"OnDemand": {
				"offer1": {
					"priceDimensions": {
						"dim1": {
							"pricePerUnit": {"USD": "%s"},
							"unit": "%s"
						}
					}
				}
			}
		}
	}`, rate, unit)
}

func allCapabilityProducts(region string) map[string]*pricing.GetProductsOutput {
	return map[string]*pricing.GetProductsOutput{
		"AmazonEKS:regionCode=" + region: {
			PriceList: []string{
				eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-Hours:perCapability", "0.03"),
				eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-CR-Hours:perCustomResource", "0.0015"),
				eksProductJSON("USE1-AmazonEKSCapabilities-ACK-Hours:perCapability", "0.005"),
				eksProductJSON("USE1-AmazonEKSCapabilities-ACK-CR-Hours:perCustomResource", "0.00005"),
				eksProductJSON("USE1-AmazonEKSCapabilities-KRO-Hours:perCapability", "0.005"),
				eksProductJSON("USE1-AmazonEKSCapabilities-KRO-CR-Hours:perCustomResource", "0.00005"),
			},
		},
		"AmazonECS:regionCode=" + region + ":productFamily=Compute:cputype=perCPU": {
			PriceList: []string{fargateJSON("0.000011244", "Second")},
		},
		"AmazonECS:regionCode=" + region + ":productFamily=Compute:memorytype=perGB": {
			PriceList: []string{fargateJSON("0.000001235", "Second")},
		},
	}
}

func TestFetchRatesSuccess(t *testing.T) {
	mock := &mockPricingAPI{responses: allCapabilityProducts("us-east-1")}

	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rates.ArgoCDBasePerHour != 0.03 {
		t.Errorf("ArgoCDBasePerHour: got %f, want 0.03", rates.ArgoCDBasePerHour)
	}
	if rates.ArgoCDAppPerHour != 0.0015 {
		t.Errorf("ArgoCDAppPerHour: got %f, want 0.0015", rates.ArgoCDAppPerHour)
	}
	if rates.ACKBasePerHour != 0.005 {
		t.Errorf("ACKBasePerHour: got %f, want 0.005", rates.ACKBasePerHour)
	}
	if rates.ACKResourcePerHour != 0.00005 {
		t.Errorf("ACKResourcePerHour: got %f, want 0.00005", rates.ACKResourcePerHour)
	}
	if rates.KroBasePerHour != 0.005 {
		t.Errorf("KroBasePerHour: got %f, want 0.005", rates.KroBasePerHour)
	}
	if rates.KroRGDPerHour != 0.00005 {
		t.Errorf("KroRGDPerHour: got %f, want 0.00005", rates.KroRGDPerHour)
	}

	// Fargate rates are per-second, converted to per-hour
	expectedVCPU := 0.000011244 * 3600
	if math.Abs(rates.FargateVCPUPerHour-expectedVCPU) > 0.001 {
		t.Errorf("FargateVCPUPerHour: got %f, want %f", rates.FargateVCPUPerHour, expectedVCPU)
	}
	expectedMem := 0.000001235 * 3600
	if math.Abs(rates.FargateMemGBPerHour-expectedMem) > 0.001 {
		t.Errorf("FargateMemGBPerHour: got %f, want %f", rates.FargateMemGBPerHour, expectedMem)
	}
}

func TestFetchRatesDifferentRegion(t *testing.T) {
	mock := &mockPricingAPI{
		responses: map[string]*pricing.GetProductsOutput{
			"AmazonEKS:regionCode=us-east-2": {
				PriceList: []string{
					eksProductJSON("USE2-AmazonEKSCapabilities-ArgoCD-Hours:perCapability", "0.03"),
					eksProductJSON("USE2-AmazonEKSCapabilities-ArgoCD-CR-Hours:perCustomResource", "0.0015"),
					eksProductJSON("USE2-AmazonEKSCapabilities-ACK-Hours:perCapability", "0.005"),
					eksProductJSON("USE2-AmazonEKSCapabilities-ACK-CR-Hours:perCustomResource", "0.00005"),
					eksProductJSON("USE2-AmazonEKSCapabilities-KRO-Hours:perCapability", "0.005"),
					eksProductJSON("USE2-AmazonEKSCapabilities-KRO-CR-Hours:perCustomResource", "0.00005"),
				},
			},
			"AmazonECS:regionCode=us-east-2:productFamily=Compute:cputype=perCPU": {
				PriceList: []string{fargateJSON("0.000012", "Second")},
			},
			"AmazonECS:regionCode=us-east-2:productFamily=Compute:memorytype=perGB": {
				PriceList: []string{fargateJSON("0.0000013", "Second")},
			},
		},
	}

	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rates.ArgoCDBasePerHour != 0.03 {
		t.Errorf("ArgoCDBasePerHour: got %f, want 0.03", rates.ArgoCDBasePerHour)
	}
	if rates.ArgoCDAppPerHour != 0.0015 {
		t.Errorf("ArgoCDAppPerHour: got %f, want 0.0015", rates.ArgoCDAppPerHour)
	}
}

func TestFetchRatesAPIError(t *testing.T) {
	mock := &mockPricingAPI{err: fmt.Errorf("access denied")}

	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err == nil {
		t.Fatal("expected error")
	}

	defaults := DefaultRates()
	if rates.ArgoCDBasePerHour != defaults.ArgoCDBasePerHour {
		t.Errorf("expected default ArgoCDBasePerHour on error")
	}
}

func TestFetchRatesMalformedJSON(t *testing.T) {
	mock := &mockPricingAPI{
		responses: map[string]*pricing.GetProductsOutput{
			"AmazonEKS:regionCode=us-east-1": {
				PriceList: []string{`{invalid json`},
			},
		},
	}

	// Malformed JSON products are skipped; the API call itself succeeded
	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defaults := DefaultRates()
	if rates.ArgoCDBasePerHour != defaults.ArgoCDBasePerHour {
		t.Errorf("expected default rates on malformed JSON")
	}
}

func TestFetchRatesMissingProducts(t *testing.T) {
	mock := &mockPricingAPI{
		responses: map[string]*pricing.GetProductsOutput{},
	}

	// Empty response is not an API error — defaults are used silently
	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defaults := DefaultRates()
	if rates.ArgoCDBasePerHour != defaults.ArgoCDBasePerHour {
		t.Errorf("expected default rates on missing products")
	}
}

func TestDefaultRates(t *testing.T) {
	r := DefaultRates()
	if r.ArgoCDBasePerHour != 0.03 {
		t.Errorf("ArgoCDBasePerHour: got %f, want 0.03", r.ArgoCDBasePerHour)
	}
	if r.ArgoCDAppPerHour != 0.0015 {
		t.Errorf("ArgoCDAppPerHour: got %f, want 0.0015", r.ArgoCDAppPerHour)
	}
	if r.ACKBasePerHour != 0.005 {
		t.Errorf("ACKBasePerHour: got %f, want 0.005", r.ACKBasePerHour)
	}
	if r.ACKResourcePerHour != 0.00005 {
		t.Errorf("ACKResourcePerHour: got %f, want 0.00005", r.ACKResourcePerHour)
	}
	if r.KroBasePerHour != 0.005 {
		t.Errorf("KroBasePerHour: got %f, want 0.005", r.KroBasePerHour)
	}
	if r.KroRGDPerHour != 0.00005 {
		t.Errorf("KroRGDPerHour: got %f, want 0.00005", r.KroRGDPerHour)
	}
	if r.FargateVCPUPerHour != 0.04048 {
		t.Errorf("FargateVCPUPerHour: got %f, want 0.04048", r.FargateVCPUPerHour)
	}
	if r.FargateMemGBPerHour != 0.004446 {
		t.Errorf("FargateMemGBPerHour: got %f, want 0.004446", r.FargateMemGBPerHour)
	}
}

func TestForCapability(t *testing.T) {
	r := DefaultRates()

	base, res := r.ForCapability(calculator.CapabilityArgoCD)
	if base != r.ArgoCDBasePerHour || res != r.ArgoCDAppPerHour {
		t.Errorf("ArgoCD: got base=%f res=%f", base, res)
	}

	base, res = r.ForCapability(calculator.CapabilityACK)
	if base != r.ACKBasePerHour || res != r.ACKResourcePerHour {
		t.Errorf("ACK: got base=%f res=%f", base, res)
	}

	base, res = r.ForCapability(calculator.CapabilityKro)
	if base != r.KroBasePerHour || res != r.KroRGDPerHour {
		t.Errorf("kro: got base=%f res=%f", base, res)
	}

	base, res = r.ForCapability(calculator.Capability(99))
	if base != 0 || res != 0 {
		t.Errorf("unknown: got base=%f res=%f", base, res)
	}
}

func TestParseRateHourly(t *testing.T) {
	json := eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-Hours:perCapability", "0.05")
	rate, err := parseRate(json)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 0.05 {
		t.Errorf("got %f, want 0.05", rate)
	}
}

func TestParseRatePerSecond(t *testing.T) {
	json := fargateJSON("0.001", "Second")
	rate, err := parseRate(json)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 0.001 * 3600
	if math.Abs(rate-expected) > 0.001 {
		t.Errorf("got %f, want %f", rate, expected)
	}
}

func TestParseRateNoOnDemand(t *testing.T) {
	json := `{"product": {"attributes": {}}, "terms": {"OnDemand": {}}}`
	_, err := parseRate(json)
	if err == nil {
		t.Fatal("expected error for empty OnDemand")
	}
}

func TestParseRateNoUSD(t *testing.T) {
	json := `{
		"product": {"attributes": {}},
		"terms": {
			"OnDemand": {
				"offer1": {
					"priceDimensions": {
						"dim1": {
							"pricePerUnit": {"EUR": "0.05"},
							"unit": "Hour"
						}
					}
				}
			}
		}
	}`
	_, err := parseRate(json)
	if err == nil {
		t.Fatal("expected error for missing USD")
	}
}

func TestFetchRatesFargateError(t *testing.T) {
	// EKS succeeds but Fargate fails (missing products)
	mock := &mockPricingAPI{
		responses: map[string]*pricing.GetProductsOutput{
			"AmazonEKS:regionCode=us-east-1": {
				PriceList: []string{
					eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-Hours:perCapability", "0.03"),
					eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-CR-Hours:perCustomResource", "0.0015"),
					eksProductJSON("USE1-AmazonEKSCapabilities-ACK-Hours:perCapability", "0.005"),
					eksProductJSON("USE1-AmazonEKSCapabilities-ACK-CR-Hours:perCustomResource", "0.00005"),
					eksProductJSON("USE1-AmazonEKSCapabilities-KRO-Hours:perCapability", "0.005"),
					eksProductJSON("USE1-AmazonEKSCapabilities-KRO-CR-Hours:perCustomResource", "0.00005"),
				},
			},
			// No Fargate entries — Fargate failure is not an error
		},
	}

	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defaults := DefaultRates()
	// EKS rates should be updated
	if rates.ArgoCDBasePerHour != 0.03 {
		t.Errorf("ArgoCDBasePerHour: got %f, want 0.03", rates.ArgoCDBasePerHour)
	}
	// Fargate rates should remain defaults
	if rates.FargateVCPUPerHour != defaults.FargateVCPUPerHour {
		t.Errorf("FargateVCPUPerHour should be default when missing")
	}
}

func TestFetchEKSArgoCDMissingBaseOnly(t *testing.T) {
	mock := &mockPricingAPI{
		responses: map[string]*pricing.GetProductsOutput{
			"AmazonEKS:regionCode=us-east-1": {
				PriceList: []string{
					eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-CR-Hours:perCustomResource", "0.00136"),
				},
			},
		},
	}

	// Missing base product is not an error — defaults are used
	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defaults := DefaultRates()
	if rates.ArgoCDBasePerHour != defaults.ArgoCDBasePerHour {
		t.Errorf("expected default ArgoCDBasePerHour on partial failure")
	}
}

func TestFetchEKSArgoCDMissingAppOnly(t *testing.T) {
	mock := &mockPricingAPI{
		responses: map[string]*pricing.GetProductsOutput{
			"AmazonEKS:regionCode=us-east-1": {
				PriceList: []string{
					eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-Hours:perCapability", "0.02771"),
				},
			},
		},
	}

	// Missing per-application product is not an error — defaults are used
	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defaults := DefaultRates()
	if rates.ArgoCDAppPerHour != defaults.ArgoCDAppPerHour {
		t.Errorf("expected default ArgoCDAppPerHour on partial failure")
	}
}

// paginatedMockPricingAPI returns results across multiple pages.
type paginatedMockPricingAPI struct {
	pages []map[string]*pricing.GetProductsOutput
}

func (m *paginatedMockPricingAPI) GetProducts(ctx context.Context, params *pricing.GetProductsInput, optFns ...func(*pricing.Options)) (*pricing.GetProductsOutput, error) {
	key := *params.ServiceCode
	for _, f := range params.Filters {
		key += ":" + *f.Field + "=" + *f.Value
	}

	pageIdx := 0
	if params.NextToken != nil {
		_, _ = fmt.Sscanf(*params.NextToken, "%d", &pageIdx)
	}

	if pageIdx >= len(m.pages) {
		return &pricing.GetProductsOutput{PriceList: []string{}}, nil
	}

	resp, ok := m.pages[pageIdx][key]
	if !ok {
		return &pricing.GetProductsOutput{PriceList: []string{}}, nil
	}

	out := *resp
	if pageIdx+1 < len(m.pages) {
		next := fmt.Sprintf("%d", pageIdx+1)
		out.NextToken = &next
	}
	return &out, nil
}

func TestFetchEKSArgoCDPaginated(t *testing.T) {
	mock := &paginatedMockPricingAPI{
		pages: []map[string]*pricing.GetProductsOutput{
			{
				"AmazonEKS:regionCode=us-east-1": {
					PriceList: []string{
						eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-CR-Hours:perCustomResource", "0.0015"),
						eksProductJSON("USE1-AmazonEKSCapabilities-ACK-Hours:perCapability", "0.005"),
						eksProductJSON("USE1-AmazonEKSCapabilities-ACK-CR-Hours:perCustomResource", "0.00005"),
						eksProductJSON("USE1-AmazonEKSCapabilities-KRO-Hours:perCapability", "0.005"),
						eksProductJSON("USE1-AmazonEKSCapabilities-KRO-CR-Hours:perCustomResource", "0.00005"),
					},
				},
				"AmazonECS:regionCode=us-east-1:productFamily=Compute:cputype=perCPU": {
					PriceList: []string{fargateJSON("0.04048", "Hour")},
				},
				"AmazonECS:regionCode=us-east-1:productFamily=Compute:memorytype=perGB": {
					PriceList: []string{fargateJSON("0.004446", "Hour")},
				},
			},
			{
				"AmazonEKS:regionCode=us-east-1": {
					PriceList: []string{
						eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-Hours:perCapability", "0.03"),
					},
				},
			},
		},
	}

	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rates.ArgoCDBasePerHour != 0.03 {
		t.Errorf("ArgoCDBasePerHour: got %f, want 0.03", rates.ArgoCDBasePerHour)
	}
	if rates.ArgoCDAppPerHour != 0.0015 {
		t.Errorf("ArgoCDAppPerHour: got %f, want 0.0015", rates.ArgoCDAppPerHour)
	}
}

func TestHasAllCapabilityRates(t *testing.T) {
	// DefaultRates should have all capability rates
	r := DefaultRates()
	if !r.HasAllCapabilityRates() {
		t.Error("DefaultRates() should have all capability rates")
	}

	// Zero ACKBasePerHour should fail
	r = DefaultRates()
	r.ACKBasePerHour = 0
	if r.HasAllCapabilityRates() {
		t.Error("should return false when ACKBasePerHour is 0")
	}

	// Zero KroRGDPerHour should fail
	r = DefaultRates()
	r.KroRGDPerHour = 0
	if r.HasAllCapabilityRates() {
		t.Error("should return false when KroRGDPerHour is 0")
	}

	// Zero ArgoCDAppPerHour should fail
	r = DefaultRates()
	r.ArgoCDAppPerHour = 0
	if r.HasAllCapabilityRates() {
		t.Error("should return false when ArgoCDAppPerHour is 0")
	}
}

func TestFetchRatesZeroRateProducts(t *testing.T) {
	// ACK and kro return $0 pricing; ArgoCD returns valid pricing
	mock := &mockPricingAPI{
		responses: map[string]*pricing.GetProductsOutput{
			"AmazonEKS:regionCode=us-east-1": {
				PriceList: []string{
					eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-Hours:perCapability", "0.03"),
					eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-CR-Hours:perCustomResource", "0.0015"),
					eksProductJSON("USE1-AmazonEKSCapabilities-ACK-Hours:perCapability", "0"),
					eksProductJSON("USE1-AmazonEKSCapabilities-ACK-CR-Hours:perCustomResource", "0"),
					eksProductJSON("USE1-AmazonEKSCapabilities-KRO-Hours:perCapability", "0"),
					eksProductJSON("USE1-AmazonEKSCapabilities-KRO-CR-Hours:perCustomResource", "0"),
				},
			},
			"AmazonECS:regionCode=us-east-1:productFamily=Compute:cputype=perCPU": {
				PriceList: []string{fargateJSON("0.04048", "Hour")},
			},
			"AmazonECS:regionCode=us-east-1:productFamily=Compute:memorytype=perGB": {
				PriceList: []string{fargateJSON("0.004446", "Hour")},
			},
		},
	}

	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ArgoCD should use live rates
	if rates.ArgoCDBasePerHour != 0.03 {
		t.Errorf("ArgoCDBasePerHour: got %f, want 0.03", rates.ArgoCDBasePerHour)
	}

	// ACK and kro should retain defaults (not be overwritten with 0)
	defaults := DefaultRates()
	if rates.ACKBasePerHour != defaults.ACKBasePerHour {
		t.Errorf("ACKBasePerHour should be default, got %f", rates.ACKBasePerHour)
	}
	if rates.ACKResourcePerHour != defaults.ACKResourcePerHour {
		t.Errorf("ACKResourcePerHour should be default, got %f", rates.ACKResourcePerHour)
	}
	if rates.KroBasePerHour != defaults.KroBasePerHour {
		t.Errorf("KroBasePerHour should be default, got %f", rates.KroBasePerHour)
	}
	if rates.KroRGDPerHour != defaults.KroRGDPerHour {
		t.Errorf("KroRGDPerHour should be default, got %f", rates.KroRGDPerHour)
	}
}

func TestParseRateMalformedJSON(t *testing.T) {
	_, err := parseRate("{invalid}")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestExtractRateInvalidUSD(t *testing.T) {
	doc := productDoc{}
	doc.Terms.OnDemand = map[string]struct {
		PriceDimensions map[string]struct {
			PricePerUnit map[string]string `json:"pricePerUnit"`
			Unit         string            `json:"unit"`
		} `json:"priceDimensions"`
	}{
		"offer1": {
			PriceDimensions: map[string]struct {
				PricePerUnit map[string]string `json:"pricePerUnit"`
				Unit         string            `json:"unit"`
			}{
				"dim1": {
					PricePerUnit: map[string]string{"USD": "notanumber"},
					Unit:         "Hour",
				},
			},
		},
	}

	_, err := extractRateFromDoc(doc)
	if err == nil {
		t.Fatal("expected error for invalid USD value")
	}
}

func TestFetchSingleRateClientError(t *testing.T) {
	mock := &mockPricingAPI{err: fmt.Errorf("connection refused")}
	svc := "AmazonECS"
	_, err := fetchSingleRate(context.Background(), mock, &pricing.GetProductsInput{
		ServiceCode: &svc,
	})
	if err == nil {
		t.Fatal("expected error from client")
	}
}

func TestFetchFargateVCPUError(t *testing.T) {
	// EKS succeeds, Fargate vCPU call returns error via selective mock
	mock := &selectiveErrorMock{
		responses: allCapabilityProducts("us-east-1"),
		errOn:     "AmazonECS:regionCode=us-east-1:productFamily=Compute:cputype=perCPU",
	}

	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fargate rates should remain defaults since vCPU fetch failed
	defaults := DefaultRates()
	if rates.FargateVCPUPerHour != defaults.FargateVCPUPerHour {
		t.Errorf("FargateVCPUPerHour should be default, got %f", rates.FargateVCPUPerHour)
	}
}

func TestFetchFargateMemError(t *testing.T) {
	// EKS succeeds, Fargate vCPU succeeds, but memory call returns error
	mock := &selectiveErrorMock{
		responses: allCapabilityProducts("us-east-1"),
		errOn:     "AmazonECS:regionCode=us-east-1:productFamily=Compute:memorytype=perGB",
	}

	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fargate rates should remain defaults since memory fetch failed
	defaults := DefaultRates()
	if rates.FargateMemGBPerHour != defaults.FargateMemGBPerHour {
		t.Errorf("FargateMemGBPerHour should be default, got %f", rates.FargateMemGBPerHour)
	}
}

// selectiveErrorMock returns an error only for a specific key.
type selectiveErrorMock struct {
	responses map[string]*pricing.GetProductsOutput
	errOn     string
}

func (m *selectiveErrorMock) GetProducts(ctx context.Context, params *pricing.GetProductsInput, optFns ...func(*pricing.Options)) (*pricing.GetProductsOutput, error) {
	key := *params.ServiceCode
	for _, f := range params.Filters {
		key += ":" + *f.Field + "=" + *f.Value
	}

	if key == m.errOn {
		return nil, fmt.Errorf("selective error on %s", key)
	}

	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}

	return &pricing.GetProductsOutput{PriceList: []string{}}, nil
}

func TestFetchPartialCapabilityFailure(t *testing.T) {
	// Only ArgoCD products available; ACK and kro missing
	mock := &mockPricingAPI{
		responses: map[string]*pricing.GetProductsOutput{
			"AmazonEKS:regionCode=us-east-1": {
				PriceList: []string{
					eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-Hours:perCapability", "0.03"),
					eksProductJSON("USE1-AmazonEKSCapabilities-ArgoCD-CR-Hours:perCustomResource", "0.0015"),
				},
			},
			"AmazonECS:regionCode=us-east-1:productFamily=Compute:cputype=perCPU": {
				PriceList: []string{fargateJSON("0.04048", "Hour")},
			},
			"AmazonECS:regionCode=us-east-1:productFamily=Compute:memorytype=perGB": {
				PriceList: []string{fargateJSON("0.004446", "Hour")},
			},
		},
	}

	rates, err := FetchRatesWithClient(context.Background(), mock, "us-east-1")
	// Missing capabilities are not errors — defaults are used silently
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ArgoCD rates should be updated
	if rates.ArgoCDBasePerHour != 0.03 {
		t.Errorf("ArgoCDBasePerHour: got %f, want 0.03", rates.ArgoCDBasePerHour)
	}

	// ACK and kro should have defaults
	defaults := DefaultRates()
	if rates.ACKBasePerHour != defaults.ACKBasePerHour {
		t.Errorf("ACKBasePerHour should be default on failure")
	}
	if rates.KroBasePerHour != defaults.KroBasePerHour {
		t.Errorf("KroBasePerHour should be default on failure")
	}
}
