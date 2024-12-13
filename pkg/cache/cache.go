package cache

import (
	"errors"
	"time"
)

var (
	ErrInvalidCacheType = errors.New("invalid cache type")
)

// CacheEntry represents a generic cache entry that can be used by any cache implementation
type CacheEntry struct {
	Content     []byte
	LastUpdated time.Time
	ETag        string
}

// Cache defines the interface that all cache implementations must satisfy
type Cache interface {
	Get(key string) (CacheEntry, bool)
	Set(key string, content []byte)
	GetMetrics() map[string]interface{}
}

// ExternalCacheType represents the type of external cache to use
type ExternalCacheType string

const (
	ExternalCacheNone       ExternalCacheType = "none"
	ExternalCacheRedis      ExternalCacheType = "redis"
	ExternalCacheFilesystem ExternalCacheType = "filesystem"
)

// CacheConfig represents configuration for any cache implementation
type CacheConfig struct {
	TTL     time.Duration
	MaxSize int
}

// ExternalCacheConfig represents configuration specific to external caches
type ExternalCacheConfig struct {
	CacheConfig
	Type ExternalCacheType

	// Redis specific config
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// Filesystem specific config
	CacheDir string
}

// CacheProvider manages the cache hierarchy
type CacheProvider struct {
	memoryCache   Cache
	externalCache Cache
}

func NewCacheProvider(memoryCache Cache, externalCache Cache) *CacheProvider {
	return &CacheProvider{
		memoryCache:   memoryCache,
		externalCache: externalCache,
	}
}

// Get retrieves an entry from the cache hierarchy
// First checks memory cache, then external cache if available
func (p *CacheProvider) Get(key string) (CacheEntry, bool) {
	// Check memory cache first if available
	if p.memoryCache != nil {
		if entry, found := p.memoryCache.Get(key); found {
			return entry, true
		}
	}

	// Check external cache if available
	if p.externalCache != nil {
		if entry, found := p.externalCache.Get(key); found {
			// Store in memory cache if available
			if p.memoryCache != nil {
				p.memoryCache.Set(key, entry.Content)
			}
			return entry, true
		}
	}

	return CacheEntry{}, false
}

// Set stores an entry in all available caches
func (p *CacheProvider) Set(key string, content []byte) {
	if p.memoryCache != nil {
		p.memoryCache.Set(key, content)
	}
	if p.externalCache != nil {
		p.externalCache.Set(key, content)
	}
}

// GetMetrics returns combined metrics from all caches
func (p *CacheProvider) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	if p.memoryCache != nil {
		metrics["memory"] = p.memoryCache.GetMetrics()
	}
	if p.externalCache != nil {
		metrics["external"] = p.externalCache.GetMetrics()
	}

	return metrics
}
