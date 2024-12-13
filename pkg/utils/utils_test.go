package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseInterceptor(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		w := httptest.NewRecorder()
		ri := NewResponseInterceptor(w)

		// Test initial state
		if ri.Status != 200 {
			t.Errorf("Expected initial status 200, got %d", ri.Status)
		}
		if len(ri.HeaderMap) != 0 {
			t.Error("Expected empty initial headers")
		}
	})

	t.Run("header operations", func(t *testing.T) {
		w := httptest.NewRecorder()
		ri := NewResponseInterceptor(w)

		// Set header
		ri.Header().Set("Content-Type", "text/plain")
		if ri.HeaderMap.Get("Content-Type") != "text/plain" {
			t.Error("Expected Content-Type header to be set")
		}

		// Add header
		ri.Header().Add("X-Test", "value1")
		ri.Header().Add("X-Test", "value2")
		values := ri.HeaderMap["X-Test"]
		if len(values) != 2 || values[0] != "value1" || values[1] != "value2" {
			t.Error("Expected multiple header values to be set")
		}
	})

	t.Run("write operations", func(t *testing.T) {
		w := httptest.NewRecorder()
		ri := NewResponseInterceptor(w)

		// Write content
		content := []byte("test content")
		n, err := ri.Write(content)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if n != len(content) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(content), n)
		}
		if string(ri.Body) != "test content" {
			t.Errorf("Expected body 'test content', got '%s'", string(ri.Body))
		}
	})

	t.Run("status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		ri := NewResponseInterceptor(w)

		// Set status code
		ri.WriteHeader(http.StatusNotFound)
		if ri.Status != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, ri.Status)
		}
	})

	t.Run("multiple writes", func(t *testing.T) {
		w := httptest.NewRecorder()
		ri := NewResponseInterceptor(w)

		// Write multiple times
		ri.Write([]byte("first "))
		ri.Write([]byte("second"))

		// Only the last write should be stored
		if string(ri.Body) != "second" {
			t.Errorf("Expected body 'second', got '%s'", string(ri.Body))
		}
	})
}

func TestSSRResponse(t *testing.T) {
	t.Run("success response", func(t *testing.T) {
		resp := SSRResponse{
			HTML: "<div>test</div>",
			Code: 200,
		}

		if resp.Error != "" {
			t.Error("Expected no error in success response")
		}
		if resp.Code != 200 {
			t.Errorf("Expected code 200, got %d", resp.Code)
		}
		if resp.HTML != "<div>test</div>" {
			t.Errorf("Expected HTML '<div>test</div>', got %q", resp.HTML)
		}
	})

	t.Run("error response", func(t *testing.T) {
		resp := SSRResponse{
			Error: "test error",
			Code:  500,
		}

		if resp.Error != "test error" {
			t.Errorf("Expected error 'test error', got %q", resp.Error)
		}
		if resp.Code != 500 {
			t.Errorf("Expected code 500, got %d", resp.Code)
		}
		if resp.HTML != "" {
			t.Error("Expected empty HTML in error response")
		}
	})

	t.Run("not found response", func(t *testing.T) {
		resp := SSRResponse{
			HTML: "<div>404 Not Found</div>",
			Code: 404,
		}

		if resp.Error != "" {
			t.Error("Expected no error in not found response")
		}
		if resp.Code != 404 {
			t.Errorf("Expected code 404, got %d", resp.Code)
		}
		if resp.HTML != "<div>404 Not Found</div>" {
			t.Errorf("Expected HTML '<div>404 Not Found</div>', got %q", resp.HTML)
		}
	})

	t.Run("zero values", func(t *testing.T) {
		var resp SSRResponse

		if resp.Error != "" {
			t.Error("Expected empty error string")
		}
		if resp.Code != 0 {
			t.Errorf("Expected code 0, got %d", resp.Code)
		}
		if resp.HTML != "" {
			t.Error("Expected empty HTML string")
		}
	})
}
