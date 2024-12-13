package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/devthefuture-org/blastra/pkg/health"
	"golang.org/x/time/rate"
)

func TestRoutes(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "static-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test static file
	testContent := []byte("test static content")
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create test config with HealthChecker
	config := &Config{
		StaticDir:     ".",
		BlastraCWD:    tmpDir,
		HealthChecker: health.NewHealthChecker(),
		SSRHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("SSR content"))
		}),
	}

	// Create RouteConfig wrapping the Config
	routeConfig := &RouteConfig{
		Config:     config,
		TrustProxy: false,
	}

	// Create rate limiter (1 request per second)
	limiter := rate.NewLimiter(rate.Limit(1), 1)

	// Setup routes
	mux := http.NewServeMux()
	SetupRoutes(mux, routeConfig, limiter)

	// Mark service as ready for health checks
	config.HealthChecker.SetReady()

	t.Run("static file serving", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test.txt", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK for static file, got %v", w.Code)
		}
		if w.Body.String() != "test static content" {
			t.Errorf("Expected 'test static content', got '%s'", w.Body.String())
		}
	})

	t.Run("SSR rate limiting", func(t *testing.T) {
		// Set a consistent IP address for testing
		testIP := "192.0.2.1"

		// First SSR request should succeed
		req := httptest.NewRequest("GET", "/some-ssr-path", nil)
		req.RemoteAddr = testIP + ":12345"
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected first SSR request to succeed with status OK, got %v", w.Code)
		}
		if w.Body.String() != "SSR content" {
			t.Errorf("Expected 'SSR content', got '%s'", w.Body.String())
		}

		// Second immediate SSR request from same IP should be rate limited
		req = httptest.NewRequest("GET", "/another-ssr-path", nil)
		req.RemoteAddr = testIP + ":12346" // Different port, same IP
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// The second request should still succeed because uber/ratelimit is token bucket based
		// and doesn't return errors, it just delays the request
		if w.Code != http.StatusOK {
			t.Errorf("Expected second SSR request to succeed with status OK, got %v", w.Code)
		}

		// Third request from a different IP should succeed immediately
		req = httptest.NewRequest("GET", "/yet-another-ssr-path", nil)
		req.RemoteAddr = "192.0.2.2:12345" // Different IP
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected request from different IP to succeed with status OK, got %v", w.Code)
		}
	})

	t.Run("static files bypass rate limiting", func(t *testing.T) {
		// Make multiple quick requests to static file
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test.txt", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected static file request %d to succeed with status OK, got %v", i+1, w.Code)
			}
			if w.Body.String() != "test static content" {
				t.Errorf("Expected 'test static content', got '%s'", w.Body.String())
			}
		}
	})

	t.Run("health checks bypass rate limiting", func(t *testing.T) {
		// Make multiple quick requests to health endpoints
		endpoints := []string{"/live", "/ready"}
		for _, endpoint := range endpoints {
			for i := 0; i < 5; i++ {
				req := httptest.NewRequest("GET", endpoint, nil)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected health check %s request %d to succeed with status OK, got %v", endpoint, i+1, w.Code)
				}
			}
		}
	})
}
