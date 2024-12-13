package cache

import (
	"fmt"
)

// NewExternalCache creates a new external cache based on the provided configuration
func NewExternalCache(config ExternalCacheConfig) (Cache, error) {
	switch config.Type {
	case ExternalCacheNone:
		return nil, nil
	case ExternalCacheRedis:
		return NewRedisCache(config)
	case ExternalCacheFilesystem:
		return NewFilesystemCache(config)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidCacheType, config.Type)
	}
}

// CreateCacheProvider creates a new CacheProvider with the specified memory and external caches
func CreateCacheProvider(memoryCache Cache, externalConfig ExternalCacheConfig) (*CacheProvider, error) {
	externalCache, err := NewExternalCache(externalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create external cache: %w", err)
	}

	return NewCacheProvider(memoryCache, externalCache), nil
}
