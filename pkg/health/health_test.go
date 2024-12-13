package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthChecker(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		checker := NewHealthChecker()
		if checker.IsReady() {
			t.Error("Expected initial state to be not ready")
		}

		// Test liveness probe
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		checker.LivenessProbeHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Body.String() != "OK" {
			t.Errorf("Expected body OK, got %s", w.Body.String())
		}
	})

	t.Run("readiness state", func(t *testing.T) {
		checker := NewHealthChecker()

		// Test readiness probe before ready
		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()
		checker.ReadinessProbeHandler(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
		}
		if w.Body.String() != "Not Ready\n" {
			t.Errorf("Expected body 'Not Ready', got %s", w.Body.String())
		}

		// Set ready
		checker.SetReady()

		if !checker.IsReady() {
			t.Error("Expected state to be ready after SetReady")
		}

		// Test readiness probe after ready
		w = httptest.NewRecorder()
		checker.ReadinessProbeHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Body.String() != "Ready" {
			t.Errorf("Expected body Ready, got %s", w.Body.String())
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		checker := NewHealthChecker()
		done := make(chan bool)

		// Start multiple goroutines to test concurrent access
		for i := 0; i < 10; i++ {
			go func() {
				checker.SetReady()
				checker.IsReady()
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		if !checker.IsReady() {
			t.Error("Expected state to be ready after concurrent access")
		}
	})

	t.Run("probe handlers under load", func(t *testing.T) {
		checker := NewHealthChecker()
		done := make(chan bool)

		// Test concurrent probe requests
		for i := 0; i < 10; i++ {
			go func() {
				// Test liveness probe
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()
				checker.LivenessProbeHandler(w, req)
				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				}

				// Test readiness probe
				req = httptest.NewRequest("GET", "/ready", nil)
				w = httptest.NewRecorder()
				checker.ReadinessProbeHandler(w, req)
				// Status could be either OK or ServiceUnavailable depending on timing
				if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
					t.Errorf("Expected status %d or %d, got %d",
						http.StatusOK, http.StatusServiceUnavailable, w.Code)
				}

				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
