package server

import (
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/devthefuture-org/blastra/pkg/middleware"
)

type ServerInitConfig struct {
	HTTPPort     int
	EnableHTTPS  bool
	HTTPSPort    int
	TLSCertPath  string
	TLSKeyPath   string
	RateLimit    rate.Limit
	Burst        int
	GzipEnabled  bool
	TrustProxy   bool
	ServerConfig *Config
}

func InitializeServer(cfg *ServerInitConfig) *http.Server {
	mux := http.NewServeMux()

	// Create rate limiter if enabled
	var limiter *rate.Limiter
	if cfg.RateLimit > 0 && cfg.Burst > 0 {
		limiter = rate.NewLimiter(cfg.RateLimit, cfg.Burst)
		log.Debugf("Rate limiting enabled with limit: %v, burst: %d", cfg.RateLimit, cfg.Burst)
	} else {
		log.Debug("Rate limiting disabled")
	}

	// Ensure ServerConfig is not nil
	if cfg.ServerConfig == nil {
		log.Fatal("ServerConfig cannot be nil")
	}

	// Create RouteConfig with TrustProxy setting
	routeConfig := &RouteConfig{
		Config:     cfg.ServerConfig,
		TrustProxy: cfg.TrustProxy,
	}

	// Setup routes with rate limiter (which may be nil if disabled)
	SetupRoutes(mux, routeConfig, limiter)

	// Add gzip middleware
	var handler http.Handler = mux
	handler = middleware.GzipMiddleware(cfg.GzipEnabled)(handler)

	// Create server with timeouts
	server := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.HTTPPort),
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Minute, // Long timeout for large files
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	return server
}
