package cache

import (
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/heathcliff26/speedtest-exporter/pkg/speedtest"
)

// Minimum grace period for cache.
// Ensures that if the service is monitored in intervals, new tests will be run roughly every cacheTime.
const (
	minimumGraceDuration    = 30 * time.Second
	additionalGraceDuration = 5 * time.Second
)

type Cache struct {
	persist      bool
	path         string
	cacheTime    time.Duration
	cachedResult *speedtest.SpeedtestResult

	sync.RWMutex
}

// Create a new Cache instance and try to initialize it from disk if persist is true.
// If the path is not writable, the cache will not persist to disk.
// This function does not fail if it cannot read from disk, it will just log the error.
func NewCache(persist bool, path string, cacheTime time.Duration) *Cache {
	cache := &Cache{
		persist:   persist,
		path:      path,
		cacheTime: cacheTime,
	}

	if path == "" {
		cache.persist = false
	}
	if !cache.persist {
		return cache
	}

	// #nosec G302: Cache does not contain sensitive data, can be world readable
	f, err := os.OpenFile(cache.path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		slog.Info("Failed to open cache file, will not persist cache to disk", slog.String("file", cache.path), slog.Any("error", err))
		cache.persist = false
		return cache
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		slog.Info("Could not initialize cache from disk", slog.String("file", cache.path), slog.Any("error", err))
		return cache
	}

	if len(data) == 0 {
		slog.Info("Cache file is empty, starting with empty cache", slog.String("file", cache.path))
		return cache
	}

	cachedResult := &speedtest.SpeedtestResult{}
	err = cachedResult.UnmarshalJSON(data)
	if err != nil {
		slog.Info("Could not unmarshal cache data from disk", slog.String("file", cache.path), slog.Any("error", err))
	} else {
		slog.Info("Initialized cache from disk", slog.String("path", cache.path))
		cache.cachedResult = cachedResult
	}
	return cache
}

// Return the currently cached result and whether it is still valid.
// This method is safe to call even if the Cache instance is nil.
func (c *Cache) Read() (result *speedtest.SpeedtestResult, valid bool) {
	if c == nil {
		return nil, false
	}
	c.RLock()
	defer c.RUnlock()

	if c.cachedResult == nil {
		return nil, false
	}

	return c.cachedResult, c.expiresAt().After(time.Now())
}

// Save the given result to the cache.
// Attempt to persist to disk if enabled, but do not fail if it fails.
// This method is safe to call even if the Cache instance is nil.
func (c *Cache) Save(result *speedtest.SpeedtestResult) {
	if c == nil {
		return
	}
	c.Lock()
	defer c.Unlock()

	c.cachedResult = result
	if !c.persist {
		return
	}

	data, err := result.MarshalJSON()
	if err != nil {
		slog.Error("Could not marshal result to JSON", slog.Any("error", err))
		return
	}
	// #nosec G306: Cache does not contain sensitive data, can be world readable
	err = os.WriteFile(c.path, data, 0644)
	if err != nil {
		slog.Error("Could not write cache to disk", slog.String("file", c.path), slog.Any("error", err))
	}
}

// Return when the cache will expire
func (c *Cache) ExpiresAt() time.Time {
	if c == nil {
		return time.Time{}
	}
	c.RLock()
	defer c.RUnlock()

	if c.cachedResult == nil {
		return time.Time{}
	}

	return c.expiresAt()
}

// Return when the cache will expire, subtracting a grace period.
// Should be called when already verified that c is not nil and c.cachedResult is not nil.
// Assumes the caller holds at least a read lock.
func (c *Cache) expiresAt() time.Time {
	timestamp := c.cachedResult.TimestampAsTime()
	gracePeriod := time.Duration(c.cachedResult.Duration())*time.Millisecond + additionalGraceDuration
	if gracePeriod < minimumGraceDuration {
		gracePeriod = minimumGraceDuration
	}
	return timestamp.Add(c.cacheTime).Add(-1 * gracePeriod)
}
