package worker

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	startTimeout    = 30 * time.Second
	shutdownTimeout = 5 * time.Second
)

// Mock command that simulates a worker process
const mockScript = `#!/bin/sh
trap 'exit 0' TERM
while true; do sleep 0.1; done
`

func createMockScript(t *testing.T) string {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "mock-worker-*.sh")
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}
	tmpPath := tmpfile.Name()
	if _, err := tmpfile.WriteString(mockScript); err != nil {
		os.Remove(tmpPath)
		t.Fatalf("Failed to write mock script: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		os.Remove(tmpPath)
		t.Fatalf("Failed to close mock script: %v", err)
	}
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		t.Fatalf("Failed to set script permissions: %v", err)
	}
	return tmpPath
}

func waitForWorkerStart(t *testing.T, wp IWorkerPool, timeout time.Duration) bool {
	t.Helper()
	start := time.Now()
	for {
		if time.Since(start) > timeout {
			return false
		}
		if endpoint := wp.GetWorkerEndpoint(); endpoint != "" {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func waitForWorkerShutdown(t *testing.T, wp IWorkerPool, timeout time.Duration) bool {
	t.Helper()
	start := time.Now()
	for {
		if time.Since(start) > timeout {
			return false
		}
		if endpoint := wp.GetWorkerEndpoint(); endpoint == "" {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestWorkerPool(t *testing.T) {
	mockCmd := createMockScript(t)
	defer os.Remove(mockCmd)

	t.Run("disabled worker pool", func(t *testing.T) {
		wp, err := StartWorkerPoolWithCommand(0, ".", mockCmd, nil)
		if err != nil {
			t.Fatalf("Failed to create disabled worker pool: %v", err)
		}
		defer wp.Shutdown()

		if wp.GetWorkerEndpoint() != "" {
			t.Error("Expected empty endpoint for disabled worker pool")
		}
	})

	t.Run("single worker", func(t *testing.T) {
		wp, err := StartWorkerPoolWithCommand(1, ".", mockCmd, nil)
		if err != nil {
			t.Fatalf("Failed to create worker pool: %v", err)
		}
		defer wp.Shutdown()

		if !waitForWorkerStart(t, wp, startTimeout) {
			t.Fatal("Worker failed to start within timeout")
		}

		endpoint := wp.GetWorkerEndpoint()
		if endpoint == "" {
			t.Error("Expected non-empty endpoint for active worker pool")
		}

		if !strings.HasPrefix(endpoint, "http://localhost:") {
			t.Errorf("Expected endpoint to start with http://localhost:, got %s", endpoint)
		}
	})

	t.Run("multiple workers", func(t *testing.T) {
		wp, err := StartWorkerPoolWithCommand(3, ".", mockCmd, nil)
		if err != nil {
			t.Fatalf("Failed to create worker pool: %v", err)
		}
		defer wp.Shutdown()

		if !waitForWorkerStart(t, wp, startTimeout) {
			t.Fatal("Workers failed to start within timeout")
		}

		firstEndpoint := wp.GetWorkerEndpoint()
		endpoints := make(map[string]bool)
		endpoints[firstEndpoint] = true

		// Get more endpoints
		for i := 0; i < 10; i++ {
			endpoint := wp.GetWorkerEndpoint()
			endpoints[endpoint] = true
		}

		if len(endpoints) < 2 {
			t.Error("Expected multiple unique endpoints")
		}
	})

	t.Run("worker shutdown", func(t *testing.T) {
		wp, err := StartWorkerPoolWithCommand(1, ".", mockCmd, nil)
		if err != nil {
			t.Fatalf("Failed to create worker pool: %v", err)
		}

		if !waitForWorkerStart(t, wp, startTimeout) {
			t.Fatal("Worker failed to start within timeout")
		}

		endpoint := wp.GetWorkerEndpoint()
		if endpoint == "" {
			t.Error("Expected non-empty endpoint before shutdown")
		}

		wp.Shutdown()

		if !waitForWorkerShutdown(t, wp, shutdownTimeout) {
			t.Fatal("Worker failed to shut down within timeout")
		}

		endpoint = wp.GetWorkerEndpoint()
		if endpoint != "" {
			t.Error("Expected empty endpoint after shutdown")
		}
	})

	t.Run("invalid worker count", func(t *testing.T) {
		wp, err := StartWorkerPoolWithCommand(-1, ".", mockCmd, nil)
		if err != nil {
			t.Error("Expected no error for invalid worker count")
		}
		if wp.GetWorkerEndpoint() != "" {
			t.Error("Expected empty endpoint for invalid worker count")
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		wp, err := StartWorkerPoolWithCommand(2, ".", mockCmd, nil)
		if err != nil {
			t.Fatalf("Failed to create worker pool: %v", err)
		}
		defer wp.Shutdown()

		if !waitForWorkerStart(t, wp, startTimeout) {
			t.Fatal("Workers failed to start within timeout")
		}

		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				endpoint := wp.GetWorkerEndpoint()
				if endpoint == "" {
					errChan <- fmt.Errorf("got empty endpoint in concurrent access")
				}
			}()
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			t.Error(err)
		}
	})

	t.Run("worker process lifecycle", func(t *testing.T) {
		wp, err := StartWorkerPoolWithCommand(1, ".", mockCmd, nil)
		if err != nil {
			t.Fatalf("Failed to create worker pool: %v", err)
		}

		if !waitForWorkerStart(t, wp, startTimeout) {
			t.Fatal("Worker failed to start within timeout")
		}

		endpoint := wp.GetWorkerEndpoint()
		if endpoint == "" {
			t.Error("Expected non-empty endpoint")
		}

		wp.Shutdown()

		if !waitForWorkerShutdown(t, wp, shutdownTimeout) {
			t.Fatal("Worker failed to shut down within timeout")
		}

		endpoint = wp.GetWorkerEndpoint()
		if endpoint != "" {
			t.Error("Expected no endpoint after shutdown")
		}
	})

	t.Run("worker restart", func(t *testing.T) {
		// Start first worker pool
		wp1, err := StartWorkerPoolWithCommand(1, ".", mockCmd, nil)
		if err != nil {
			t.Fatalf("Failed to create first worker pool: %v", err)
		}

		if !waitForWorkerStart(t, wp1, startTimeout) {
			t.Fatal("First worker failed to start within timeout")
		}

		endpoint1 := wp1.GetWorkerEndpoint()
		if endpoint1 == "" {
			t.Error("Expected non-empty initial endpoint")
		}

		wp1.Shutdown()

		if !waitForWorkerShutdown(t, wp1, shutdownTimeout) {
			t.Fatal("First worker failed to shut down within timeout")
		}

		// Start second worker pool
		wp2, err := StartWorkerPoolWithCommand(1, ".", mockCmd, nil)
		if err != nil {
			t.Fatalf("Failed to create second worker pool: %v", err)
		}
		defer wp2.Shutdown()

		if !waitForWorkerStart(t, wp2, startTimeout) {
			t.Fatal("Second worker failed to start within timeout")
		}

		endpoint2 := wp2.GetWorkerEndpoint()
		if endpoint2 == "" {
			t.Error("Expected non-empty endpoint after restart")
		}

		if endpoint1 == endpoint2 {
			t.Error("Expected different endpoints after restart")
		}
	})
}
