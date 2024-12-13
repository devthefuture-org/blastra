package cache

import (
	"testing"
	"time"
)

func TestCacheProvider(t *testing.T) {
	t.Run("memory cache only", func(t *testing.T) {
		memCache := NewSSRInMemoryCache(CacheConfig{
			TTL:     time.Second,
			MaxSize: 10,
		})
		provider := NewCacheProvider(memCache, nil)

		// Test Set and Get
		content := []byte("test content")
		provider.Set("key1", content)

		entry, found := provider.Get("key1")
		if !found {
			t.Error("Expected to find entry in memory cache")
		}
		if string(entry.Content) != string(content) {
			t.Errorf("Expected content %s, got %s", content, entry.Content)
		}

		// Test metrics
		metrics := provider.GetMetrics()
		if metrics["memory"] == nil {
			t.Error("Expected memory metrics to be present")
		}
		if metrics["external"] != nil {
			t.Error("Expected no external metrics")
		}
	})

	t.Run("memory and external cache", func(t *testing.T) {
		memCache := NewSSRInMemoryCache(CacheConfig{
			TTL:     time.Second,
			MaxSize: 10,
		})
		externalCache := NewSSRInMemoryCache(CacheConfig{ // Using memory cache as external for testing
			TTL:     time.Second,
			MaxSize: 10,
		})
		provider := NewCacheProvider(memCache, externalCache)

		// Test Set propagates to both caches
		content := []byte("test content")
		provider.Set("key1", content)

		// Check memory cache
		memEntry, memFound := memCache.Get("key1")
		if !memFound {
			t.Error("Expected to find entry in memory cache")
		}
		if string(memEntry.Content) != string(content) {
			t.Errorf("Expected content %s, got %s", content, memEntry.Content)
		}

		// Check external cache
		extEntry, extFound := externalCache.Get("key1")
		if !extFound {
			t.Error("Expected to find entry in external cache")
		}
		if string(extEntry.Content) != string(content) {
			t.Errorf("Expected content %s, got %s", content, extEntry.Content)
		}

		// Test memory cache miss, external cache hit
		memCache.data = make(map[string]CacheEntry) // Clear memory cache
		entry, found := provider.Get("key1")
		if !found {
			t.Error("Expected to find entry via external cache")
		}
		if string(entry.Content) != string(content) {
			t.Errorf("Expected content %s, got %s", content, entry.Content)
		}

		// Verify memory cache was populated from external
		memEntry, memFound = memCache.Get("key1")
		if !memFound {
			t.Error("Expected memory cache to be populated from external")
		}
		if string(memEntry.Content) != string(content) {
			t.Errorf("Expected content %s, got %s", content, memEntry.Content)
		}
	})

	t.Run("metrics", func(t *testing.T) {
		memCache := NewSSRInMemoryCache(CacheConfig{
			TTL:     time.Second,
			MaxSize: 10,
		})
		externalCache := NewSSRInMemoryCache(CacheConfig{
			TTL:     time.Second,
			MaxSize: 10,
		})
		provider := NewCacheProvider(memCache, externalCache)

		// Generate some hits and misses
		provider.Set("key1", []byte("content1"))
		provider.Get("key1") // Hit
		provider.Get("key2") // Miss

		metrics := provider.GetMetrics()
		memMetrics := metrics["memory"].(map[string]interface{})
		extMetrics := metrics["external"].(map[string]interface{})

		if memMetrics["hits"].(int64) != 1 {
			t.Errorf("Expected 1 memory cache hit, got %d", memMetrics["hits"])
		}
		if memMetrics["misses"].(int64) != 1 {
			t.Errorf("Expected 1 memory cache miss, got %d", memMetrics["misses"])
		}
		if extMetrics["hits"].(int64) != 0 {
			t.Errorf("Expected 0 external cache hits, got %d", extMetrics["hits"])
		}
		if extMetrics["misses"].(int64) != 1 {
			t.Errorf("Expected 1 external cache miss, got %d", extMetrics["misses"])
		}
	})
}
