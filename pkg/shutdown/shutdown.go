package shutdown

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devthefuture-org/blastra/pkg/worker"
	log "github.com/sirupsen/logrus"
)

// Server interface defines the methods we need from http.Server
type Server interface {
	Shutdown(ctx context.Context) error
	Close() error
	ListenAndServe() error
}

type ShutdownConfig struct {
	Server          Server
	WorkerPool      worker.IWorkerPool
	ShutdownTimeout time.Duration
	TestShutdown    chan struct{} // Used for testing only
}

func HandleGracefulShutdown(cfg *ShutdownConfig) chan error {
	serverErrors := make(chan error, 1)

	go func() {
		var quit chan os.Signal
		if cfg.TestShutdown == nil {
			// Production mode: use OS signals
			quit = make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		}

		// Wait for either OS signal or test shutdown
		if cfg.TestShutdown != nil {
			<-cfg.TestShutdown
		} else {
			<-quit
		}

		log.Info("Shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		// Shutdown server
		if err := cfg.Server.Shutdown(ctx); err != nil {
			log.Errorf("Graceful shutdown failed: %v", err)
			if err := cfg.Server.Close(); err != nil {
				log.Errorf("Server close failed: %v", err)
			}
		}

		// Shutdown worker pool
		if cfg.WorkerPool != nil {
			cfg.WorkerPool.Shutdown()
		}

		// Send ErrServerClosed to indicate normal shutdown
		serverErrors <- http.ErrServerClosed
		close(serverErrors)
	}()

	return serverErrors
}
