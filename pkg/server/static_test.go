package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestStaticFileHandler(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "static-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Get absolute path
	tmpDir, err = filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create a test file
	testContent := []byte("Hello, World!")
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a medium test file (5MB)
	mediumContent := make([]byte, 5*1024*1024)
	for i := range mediumContent {
		mediumContent[i] = byte(i % 256)
	}
	mediumFile := filepath.Join(tmpDir, "medium.bin")
	if err := os.WriteFile(mediumFile, mediumContent, 0644); err != nil {
		t.Fatalf("Failed to create medium file: %v", err)
	}

	// Create a large test file (50MB)
	largeFile := filepath.Join(tmpDir, "large.bin")
	f, err := os.Create(largeFile)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	// Write 50MB in chunks to avoid memory pressure
	chunk := make([]byte, 1024*1024) // 1MB chunk
	for i := 0; i < 50; i++ {
		for j := range chunk {
			chunk[j] = byte((i + j) % 256)
		}
		if _, err := f.Write(chunk); err != nil {
			f.Close()
			t.Fatalf("Failed to write to large file: %v", err)
		}
	}
	f.Close()

	config := &Config{
		StaticDir:  ".",    // Use current directory as base
		BlastraCWD: tmpDir, // Use temp dir as CWD
	}

	handler := CreateFileServer(config)

	// Create test server with our handler
	ts := httptest.NewUnstartedServer(handler)
	ts.Config.ReadTimeout = 5 * time.Second
	ts.Config.WriteTimeout = 30 * time.Minute
	ts.Config.IdleTimeout = 120 * time.Second
	ts.Start()
	defer ts.Close()

	// Helper function to verify content
	verifyContent := func(t *testing.T, url string, start, length int64) {
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		if length > 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, start+length-1))
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if length > 0 {
			if resp.StatusCode != http.StatusPartialContent {
				t.Errorf("Expected status 206, got %v", resp.Status)
			}
		} else {
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %v", resp.Status)
			}
		}

		// Verify connection headers
		if resp.Header.Get("Connection") != "keep-alive" {
			t.Error("Expected Connection: keep-alive header")
		}
		if resp.Header.Get("Keep-Alive") == "" {
			t.Error("Expected Keep-Alive header")
		}

		// Read response in chunks and verify each chunk
		chunk := make([]byte, 64*1024) // 64KB chunks
		var totalRead int64
		expectedFile, err := os.Open(filepath.Join(tmpDir, filepath.Base(req.URL.Path)))
		if err != nil {
			t.Fatalf("Failed to open expected file: %v", err)
		}
		defer expectedFile.Close()

		// Seek to start position if range request
		if length > 0 {
			if _, err := expectedFile.Seek(start, io.SeekStart); err != nil {
				t.Fatalf("Failed to seek in expected file: %v", err)
			}
		}

		for {
			n, err := resp.Body.Read(chunk)
			if n > 0 {
				totalRead += int64(n)

				// Read same amount from expected file
				expected := make([]byte, n)
				if _, err := io.ReadFull(expectedFile, expected); err != nil {
					t.Fatalf("Failed to read from expected file: %v", err)
				}

				if !bytes.Equal(chunk[:n], expected) {
					t.Error("Content mismatch")
					break
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Error reading response: %v", err)
			}
		}

		// Verify total length
		if length > 0 {
			if totalRead != length {
				t.Errorf("Expected body length %d, got %d", length, totalRead)
			}
		} else {
			info, err := expectedFile.Stat()
			if err != nil {
				t.Fatalf("Failed to stat expected file: %v", err)
			}
			if totalRead != info.Size() {
				t.Errorf("Expected body length %d, got %d", info.Size(), totalRead)
			}
		}
	}

	t.Run("small file", func(t *testing.T) {
		verifyContent(t, ts.URL+"/test.txt", 0, 0)
	})

	t.Run("medium file full", func(t *testing.T) {
		verifyContent(t, ts.URL+"/medium.bin", 0, 0)
	})

	t.Run("medium file range", func(t *testing.T) {
		verifyContent(t, ts.URL+"/medium.bin", 1024*1024, 1024*1024)
	})

	t.Run("large file range start", func(t *testing.T) {
		// Read first 5MB
		verifyContent(t, ts.URL+"/large.bin", 0, 5*1024*1024)
	})

	t.Run("large file range middle", func(t *testing.T) {
		// Read 5MB from the middle
		verifyContent(t, ts.URL+"/large.bin", 25*1024*1024, 5*1024*1024)
	})

	t.Run("large file range end", func(t *testing.T) {
		// Read last 5MB
		verifyContent(t, ts.URL+"/large.bin", 45*1024*1024, 5*1024*1024)
	})

	t.Run("concurrent requests", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				start := int64(i * 1024 * 1024)
				length := int64(1024 * 1024)
				verifyContent(t, ts.URL+"/large.bin", start, length)
			}(i)
		}
		wg.Wait()
	})
}
