package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type FilesystemCache struct {
	cacheDir string
	ttl      time.Duration
	mutex    sync.RWMutex
	metrics  struct {
		hits   int64
		misses int64
	}
}

func NewFilesystemCache(config ExternalCacheConfig) (*FilesystemCache, error) {
	if config.Type != ExternalCacheFilesystem {
		return nil, ErrInvalidCacheType
	}

	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, err
	}

	cache := &FilesystemCache{
		cacheDir: config.CacheDir,
		ttl:      config.TTL,
	}

	// Start cleanup routine
	go cache.cleanupRoutine()

	return cache, nil
}

func (c *FilesystemCache) getFilePath(key string) string {
	// Hash the key to create a safe filename
	hasher := sha256.New()
	hasher.Write([]byte(key))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return filepath.Join(c.cacheDir, hash+".cache")
}

func (c *FilesystemCache) Get(key string) (CacheEntry, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	filePath := c.getFilePath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		c.metrics.misses++
		return CacheEntry{}, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		log.Errorf("Failed to unmarshal cache entry: %v", err)
		c.metrics.misses++
		return CacheEntry{}, false
	}

	c.metrics.hits++
	return entry, true
}

func (c *FilesystemCache) Set(key string, content []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Generate ETag
	hasher := sha256.New()
	hasher.Write(content)
	etag := `"` + hex.EncodeToString(hasher.Sum(nil)) + `"`

	entry := CacheEntry{
		Content:     content,
		LastUpdated: time.Now(),
		ETag:        etag,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log.Errorf("Failed to marshal cache entry: %v", err)
		return
	}

	filePath := c.getFilePath(key)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.Errorf("Failed to write cache file: %v", err)
	}
}

func (c *FilesystemCache) cleanupRoutine() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *FilesystemCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		log.Errorf("Failed to read cache directory: %v", err)
		return
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(c.cacheDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > c.ttl {
			if err := os.Remove(filePath); err != nil {
				log.Errorf("Failed to remove expired cache file: %v", err)
			}
		}
	}
}

func (c *FilesystemCache) GetMetrics() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var size int64
	entries, err := os.ReadDir(c.cacheDir)
	if err == nil {
		size = int64(len(entries))
	} else {
		size = -1
	}

	return map[string]interface{}{
		"type":   "filesystem",
		"size":   size,
		"hits":   c.metrics.hits,
		"misses": c.metrics.misses,
	}
}
