package pricing

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestCache(t *testing.T) *Cache {
	t.Helper()
	return &Cache{
		dir: t.TempDir(),
		now: time.Now,
	}
}

func TestCacheSaveAndLoad(t *testing.T) {
	c := newTestCache(t)
	rates := Rates{
		ArgoCDBasePerHour:   0.03,
		ArgoCDAppPerHour:    0.0015,
		ACKBasePerHour:      0.03,
		ACKResourcePerHour:  0.0015,
		KroBasePerHour:      0.03,
		KroRGDPerHour:       0.0015,
		FargateVCPUPerHour:  0.05,
		FargateMemGBPerHour: 0.005,
	}

	if err := c.Save("us-east-1", rates); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded := c.Load("us-east-1")
	if loaded == nil {
		t.Fatal("Load returned nil for cached region")
	}
	if *loaded != rates {
		t.Errorf("loaded rates %+v != saved rates %+v", *loaded, rates)
	}
}

func TestCacheLoadMiss(t *testing.T) {
	c := newTestCache(t)

	if loaded := c.Load("eu-west-1"); loaded != nil {
		t.Error("expected nil for uncached region")
	}
}

func TestCacheExpiry(t *testing.T) {
	c := newTestCache(t)
	staleTime := time.Now().Add(-25 * time.Hour)
	c.now = func() time.Time { return staleTime }

	rates := DefaultRates()
	if err := c.Save("us-east-1", rates); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reset clock to current time; the entry was saved 25h ago
	c.now = time.Now

	if loaded := c.Load("us-east-1"); loaded != nil {
		t.Error("expected nil for expired cache entry")
	}
}

func TestCacheNotExpiredAt24h(t *testing.T) {
	c := newTestCache(t)
	recentTime := time.Now().Add(-23 * time.Hour)
	c.now = func() time.Time { return recentTime }

	rates := DefaultRates()
	if err := c.Save("us-east-1", rates); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	c.now = time.Now

	if loaded := c.Load("us-east-1"); loaded == nil {
		t.Error("cache entry should still be valid at 23h")
	}
}

func TestCacheCorruptFile(t *testing.T) {
	c := newTestCache(t)

	path := c.path("us-east-1")
	if err := os.MkdirAll(c.dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	if loaded := c.Load("us-east-1"); loaded != nil {
		t.Error("expected nil for corrupt cache file")
	}
}

func TestCacheIsolatesRegions(t *testing.T) {
	c := newTestCache(t)

	r1 := Rates{ArgoCDBasePerHour: 0.03}
	r2 := Rates{ArgoCDBasePerHour: 0.04}

	if err := c.Save("us-east-1", r1); err != nil {
		t.Fatal(err)
	}
	if err := c.Save("eu-west-1", r2); err != nil {
		t.Fatal(err)
	}

	loaded1 := c.Load("us-east-1")
	loaded2 := c.Load("eu-west-1")

	if loaded1 == nil || loaded1.ArgoCDBasePerHour != 0.03 {
		t.Errorf("us-east-1: got %+v", loaded1)
	}
	if loaded2 == nil || loaded2.ArgoCDBasePerHour != 0.04 {
		t.Errorf("eu-west-1: got %+v", loaded2)
	}
}

func TestCacheSaveCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "cache")
	c := &Cache{dir: dir, now: time.Now}

	if err := c.Save("us-east-1", DefaultRates()); err != nil {
		t.Fatalf("Save should create intermediate dirs: %v", err)
	}

	if loaded := c.Load("us-east-1"); loaded == nil {
		t.Error("should load after creating directory")
	}
}

func TestCacheLoadRejectsZeroRates(t *testing.T) {
	c := newTestCache(t)
	// Simulate a stale cache entry that was written before ACK/KRO fields existed.
	// The JSON will have zero values for ACK/KRO fields.
	staleRates := Rates{
		ArgoCDBasePerHour:   0.03,
		ArgoCDAppPerHour:    0.0015,
		FargateVCPUPerHour:  0.04048,
		FargateMemGBPerHour: 0.00511175,
		// ACK and KRO fields are zero (as if missing from old JSON)
	}

	if err := c.Save("us-east-1", staleRates); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Cache.Load returns the struct (it doesn't know about validity)
	loaded := c.Load("us-east-1")
	if loaded == nil {
		t.Fatal("Load should return the cached struct")
	}

	// But HasAllCapabilityRates should reject it
	if loaded.HasAllCapabilityRates() {
		t.Error("stale cache with zero ACK/KRO rates should fail HasAllCapabilityRates")
	}
}

func TestNewCache(t *testing.T) {
	c := NewCache()
	if c == nil {
		t.Fatal("NewCache returned nil")
	}
	if c.dir == "" {
		t.Error("NewCache dir should not be empty")
	}
	if c.now == nil {
		t.Error("NewCache now func should not be nil")
	}
}

func TestCacheSaveMkdirError(t *testing.T) {
	tmp := t.TempDir()
	blockingFile := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blockingFile, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	c := &Cache{
		dir: filepath.Join(blockingFile, "subdir"),
		now: time.Now,
	}

	if err := c.Save("us-east-1", DefaultRates()); err == nil {
		t.Error("expected error when cache dir is under a file")
	}
}

func TestCacheSaveOverwrites(t *testing.T) {
	c := newTestCache(t)

	old := Rates{ArgoCDBasePerHour: 0.01}
	updated := Rates{ArgoCDBasePerHour: 0.05}

	if err := c.Save("us-east-1", old); err != nil {
		t.Fatal(err)
	}
	if err := c.Save("us-east-1", updated); err != nil {
		t.Fatal(err)
	}

	loaded := c.Load("us-east-1")
	if loaded == nil || loaded.ArgoCDBasePerHour != 0.05 {
		t.Errorf("expected updated rate 0.05, got %+v", loaded)
	}
}

func TestCacheSaveJsonMarshalError(t *testing.T) {
	orig := cacheJSONMarshal
	defer func() { cacheJSONMarshal = orig }()

	cacheJSONMarshal = func(v any) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	}

	c := newTestCache(t)
	err := c.Save("us-east-1", DefaultRates())
	if err == nil {
		t.Error("expected error from json marshal")
	}
	if err.Error() != "marshal error" {
		t.Errorf("unexpected error: %v", err)
	}
}
