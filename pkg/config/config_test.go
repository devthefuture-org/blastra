package config

import (
	"os"
	"testing"
	"time"

	"github.com/devthefuture-org/blastra/pkg/cache"
)

func TestLoadConfiguration(t *testing.T) {
	// Save original env vars
	originalEnv := map[string]string{
		"BLASTRA_HTTP_PORT":           os.Getenv("BLASTRA_HTTP_PORT"),
		"BLASTRA_HTTPS_PORT":          os.Getenv("BLASTRA_HTTPS_PORT"),
		"BLASTRA_ENABLE_HTTPS":        os.Getenv("BLASTRA_ENABLE_HTTPS"),
		"BLASTRA_TLS_CERT_PATH":       os.Getenv("BLASTRA_TLS_CERT_PATH"),
		"BLASTRA_TLS_KEY_PATH":        os.Getenv("BLASTRA_TLS_KEY_PATH"),
		"BLASTRA_STATIC_DIR":          os.Getenv("BLASTRA_STATIC_DIR"),
		"BLASTRA_SSR_SCRIPT":          os.Getenv("BLASTRA_SSR_SCRIPT"),
		"BLASTRA_SSR_CACHE_ENABLED":   os.Getenv("BLASTRA_SSR_CACHE_ENABLED"),
		"BLASTRA_CACHE_TTL":           os.Getenv("BLASTRA_CACHE_TTL"),
		"BLASTRA_CACHE_SIZE":          os.Getenv("BLASTRA_CACHE_SIZE"),
		"BLASTRA_EXTERNAL_CACHE_TYPE": os.Getenv("BLASTRA_EXTERNAL_CACHE_TYPE"),
		"BLASTRA_REDIS_URL":           os.Getenv("BLASTRA_REDIS_URL"),
		"BLASTRA_REDIS_PASSWORD":      os.Getenv("BLASTRA_REDIS_PASSWORD"),
		"BLASTRA_REDIS_DB":            os.Getenv("BLASTRA_REDIS_DB"),
		"BLASTRA_CACHE_DIR":           os.Getenv("BLASTRA_CACHE_DIR"),
		"BLASTRA_RATE_LIMIT":          os.Getenv("BLASTRA_RATE_LIMIT"),
		"BLASTRA_BURST":               os.Getenv("BLASTRA_BURST"),
		"BLASTRA_MAX_AGE_STATIC":      os.Getenv("BLASTRA_MAX_AGE_STATIC"),
		"BLASTRA_MAX_AGE_SSR":         os.Getenv("BLASTRA_MAX_AGE_SSR"),
		"BLASTRA_SHUTDOWN_TIMEOUT":    os.Getenv("BLASTRA_SHUTDOWN_TIMEOUT"),
		"BLASTRA_CWD":                 os.Getenv("BLASTRA_CWD"),
		"BLASTRA_GZIP_ENABLED":        os.Getenv("BLASTRA_GZIP_ENABLED"),
		"BLASTRA_CPU_LIMIT":           os.Getenv("BLASTRA_CPU_LIMIT"),
		"BLASTRA_SSR_WORKERS":         os.Getenv("BLASTRA_SSR_WORKERS"),
	}

	// Cleanup function to restore original env vars
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("default configuration", func(t *testing.T) {
		// Clear all relevant env vars
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		cfg, err := LoadConfiguration()
		if err != nil {
			t.Fatalf("Failed to load default configuration: %v", err)
		}

		// Verify default values
		if cfg.HTTPPort != DefaultHTTPPort {
			t.Errorf("Expected HTTP port %d, got %d", DefaultHTTPPort, cfg.HTTPPort)
		}
		if cfg.HTTPSPort != DefaultHTTPSPort {
			t.Errorf("Expected HTTPS port %d, got %d", DefaultHTTPSPort, cfg.HTTPSPort)
		}
		if cfg.StaticDir != DefaultStaticDir {
			t.Errorf("Expected static dir %s, got %s", DefaultStaticDir, cfg.StaticDir)
		}
		if cfg.CacheTTL != DefaultCacheTTL {
			t.Errorf("Expected cache TTL %v, got %v", DefaultCacheTTL, cfg.CacheTTL)
		}
	})

	t.Run("custom configuration", func(t *testing.T) {
		// Set custom env vars
		os.Setenv("BLASTRA_HTTP_PORT", "3000")
		os.Setenv("BLASTRA_HTTPS_PORT", "3443")
		os.Setenv("BLASTRA_STATIC_DIR", "./public")
		os.Setenv("BLASTRA_SSR_CACHE_ENABLED", "true")
		os.Setenv("BLASTRA_CACHE_TTL", "10m")
		os.Setenv("BLASTRA_CACHE_SIZE", "2000")
		os.Setenv("BLASTRA_EXTERNAL_CACHE_TYPE", "redis")
		os.Setenv("BLASTRA_REDIS_URL", "localhost:6379")
		os.Setenv("BLASTRA_GZIP_ENABLED", "true")

		cfg, err := LoadConfiguration()
		if err != nil {
			t.Fatalf("Failed to load custom configuration: %v", err)
		}

		// Verify custom values
		if cfg.HTTPPort != 3000 {
			t.Errorf("Expected HTTP port 3000, got %d", cfg.HTTPPort)
		}
		if cfg.HTTPSPort != 3443 {
			t.Errorf("Expected HTTPS port 3443, got %d", cfg.HTTPSPort)
		}
		if cfg.StaticDir != "./public" {
			t.Errorf("Expected static dir ./public, got %s", cfg.StaticDir)
		}
		if !cfg.SSRCacheEnabled {
			t.Error("Expected SSR cache to be enabled")
		}
		if cfg.CacheTTL != 10*time.Minute {
			t.Errorf("Expected cache TTL 10m, got %v", cfg.CacheTTL)
		}
		if cfg.CacheSize != 2000 {
			t.Errorf("Expected cache size 2000, got %d", cfg.CacheSize)
		}
		if cfg.ExternalCacheType != cache.ExternalCacheRedis {
			t.Errorf("Expected external cache type redis, got %s", cfg.ExternalCacheType)
		}
		if cfg.RedisURL != "localhost:6379" {
			t.Errorf("Expected Redis URL localhost:6379, got %s", cfg.RedisURL)
		}
		if !cfg.GzipEnabled {
			t.Error("Expected Gzip to be enabled")
		}
	})

	t.Run("invalid configuration", func(t *testing.T) {
		testCases := []struct {
			name    string
			envVars map[string]string
		}{
			{
				name: "invalid HTTP port",
				envVars: map[string]string{
					"BLASTRA_HTTP_PORT": "invalid",
				},
			},
			{
				name: "invalid HTTPS port",
				envVars: map[string]string{
					"BLASTRA_HTTPS_PORT": "invalid",
				},
			},
			{
				name: "invalid cache TTL",
				envVars: map[string]string{
					"BLASTRA_CACHE_TTL": "invalid",
				},
			},
			{
				name: "invalid cache size",
				envVars: map[string]string{
					"BLASTRA_CACHE_SIZE": "invalid",
				},
			},
			{
				name: "HTTPS enabled without cert/key",
				envVars: map[string]string{
					"BLASTRA_ENABLE_HTTPS": "true",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Clear all env vars first
				for key := range originalEnv {
					os.Unsetenv(key)
				}

				// Set test env vars
				for key, value := range tc.envVars {
					os.Setenv(key, value)
				}

				_, err := LoadConfiguration()
				if err == nil {
					t.Error("Expected error for invalid configuration")
				}
			})
		}
	})

	t.Run("cache configuration", func(t *testing.T) {
		// Clear all env vars first
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		os.Setenv("BLASTRA_SSR_CACHE_ENABLED", "true")
		os.Setenv("BLASTRA_CACHE_TTL", "5m")
		os.Setenv("BLASTRA_CACHE_SIZE", "1000")
		os.Setenv("BLASTRA_NOTFOUND_CACHE_TTL", "2m")
		os.Setenv("BLASTRA_NOTFOUND_CACHE_SIZE", "500")
		os.Setenv("BLASTRA_EXTERNAL_CACHE_TYPE", "redis")
		os.Setenv("BLASTRA_REDIS_URL", "localhost:6379")
		os.Setenv("BLASTRA_REDIS_PASSWORD", "secret")
		os.Setenv("BLASTRA_REDIS_DB", "1")

		cfg, err := LoadConfiguration()
		if err != nil {
			t.Fatalf("Failed to load cache configuration: %v", err)
		}

		// Test SSR cache config
		ttl, size := cfg.GetSSRCacheConfig()
		if ttl != 5*time.Minute {
			t.Errorf("Expected SSR cache TTL 5m, got %v", ttl)
		}
		if size != 1000 {
			t.Errorf("Expected SSR cache size 1000, got %d", size)
		}

		// Test NotFound cache config
		notFoundTTL, notFoundSize := cfg.GetNotFoundCacheConfig()
		if notFoundTTL != 2*time.Minute {
			t.Errorf("Expected NotFound cache TTL 2m, got %v", notFoundTTL)
		}
		if notFoundSize != 500 {
			t.Errorf("Expected NotFound cache size 500, got %d", notFoundSize)
		}

		// Test external cache config
		extConfig := cfg.GetExternalCacheConfig()
		if extConfig.Type != cache.ExternalCacheRedis {
			t.Errorf("Expected external cache type redis, got %s", extConfig.Type)
		}
		if extConfig.RedisURL != "localhost:6379" {
			t.Errorf("Expected Redis URL localhost:6379, got %s", extConfig.RedisURL)
		}
		if extConfig.RedisPassword != "secret" {
			t.Errorf("Expected Redis password secret, got %s", extConfig.RedisPassword)
		}
		if extConfig.RedisDB != 1 {
			t.Errorf("Expected Redis DB 1, got %d", extConfig.RedisDB)
		}
	})
}
