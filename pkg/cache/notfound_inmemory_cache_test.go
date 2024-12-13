package cache

import (
	"testing"
	"time"
)

func TestNotFoundInMemoryCache(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		cache := NewNotFoundInMemoryCache(CacheConfig{
			TTL:     time.Second,
			MaxSize: 10,
		})

		// Test Set and Get
		content := []byte("404 not found content")
		cache.Set("key1", content)

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

	t.Run("capacity limit", func(t *testing.T) {
		cache := NewNotFoundInMemoryCache(CacheConfig{
			TTL:     time.Second,
			MaxSize: 2,
		})

		// Fill cache
		cache.Set("key1", []byte("404 content1"))
		cache.Set("key2", []byte("404 content2"))

		// Add one more, should evict oldest
		cache.Set("key3", []byte("404 content3"))

		// Check key1 was evicted
		_, found := cache.Get("key1")
		if found {
			t.Error("Expected key1 to be evicted")
		}

		// Check key2 and key3 are still there
		_, found = cache.Get("key2")
		if !found {
			t.Error("Expected key2 to be present")
		}
		_, found = cache.Get("key3")
		if !found {
			t.Error("Expected key3 to be present")
		}
	})

	t.Run("cleanup", func(t *testing.T) {
		cache := NewNotFoundInMemoryCache(CacheConfig{
			TTL:     50 * time.Millisecond,
			MaxSize: 10,
		})

		cache.Set("key1", []byte("404 content1"))

		// Wait for TTL to expire
		time.Sleep(100 * time.Millisecond)

		// Trigger cleanup
		cache.cleanup()

		// Check entry was removed
		_, found := cache.Get("key1")
		if found {
			t.Error("Expected entry to be cleaned up")
		}
	})

	t.Run("metrics", func(t *testing.T) {
		cache := NewNotFoundInMemoryCache(CacheConfig{
			TTL:     time.Second,
			MaxSize: 10,
		})

		// Generate some hits and misses
		cache.Set("key1", []byte("404 content1"))
		cache.Get("key1") // Hit
		cache.Get("key2") // Miss

		metrics := cache.GetMetrics()

		if metrics["type"] != "memory" {
			t.Error("Expected type to be 'memory'")
		}
		if metrics["hits"].(int64) != 1 {
			t.Errorf("Expected 1 hit, got %d", metrics["hits"])
		}
		if metrics["misses"].(int64) != 1 {
			t.Errorf("Expected 1 miss, got %d", metrics["misses"])
		}
		if metrics["size"].(int) != 1 {
			t.Errorf("Expected size 1, got %d", metrics["size"])
		}
	})

	t.Run("etag generation", func(t *testing.T) {
		cache := NewNotFoundInMemoryCache(CacheConfig{
			TTL:     time.Second,
			MaxSize: 10,
		})

		// Set same content twice, should get same ETag
		content := []byte("404 not found content")
		cache.Set("key1", content)
		cache.Set("key2", content)

		entry1, _ := cache.Get("key1")
		entry2, _ := cache.Get("key2")

		if entry1.ETag != entry2.ETag {
			t.Error("Expected same ETag for same content")
		}

		// Set different content, should get different ETag
		cache.Set("key3", []byte("different 404 content"))
		entry3, _ := cache.Get("key3")

		if entry1.ETag == entry3.ETag {
			t.Error("Expected different ETag for different content")
		}
	})

	t.Run("default size", func(t *testing.T) {
		cache := NewNotFoundInMemoryCache(CacheConfig{
			TTL:     time.Second,
			MaxSize: 0, // Should use default size
		})

		metrics := cache.GetMetrics()
		if metrics["maxSize"].(int) != 250 {
			t.Errorf("Expected default size 250, got %d", metrics["maxSize"])
		}
	})
}
