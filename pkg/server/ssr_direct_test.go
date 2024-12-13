package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/devthefuture-org/blastra/pkg/cache"
)

func createTestScript(t *testing.T, dir string) string {
	scriptContent := `#!/bin/sh
echo '{"html": "test content", "code": 200}'
`
	scriptPath := filepath.Join(dir, "test-ssr.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}
	return scriptPath
}

func create404Script(t *testing.T, dir string) string {
	scriptContent := `#!/bin/sh
echo '{"html": "404 not found", "code": 404}'
`
	scriptPath := filepath.Join(dir, "test-404.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create 404 script: %v", err)
	}
	return scriptPath
}

func createErrorScript(t *testing.T, dir string) string {
	scriptContent := `#!/bin/sh
echo '{"error": "test error", "code": 500}'
`
	scriptPath := filepath.Join(dir, "test-error.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create error script: %v", err)
	}
	return scriptPath
}

func TestHandleDirectSSR(t *testing.T) {
	// Create temporary directory for test scripts
	tempDir, err := os.MkdirTemp("", "ssr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("successful response", func(t *testing.T) {
		scriptPath := createTestScript(t, tempDir)

		// Create caches
		ssrCache := cache.NewCacheProvider(
			cache.NewSSRInMemoryCache(cache.CacheConfig{TTL: time.Minute}),
			nil,
		)
		notFoundCache := cache.NewCacheProvider(
			cache.NewNotFoundInMemoryCache(cache.CacheConfig{TTL: time.Minute}),
			nil,
		)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		handleDirectSSR(w, req, []string{scriptPath}, tempDir, ssrCache, notFoundCache, "/test", 60)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Error("Expected HTML content type")
		}
		if w.Body.String() != "test content" {
			t.Errorf("Expected body %s, got %s", "test content", w.Body.String())
		}

		// Verify content was cached
		entry, found := ssrCache.Get("/test")
		if !found {
			t.Error("Expected content to be cached")
		}
		if string(entry.Content) != "test content" {
			t.Errorf("Expected cached content %s, got %s", "test content", entry.Content)
		}
	})

	t.Run("404 response", func(t *testing.T) {
		scriptPath := create404Script(t, tempDir)

		// Create caches
		ssrCache := cache.NewCacheProvider(
			cache.NewSSRInMemoryCache(cache.CacheConfig{TTL: time.Minute}),
			nil,
		)
		notFoundCache := cache.NewCacheProvider(
			cache.NewNotFoundInMemoryCache(cache.CacheConfig{TTL: time.Minute}),
			nil,
		)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		handleDirectSSR(w, req, []string{scriptPath}, tempDir, ssrCache, notFoundCache, "/test", 60)

		// Verify response
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
		if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Error("Expected HTML content type")
		}
		if w.Body.String() != "404 not found" {
			t.Errorf("Expected body %s, got %s", "404 not found", w.Body.String())
		}

		// Verify content was cached in notFoundCache
		entry, found := notFoundCache.Get("/test")
		if !found {
			t.Error("Expected content to be cached in notFoundCache")
		}
		if string(entry.Content) != "404 not found" {
			t.Errorf("Expected cached content %s, got %s", "404 not found", entry.Content)
		}
	})

	t.Run("error response", func(t *testing.T) {
		scriptPath := createErrorScript(t, tempDir)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		handleDirectSSR(w, req, []string{scriptPath}, tempDir, nil, nil, "/test", 60)

		// Verify response
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
		if w.Body.String() != "test error\n" {
			t.Errorf("Expected body %s, got %s", "test error", w.Body.String())
		}
	})

	t.Run("invalid script", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request with non-existent script
		handleDirectSSR(w, req, []string{"/nonexistent"}, tempDir, nil, nil, "/test", 60)

		// Verify response
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("invalid json response", func(t *testing.T) {
		scriptContent := `#!/bin/sh
echo 'invalid json'
`
		scriptPath := filepath.Join(tempDir, "invalid-json.sh")
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create invalid json script: %v", err)
		}

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		handleDirectSSR(w, req, []string{scriptPath}, tempDir, nil, nil, "/test", 60)

		// Verify response
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("caching disabled", func(t *testing.T) {
		scriptPath := createTestScript(t, tempDir)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request without caches
		handleDirectSSR(w, req, []string{scriptPath}, tempDir, nil, nil, "/test", 60)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Error("Expected HTML content type")
		}
		if w.Body.String() != "test content" {
			t.Errorf("Expected body %s, got %s", "test content", w.Body.String())
		}
	})
}
