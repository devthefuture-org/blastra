package shutdown

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// mockServer implements the necessary methods of http.Server
type mockServer struct {
	shutdownCalled bool
	closeCalled    bool
}

func (s *mockServer) Shutdown(ctx context.Context) error {
	s.shutdownCalled = true
	return nil
}

func (s *mockServer) Close() error {
	s.closeCalled = true
	return nil
}

func (s *mockServer) ListenAndServe() error {
	return http.ErrServerClosed
}

type mockWorkerPool struct {
	shutdownCalled bool
}

func (w *mockWorkerPool) GetWorkerEndpoint() string {
	return ""
}

func (w *mockWorkerPool) Shutdown() {
	w.shutdownCalled = true
}

func TestHandleGracefulShutdown(t *testing.T) {
	t.Run("normal shutdown", func(t *testing.T) {
		server := &mockServer{}
		workerPool := &mockWorkerPool{}
		testShutdown := make(chan struct{})

		config := &ShutdownConfig{
			Server:          server,
			WorkerPool:      workerPool,
			ShutdownTimeout: 5 * time.Second,
			TestShutdown:    testShutdown,
		}

		errors := HandleGracefulShutdown(config)

		// Trigger shutdown and wait for completion
		close(testShutdown)
		err := <-errors

		// Give a small amount of time for shutdown operations to complete
		time.Sleep(100 * time.Millisecond)

		// Verify shutdown was called
		if !server.shutdownCalled {
			t.Error("Expected server.Shutdown to be called")
		}
		if server.closeCalled {
			t.Error("Expected server.Close not to be called in normal shutdown")
		}
		if !workerPool.shutdownCalled {
			t.Error("Expected worker pool shutdown to be called")
		}
		if err != http.ErrServerClosed {
			t.Errorf("Expected error %v, got %v", http.ErrServerClosed, err)
		}
	})

	t.Run("shutdown timeout", func(t *testing.T) {
		server := &mockServer{}
		workerPool := &mockWorkerPool{}
		testShutdown := make(chan struct{})

		config := &ShutdownConfig{
			Server:          server,
			WorkerPool:      workerPool,
			ShutdownTimeout: 1 * time.Millisecond, // Very short timeout to trigger timeout case
			TestShutdown:    testShutdown,
		}

		errors := HandleGracefulShutdown(config)

		// Trigger shutdown and wait for completion
		close(testShutdown)
		<-errors

		// Give a small amount of time for shutdown operations to complete
		time.Sleep(100 * time.Millisecond)

		// Verify both shutdown and close were called due to timeout
		if !server.shutdownCalled {
			t.Error("Expected server.Shutdown to be called")
		}
		if !workerPool.shutdownCalled {
			t.Error("Expected worker pool shutdown to be called")
		}
	})

	t.Run("nil worker pool", func(t *testing.T) {
		server := &mockServer{}
		testShutdown := make(chan struct{})

		config := &ShutdownConfig{
			Server:          server,
			WorkerPool:      nil,
			ShutdownTimeout: 5 * time.Second,
			TestShutdown:    testShutdown,
		}

		errors := HandleGracefulShutdown(config)

		// Trigger shutdown and wait for completion
		close(testShutdown)
		err := <-errors

		// Give a small amount of time for shutdown operations to complete
		time.Sleep(100 * time.Millisecond)

		// Verify shutdown was called
		if !server.shutdownCalled {
			t.Error("Expected server.Shutdown to be called")
		}
		if err != http.ErrServerClosed {
			t.Errorf("Expected error %v, got %v", http.ErrServerClosed, err)
		}
	})
}
