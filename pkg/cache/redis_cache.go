package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type RedisCache struct {
	client  *redis.Client
	ttl     time.Duration
	metrics struct {
		hits   int64
		misses int64
	}
}

func NewRedisCache(config ExternalCacheConfig) (*RedisCache, error) {
	if config.Type != ExternalCacheRedis {
		return nil, ErrInvalidCacheType
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisURL,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		ttl:    config.TTL,
	}, nil
}

func (c *RedisCache) prefixKey(key string) string {
	return "blastra:" + key
}

func (c *RedisCache) Get(key string) (CacheEntry, bool) {
	ctx := context.Background()
	data, err := c.client.Get(ctx, c.prefixKey(key)).Bytes()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("Redis get error: %v", err)
		}
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

func (c *RedisCache) Set(key string, content []byte) {
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

	ctx := context.Background()
	if err := c.client.Set(ctx, c.prefixKey(key), data, c.ttl).Err(); err != nil {
		log.Errorf("Redis set error: %v", err)
	}
}

func (c *RedisCache) GetMetrics() map[string]interface{} {
	ctx := context.Background()
	dbSize, err := c.client.DBSize(ctx).Result()
	if err != nil {
		log.Errorf("Failed to get Redis DB size: %v", err)
		dbSize = -1
	}

	return map[string]interface{}{
		"type":   "redis",
		"size":   dbSize,
		"hits":   c.metrics.hits,
		"misses": c.metrics.misses,
	}
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
