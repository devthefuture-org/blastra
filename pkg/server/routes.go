package server

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"mime"

	log "github.com/sirupsen/logrus"
	ratelimit "go.uber.org/ratelimit"
	"golang.org/x/time/rate"
)

// IPRateLimiter manages per-IP rate limiters
type IPRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]ratelimit.Limiter
	rps      int
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(rps int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]ratelimit.Limiter),
		rps:      rps,
	}
}

// GetLimiter returns the rate limiter for the provided IP
func (rl *IPRateLimiter) GetLimiter(ip string) ratelimit.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = ratelimit.New(rl.rps)
		rl.limiters[ip] = limiter
		log.Debugf("Created new rate limiter for IP: %s", ip)
	}

	return limiter
}

func getClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		// Check X-Forwarded-For header
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				return strings.TrimSpace(ips[0])
			}
		}
		// Fallback to X-Real-IP
		if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
			return strings.TrimSpace(xrip)
		}
	}
	// Default to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RouteConfig contains configuration for route setup
type RouteConfig struct {
	*Config
	TrustProxy bool
}

func SetupRoutes(mux *http.ServeMux, config *RouteConfig, rateLimiter *rate.Limiter) {
	staticDir := filepath.Join(config.BlastraCWD, config.StaticDir)
	log.Debugf("Setting up routes with static directory: %s", staticDir)

	fileServer := http.StripPrefix("/", CreateFileServer(config.Config))

	// Create IP-based rate limiter if enabled
	var ipLimiter *IPRateLimiter
	if rateLimiter != nil {
		// Convert from rate.Limit (requests per second) to integer RPS
		rps := int(rateLimiter.Limit())
		if rps < 1 {
			rps = 1
		}
		ipLimiter = NewIPRateLimiter(rps)
		log.Debugf("Rate limiting enabled with limit: %v requests/second per IP", rps)
	}

	// Main handler for all routes except health checks
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Received request for: %s", r.URL.Path)

		// Check if the path exists in static files
		isStaticFile := false
		if handler, ok := fileServer.(*staticFileHandler); ok && handler.staticCache != nil {
			// If static cache is available, use the file list
			isStaticFile = handler.staticCache.IsStaticFile(r.URL.Path)
			log.Debugf("Checking static files list for: %s, found: %v", r.URL.Path, isStaticFile)
		} else {
			// If static cache is not available, check if file exists
			filePath := filepath.Join(staticDir, r.URL.Path)
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				isStaticFile = true
				log.Debugf("Found static file at: %s", filePath)
			}
		}

		if isStaticFile {
			// Set common headers before serving static file
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Connection", "keep-alive")

			// Ensure proper content type for the file
			ext := strings.ToLower(filepath.Ext(r.URL.Path))
			if mimeType := mime.TypeByExtension(ext); mimeType != "" {
				w.Header().Set("Content-Type", mimeType)
				log.Debugf("Set Content-Type: %s for file: %s", mimeType, r.URL.Path)
			}

			// Serve the static file directly without rate limiting
			fileServer.ServeHTTP(w, r)
			return
		}

		// If not a static file, it's an SSR request
		log.Debugf("No static file found for: %s, serving SSR", r.URL.Path)

		// Apply rate limiting for SSR if enabled
		if ipLimiter != nil {
			clientIP := getClientIP(r, config.TrustProxy)
			limiter := ipLimiter.GetLimiter(clientIP)

			// Take a token from the bucket
			limiter.Take()
			log.Debugf("Rate limit token taken for IP: %s", clientIP)
		}

		// Ensure SSR handler exists
		if config.SSRHandler == nil {
			log.Error("SSR handler is nil")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Serve SSR
		config.SSRHandler(w, r)
	})

	// Health check endpoints (no rate limiting)
	mux.HandleFunc("/live", config.HealthChecker.LivenessProbeHandler)
	mux.HandleFunc("/ready", config.HealthChecker.ReadinessProbeHandler)
}
