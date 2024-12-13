package main

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/devthefuture-org/blastra/pkg/cache"
	"github.com/devthefuture-org/blastra/pkg/config"
	"github.com/devthefuture-org/blastra/pkg/health"
	"github.com/devthefuture-org/blastra/pkg/logging"
	"github.com/devthefuture-org/blastra/pkg/server"
	"github.com/devthefuture-org/blastra/pkg/shutdown"
	"github.com/devthefuture-org/blastra/pkg/worker"
)

func waitForServer(port int, maxRetries int) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	for i := 0; i < maxRetries; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func main() {
	// Configure logging
	logging.ConfigureLogging()

	// Load configuration
	cfg, err := config.LoadConfiguration()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Set GOMAXPROCS according to CPUCount
	runtime.GOMAXPROCS(cfg.CPUCount)
	log.Debugf("GOMAXPROCS set to %d", cfg.CPUCount)

	// Initialize components
	var ssrCacheProvider *cache.CacheProvider
	var notFoundCacheProvider *cache.CacheProvider

	if cfg.SSRCacheEnabled {
		// Initialize SSR cache if enabled
		cacheTTL, cacheSize := cfg.GetSSRCacheConfig()
		ssrMemoryCache := cache.NewSSRInMemoryCache(cache.CacheConfig{
			TTL:     cacheTTL,
			MaxSize: cacheSize,
		})

		// Initialize NotFoundCache with configuration from config package
		notFoundTTL, notFoundSize := cfg.GetNotFoundCacheConfig()
		notFoundMemoryCache := cache.NewNotFoundInMemoryCache(cache.CacheConfig{
			TTL:     notFoundTTL,
			MaxSize: notFoundSize,
		})

		// Create cache providers with external caches if configured
		externalConfig := cfg.GetExternalCacheConfig()

		var err error
		ssrCacheProvider, err = cache.CreateCacheProvider(ssrMemoryCache, externalConfig)
		if err != nil {
			log.Fatalf("Failed to create SSR cache provider: %v", err)
		}

		notFoundCacheProvider, err = cache.CreateCacheProvider(notFoundMemoryCache, externalConfig)
		if err != nil {
			log.Fatalf("Failed to create NotFound cache provider: %v", err)
		}

		log.Infof("SSR caching enabled with external cache type: %s", externalConfig.Type)
	} else {
		log.Info("SSR caching is disabled")
	}

	healthChecker := health.NewHealthChecker()

	// Initialize worker pool with configuration
	var wp worker.IWorkerPool
	wp, err = worker.StartWorkerPoolWithConfig(cfg.WorkerCount, cfg.BlastraCWD, cfg.WorkerCommand, cfg.WorkerArgs, cfg.WorkerURLs)
	if err != nil {
		log.Warnf("Failed to start worker pool: %v, fallback to direct SSR command mode", err)
		wp, _ = worker.StartWorkerPoolWithConfig(0, cfg.BlastraCWD, cfg.WorkerCommand, cfg.WorkerArgs, nil) // Create disabled worker pool
	}

	// Initialize server
	ssrHandler := server.SSRHandler(ssrCacheProvider, notFoundCacheProvider, cfg.SSRScript, cfg.MaxAgeSSR, cfg.BlastraCWD, wp)
	serverConfig := &server.Config{
		BlastraCWD:    cfg.BlastraCWD,
		StaticDir:     cfg.StaticDir,
		SSRHandler:    ssrHandler,
		HealthChecker: healthChecker,
	}

	serverInitConfig := &server.ServerInitConfig{
		HTTPPort:     cfg.HTTPPort,
		EnableHTTPS:  cfg.EnableHTTPS,
		HTTPSPort:    cfg.HTTPSPort,
		TLSCertPath:  cfg.TLSCertPath,
		TLSKeyPath:   cfg.TLSKeyPath,
		RateLimit:    cfg.RateLimit,
		Burst:        cfg.Burst,
		GzipEnabled:  cfg.GzipEnabled,
		TrustProxy:   cfg.TrustProxy,
		ServerConfig: serverConfig,
	}

	srv := server.InitializeServer(serverInitConfig)

	// Setup shutdown handling
	shutdownConfig := &shutdown.ShutdownConfig{
		Server:          srv,
		WorkerPool:      wp,
		ShutdownTimeout: cfg.ShutdownTimeout,
	}
	serverErrors := shutdown.HandleGracefulShutdown(shutdownConfig)

	// Start HTTP server
	go func() {
		log.Infof("Starting HTTP server on port %d", cfg.HTTPPort)
		serverErrors <- srv.ListenAndServe()
	}()

	// Start HTTPS server if enabled
	if cfg.EnableHTTPS {
		go func() {
			log.Infof("Starting HTTPS server on port %d", cfg.HTTPSPort)
			serverErrors <- srv.ListenAndServeTLS(cfg.TLSCertPath, cfg.TLSKeyPath)
		}()
	}

	// Check server availability and set ready status
	go func() {
		waitForServer(cfg.HTTPPort, 10) // Try up to 10 times

		logging.LogAsciiArt()

		healthChecker.SetReady()
	}()

	// Wait for server errors
	err = <-serverErrors
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Info("Server stopped gracefully")
}
