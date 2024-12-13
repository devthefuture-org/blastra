package middleware

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	t.Run("gzip enabled with accept-encoding", func(t *testing.T) {
		middleware := GzipMiddleware(true)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("test content"))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Verify response headers
		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Expected Content-Encoding to be gzip")
			return
		}
		if w.Header().Get("Content-Type") != "text/plain" {
			t.Error("Expected Content-Type to be preserved")
			return
		}

		// Create a buffer to store the gzipped content
		var buf bytes.Buffer
		buf.Write(w.Body.Bytes())

		// Verify content is gzipped and can be decompressed
		reader, err := gzip.NewReader(&buf)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read gzipped content: %v", err)
		}
		reader.Close()

		if string(content) != "test content" {
			t.Errorf("Expected content 'test content', got '%s'", string(content))
		}
	})

	t.Run("gzip disabled", func(t *testing.T) {
		middleware := GzipMiddleware(false)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("test content"))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Verify no compression when disabled
		if w.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected Content-Encoding not to be gzip when disabled")
		}
		if w.Body.String() != "test content" {
			t.Errorf("Expected uncompressed content 'test content', got '%s'", w.Body.String())
		}
	})

	t.Run("no accept-encoding header", func(t *testing.T) {
		middleware := GzipMiddleware(true)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("test content"))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Verify no compression without Accept-Encoding header
		if w.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected Content-Encoding not to be gzip without Accept-Encoding header")
		}
		if w.Body.String() != "test content" {
			t.Errorf("Expected uncompressed content 'test content', got '%s'", w.Body.String())
		}
	})

	t.Run("concurrent requests", func(t *testing.T) {
		middleware := GzipMiddleware(true)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("test content"))
		}))

		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Accept-Encoding", "gzip")
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)

				// Verify response headers
				if w.Header().Get("Content-Encoding") != "gzip" {
					errChan <- errors.New("Expected Content-Encoding to be gzip")
					return
				}

				// Create a buffer to store the gzipped content
				var buf bytes.Buffer
				buf.Write(w.Body.Bytes())

				// Verify content
				reader, err := gzip.NewReader(&buf)
				if err != nil {
					errChan <- fmt.Errorf("Failed to create gzip reader: %v", err)
					return
				}

				content, err := io.ReadAll(reader)
				if err != nil {
					errChan <- fmt.Errorf("Failed to read gzipped content: %v", err)
					return
				}
				reader.Close()

				if string(content) != "test content" {
					errChan <- fmt.Errorf("Expected content 'test content', got '%s'", string(content))
				}
			}()
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			t.Error(err)
		}
	})
}
