package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devthefuture-org/blastra/pkg/cache"
	"github.com/devthefuture-org/blastra/pkg/worker"
)

// testWorkerPool implements worker.IWorkerPool for testing
type testWorkerPool struct {
	endpoint string
	enabled  bool
}

func newTestWorkerPool(endpoint string, enabled bool) worker.IWorkerPool {
	return &testWorkerPool{
		endpoint: endpoint,
		enabled:  enabled,
	}
}

func (t *testWorkerPool) GetWorkerEndpoint() string {
	if !t.enabled {
		return ""
	}
	return t.endpoint
}

func (t *testWorkerPool) Shutdown() {
	// No-op for testing
}

func TestHandleWorkerSSR(t *testing.T) {
	t.Run("successful worker response", func(t *testing.T) {
		// Create test server to mock worker endpoint
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("worker response"))
		}))
		defer ts.Close()

		// Create worker pool with test endpoint
		wp := newTestWorkerPool(ts.URL, true)

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
		handled := handleWorkerSSR(w, req, wp, ssrCache, notFoundCache, "/test")

		// Verify response
		if !handled {
			t.Error("Expected request to be handled")
		}
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Error("Expected HTML content type")
		}
		if w.Body.String() != "worker response" {
			t.Errorf("Expected body %s, got %s", "worker response", w.Body.String())
		}

		// Verify content was cached
		entry, found := ssrCache.Get("/test")
		if !found {
			t.Error("Expected content to be cached")
		}
		if string(entry.Content) != "worker response" {
			t.Errorf("Expected cached content %s, got %s", "worker response", entry.Content)
		}
	})

	t.Run("404 worker response", func(t *testing.T) {
		// Create test server to mock worker endpoint
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 not found"))
		}))
		defer ts.Close()

		// Create worker pool with test endpoint
		wp := newTestWorkerPool(ts.URL, true)

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
		handled := handleWorkerSSR(w, req, wp, ssrCache, notFoundCache, "/test")

		// Verify response
		if !handled {
			t.Error("Expected request to be handled")
		}
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

	t.Run("no worker endpoint", func(t *testing.T) {
		// Create disabled worker pool
		wp := newTestWorkerPool("", false)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		handled := handleWorkerSSR(w, req, wp, nil, nil, "/test")

		// Verify request was not handled
		if handled {
			t.Error("Expected request not to be handled")
		}
	})

	t.Run("worker error response", func(t *testing.T) {
		// Create test server to mock worker endpoint
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		// Create worker pool with test endpoint
		wp := newTestWorkerPool(ts.URL, true)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		handled := handleWorkerSSR(w, req, wp, nil, nil, "/test")

		// Verify response
		if !handled {
			t.Error("Expected request to be handled")
		}
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}
