package pricing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// cacheTTL is how long cached rates remain valid.
const cacheTTL = 24 * time.Hour

const cacheSubdir = "aws-eks-calculator"

var cacheJSONMarshal = json.Marshal

// cachedRates is the on-disk format for cached pricing data.
type cachedRates struct {
	Rates     Rates     `json:"rates"`
	FetchedAt time.Time `json:"fetched_at"`
}

// Cache handles reading and writing pricing rates to a temporary file.
type Cache struct {
	dir string
	now func() time.Time
}

// NewCache creates a cache that stores files under os.TempDir().
func NewCache() *Cache {
	return &Cache{
		dir: filepath.Join(os.TempDir(), cacheSubdir),
		now: time.Now,
	}
}

func (c *Cache) path(region string) string {
	return filepath.Join(c.dir, fmt.Sprintf("rates-%s.json", region))
}

// Load returns cached rates for the given region if a valid (non-expired)
// cache file exists. Returns nil if the cache is missing, expired, or corrupt.
func (c *Cache) Load(region string) *Rates {
	data, err := os.ReadFile(c.path(region))
	if err != nil {
		return nil
	}

	var entry cachedRates
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	if c.now().Sub(entry.FetchedAt) > cacheTTL {
		return nil
	}

	return &entry.Rates
}

// Save writes rates to the cache file for the given region.
// Errors are returned but callers may choose to ignore them.
func (c *Cache) Save(region string, rates Rates) error {
	if err := os.MkdirAll(c.dir, 0o700); err != nil {
		return err
	}

	entry := cachedRates{
		Rates:     rates,
		FetchedAt: c.now(),
	}

	data, err := cacheJSONMarshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(c.path(region), data, 0o600)
}
