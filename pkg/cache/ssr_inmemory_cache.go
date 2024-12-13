package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type SSRInMemoryCache struct {
	data     map[string]CacheEntry
	rwMutex  sync.RWMutex
	ttl      time.Duration
	maxSize  int
	hits     int64
	misses   int64
	cleanups int64
}

func NewSSRInMemoryCache(config CacheConfig) *SSRInMemoryCache {
	if config.MaxSize <= 0 {
		config.MaxSize = 1000 // Default max entries
	}

	cache := &SSRInMemoryCache{
		data:    make(map[string]CacheEntry, config.MaxSize),
		ttl:     config.TTL,
		maxSize: config.MaxSize,
	}

	go func() {
		ticker := time.NewTicker(time.Minute * 5) // Reduced cleanup frequency
		defer ticker.Stop()
		for range ticker.C {
			if len(cache.data) > 0 { // Only cleanup if there's data
				cache.cleanup()
			}
		}
	}()

	return cache
}

func (c *SSRInMemoryCache) Get(key string) (CacheEntry, bool) {
	c.rwMutex.RLock() // Use RLock for concurrent reads
	entry, exists := c.data[key]
	c.rwMutex.RUnlock()

	if !exists {
		c.misses++
		log.Debugf("Cache miss for key: %s", key)
		return CacheEntry{}, false
	}

	c.hits++
	log.Debugf("Cache hit for key: %s", key)
	return entry, true
}

func (c *SSRInMemoryCache) Set(key string, content []byte) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	// If we're at capacity, remove oldest entry
	if len(c.data) >= c.maxSize {
		var oldestKey string
		var oldestTime time.Time
		first := true

		for k, v := range c.data {
			if first || v.LastUpdated.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.LastUpdated
				first = false
			}
		}
		delete(c.data, oldestKey)
		log.Debugf("Removed oldest cache entry: %s", oldestKey)
	}

	// Generate ETag
	hasher := sha256.New()
	hasher.Write(content)
	etag := `"` + hex.EncodeToString(hasher.Sum(nil)) + `"`

	c.data[key] = CacheEntry{
		Content:     content,
		LastUpdated: time.Now(),
		ETag:        etag,
	}
	log.Debugf("Cache entry set for key: %s", key)
}

func (c *SSRInMemoryCache) cleanup() {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	before := len(c.data)
	now := time.Now()

	for key, entry := range c.data {
		if now.Sub(entry.LastUpdated) > c.ttl {
			delete(c.data, key)
		}
	}

	after := len(c.data)
	if removed := before - after; removed > 0 {
		c.cleanups++
		log.Debugf("Cache cleanup: removed %d expired entries", removed)
	}
}

func (c *SSRInMemoryCache) GetMetrics() map[string]interface{} {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return map[string]interface{}{
		"type":     "memory",
		"size":     len(c.data),
		"maxSize":  c.maxSize,
		"hits":     c.hits,
		"misses":   c.misses,
		"cleanups": c.cleanups,
	}
}
