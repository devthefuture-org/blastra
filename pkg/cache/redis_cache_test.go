package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisCache(t *testing.T) {
	// Start miniredis server
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer s.Close()

	t.Run("basic operations", func(t *testing.T) {
		cache, err := NewRedisCache(ExternalCacheConfig{
			CacheConfig: CacheConfig{
				TTL:     time.Second,
				MaxSize: 10,
			},
			Type:     ExternalCacheRedis,
			RedisURL: s.Addr(),
		})
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}
		defer cache.Close()

		// Test Set and Get
		content := []byte("test content")
		cache.Set("key1", content)

		// Verify key is stored with prefix
		if !s.Exists("blastra:key1") {
			t.Error("Expected to find prefixed key in Redis")
		}

		entry, found := cache.Get("key1")
		if !found {
			t.Error("Expected to find entry")
		}
		if string(entry.Content) != string(content) {
			t.Errorf("Expected content %s, got %s", content, entry.Content)
		}
		if entry.ETag == "" {
			t.Error("Expected ETag to be generated")
		}

		// Test miss
		_, found = cache.Get("nonexistent")
		if found {
			t.Error("Expected not to find nonexistent entry")
		}
	})

	t.Run("ttl expiration", func(t *testing.T) {
		cache, err := NewRedisCache(ExternalCacheConfig{
			CacheConfig: CacheConfig{
				TTL:     50 * time.Millisecond,
				MaxSize: 10,
			},
			Type:     ExternalCacheRedis,
			RedisURL: s.Addr(),
		})
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}
		defer cache.Close()

		cache.Set("key1", []byte("content1"))

		// Verify key exists with prefix
		if !s.Exists("blastra:key1") {
			t.Error("Expected to find prefixed key in Redis")
		}

		// Verify key exists
		s.FastForward(25 * time.Millisecond)
		if _, found := cache.Get("key1"); !found {
			t.Error("Expected to find entry before TTL expiration")
		}

		// Wait for TTL to expire
		s.FastForward(50 * time.Millisecond)

		// Check entry was removed
		if _, found := cache.Get("key1"); found {
			t.Error("Expected entry to be expired")
		}
	})

	t.Run("metrics", func(t *testing.T) {
		cache, err := NewRedisCache(ExternalCacheConfig{
			CacheConfig: CacheConfig{
				TTL:     time.Second,
				MaxSize: 10,
			},
			Type:     ExternalCacheRedis,
			RedisURL: s.Addr(),
		})
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}
		defer cache.Close()

		// Generate some hits and misses
		cache.Set("key1", []byte("content1"))
		cache.Get("key1") // Hit
		cache.Get("key2") // Miss

		metrics := cache.GetMetrics()

		if metrics["type"] != "redis" {
			t.Error("Expected type to be 'redis'")
		}
		if metrics["hits"].(int64) != 1 {
			t.Errorf("Expected 1 hit, got %d", metrics["hits"])
		}
		if metrics["misses"].(int64) != 1 {
			t.Errorf("Expected 1 miss, got %d", metrics["misses"])
		}
	})

	t.Run("etag generation", func(t *testing.T) {
		cache, err := NewRedisCache(ExternalCacheConfig{
			CacheConfig: CacheConfig{
				TTL:     time.Second,
				MaxSize: 10,
			},
			Type:     ExternalCacheRedis,
			RedisURL: s.Addr(),
		})
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}
		defer cache.Close()

		// Set same content twice, should get same ETag
		content := []byte("test content")
		cache.Set("key1", content)
		cache.Set("key2", content)

		// Verify keys exist with prefix
		if !s.Exists("blastra:key1") || !s.Exists("blastra:key2") {
			t.Error("Expected to find prefixed keys in Redis")
		}

		entry1, _ := cache.Get("key1")
		entry2, _ := cache.Get("key2")

		if entry1.ETag != entry2.ETag {
			t.Error("Expected same ETag for same content")
		}

		// Set different content, should get different ETag
		cache.Set("key3", []byte("different content"))
		entry3, _ := cache.Get("key3")

		if entry1.ETag == entry3.ETag {
			t.Error("Expected different ETag for different content")
		}
	})

	t.Run("invalid cache type", func(t *testing.T) {
		_, err := NewRedisCache(ExternalCacheConfig{
			Type:     ExternalCacheFilesystem,
			RedisURL: s.Addr(),
		})
		if err != ErrInvalidCacheType {
			t.Error("Expected ErrInvalidCacheType")
		}
	})

	t.Run("connection error", func(t *testing.T) {
		_, err := NewRedisCache(ExternalCacheConfig{
			Type:     ExternalCacheRedis,
			RedisURL: "invalid:6379",
		})
		if err == nil {
			t.Error("Expected error for invalid Redis URL")
		}
	})
}
