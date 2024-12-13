package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/devthefuture-org/blastra/pkg/cache"
)

// mockSSRCommand creates a mock SSR command that returns JSON response
func mockSSRCommand(t *testing.T, content string, statusCode int) []string {
	// Create temporary directory for test script
	tempDir, err := os.MkdirTemp("", "ssr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	// Create test script
	scriptContent := fmt.Sprintf(`#!/bin/sh
printf '{"html":"%s","code":%d}'
`, content, statusCode)

	scriptPath := filepath.Join(tempDir, "test-ssr.sh")
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	return []string{scriptPath}
}

func TestSSRHandler(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		// Create memory cache with test content
		memCache := cache.NewSSRInMemoryCache(cache.CacheConfig{
			TTL:     time.Minute,
			MaxSize: 10,
		})
		provider := cache.NewCacheProvider(memCache, nil)

		testContent := []byte("cached content")
		provider.Set("/test", testContent)

		// Create handler
		handler := SSRHandler(provider, nil, mockSSRCommand(t, "test content", http.StatusOK), 60, ".", nil)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Error("Expected HTML content type")
		}
		if w.Header().Get("Cache-Control") == "" {
			t.Error("Expected Cache-Control header")
		}
		if w.Header().Get("ETag") == "" {
			t.Error("Expected ETag header")
		}
		if string(w.Body.Bytes()) != string(testContent) {
			t.Errorf("Expected body %s, got %s", testContent, w.Body.String())
		}
	})

	t.Run("404 cache hit", func(t *testing.T) {
		// Create notfound cache with test content
		notFoundCache := cache.NewNotFoundInMemoryCache(cache.CacheConfig{
			TTL:     time.Minute,
			MaxSize: 10,
		})
		provider := cache.NewCacheProvider(notFoundCache, nil)

		testContent := []byte("404 content")
		provider.Set("/notfound", testContent)

		// Create handler
		handler := SSRHandler(nil, provider, mockSSRCommand(t, "404 not found", http.StatusNotFound), 60, ".", nil)

		// Create test request
		req := httptest.NewRequest("GET", "/notfound", nil)
		w := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(w, req)

		// Verify response
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
		if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Error("Expected HTML content type")
		}
		if w.Header().Get("Cache-Control") == "" {
			t.Error("Expected Cache-Control header")
		}
		if w.Header().Get("ETag") == "" {
			t.Error("Expected ETag header")
		}
		if string(w.Body.Bytes()) != string(testContent) {
			t.Errorf("Expected body %s, got %s", testContent, w.Body.String())
		}
	})

	t.Run("304 not modified", func(t *testing.T) {
		// Create memory cache with test content
		memCache := cache.NewSSRInMemoryCache(cache.CacheConfig{
			TTL:     time.Minute,
			MaxSize: 10,
		})
		provider := cache.NewCacheProvider(memCache, nil)

		testContent := []byte("cached content")
		provider.Set("/test", testContent)

		// Get ETag from cache
		entry, _ := provider.Get("/test")
		etag := entry.ETag

		// Create handler
		handler := SSRHandler(provider, nil, mockSSRCommand(t, "test content", http.StatusOK), 60, ".", nil)

		// Create test request with If-None-Match header
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("If-None-Match", etag)
		w := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(w, req)

		// Verify response
		if w.Code != http.StatusNotModified {
			t.Errorf("Expected status %d, got %d", http.StatusNotModified, w.Code)
		}
		if w.Body.Len() != 0 {
			t.Error("Expected empty body for 304 response")
		}
	})

	t.Run("worker fallback", func(t *testing.T) {
		// Create handler with mock command
		handler := SSRHandler(nil, nil, mockSSRCommand(t, "test content", http.StatusOK), 60, ".", nil)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Error("Expected HTML content type")
		}
		if w.Body.String() != "test content" {
			t.Errorf("Expected body %s got %s", "test content", w.Body.String())
		}
	})

	t.Run("caching disabled", func(t *testing.T) {
		// Create handler with no caches
		handler := SSRHandler(nil, nil, mockSSRCommand(t, "test content", http.StatusOK), 60, ".", nil)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Error("Expected HTML content type")
		}
		if w.Body.String() != "test content" {
			t.Errorf("Expected body %s got %s", "test content", w.Body.String())
		}
	})
}
