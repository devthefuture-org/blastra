package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFilesystemCache(t *testing.T) {
	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("basic operations", func(t *testing.T) {
		cache, err := NewFilesystemCache(ExternalCacheConfig{
			CacheConfig: CacheConfig{
				TTL:     time.Second,
				MaxSize: 10,
			},
			Type:     ExternalCacheFilesystem,
			CacheDir: tempDir,
		})
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Test Set and Get
		content := []byte("test content")
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

	t.Run("cleanup", func(t *testing.T) {
		cleanupDir := filepath.Join(tempDir, "cleanup-test")
		cache, err := NewFilesystemCache(ExternalCacheConfig{
			CacheConfig: CacheConfig{
				TTL:     50 * time.Millisecond,
				MaxSize: 10,
			},
			Type:     ExternalCacheFilesystem,
			CacheDir: cleanupDir,
		})
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		cache.Set("key1", []byte("content1"))

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
		metricsDir := filepath.Join(tempDir, "metrics-test")
		cache, err := NewFilesystemCache(ExternalCacheConfig{
			CacheConfig: CacheConfig{
				TTL:     time.Second,
				MaxSize: 10,
			},
			Type:     ExternalCacheFilesystem,
			CacheDir: metricsDir,
		})
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Generate some hits and misses
		cache.Set("key1", []byte("content1"))
		cache.Get("key1") // Hit
		cache.Get("key2") // Miss

		metrics := cache.GetMetrics()

		if metrics["type"] != "filesystem" {
			t.Error("Expected type to be 'filesystem'")
		}
		if metrics["hits"].(int64) != 1 {
			t.Errorf("Expected 1 hit, got %d", metrics["hits"])
		}
		if metrics["misses"].(int64) != 1 {
			t.Errorf("Expected 1 miss, got %d", metrics["misses"])
		}
		if metrics["size"].(int64) != 1 {
			t.Errorf("Expected size 1, got %d", metrics["size"])
		}
	})

	t.Run("etag generation", func(t *testing.T) {
		etagDir := filepath.Join(tempDir, "etag-test")
		cache, err := NewFilesystemCache(ExternalCacheConfig{
			CacheConfig: CacheConfig{
				TTL:     time.Second,
				MaxSize: 10,
			},
			Type:     ExternalCacheFilesystem,
			CacheDir: etagDir,
		})
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Set same content twice, should get same ETag
		content := []byte("test content")
		cache.Set("key1", content)
		cache.Set("key2", content)

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
		_, err := NewFilesystemCache(ExternalCacheConfig{
			Type:     ExternalCacheRedis,
			CacheDir: tempDir,
		})
		if err != ErrInvalidCacheType {
			t.Error("Expected ErrInvalidCacheType")
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		concurrentDir := filepath.Join(tempDir, "concurrent-test")
		cache, err := NewFilesystemCache(ExternalCacheConfig{
			CacheConfig: CacheConfig{
				TTL:     time.Second,
				MaxSize: 10,
			},
			Type:     ExternalCacheFilesystem,
			CacheDir: concurrentDir,
		})
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(i int) {
				key := "key" + string(rune(i+'0'))
				cache.Set(key, []byte("content"))
				cache.Get(key)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
