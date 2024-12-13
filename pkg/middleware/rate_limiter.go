package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// ClientLimiter holds the rate limiter for a specific client
type ClientLimiter struct {
	limiter  *rate.Limiter
	lastSeen int64
}

// IPRateLimiter manages per-client rate limiters
type IPRateLimiter struct {
	mu         sync.RWMutex
	limiters   map[string]*ClientLimiter
	limit      rate.Limit
	burst      int
	trustProxy bool
}

// NewIPRateLimiter creates a new rate limiter that tracks clients by IP address
func NewIPRateLimiter(limit rate.Limit, burst int, trustProxy bool) *IPRateLimiter {
	return &IPRateLimiter{
		limiters:   make(map[string]*ClientLimiter),
		limit:      limit,
		burst:      burst,
		trustProxy: trustProxy,
	}
}

// GetLimiter returns the rate limiter for the provided IP address
func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = &ClientLimiter{
			limiter: rate.NewLimiter(rl.limit, rl.burst),
		}
		rl.limiters[ip] = limiter
	}

	return limiter.limiter
}

// getIP extracts the real IP address from the request, considering proxy headers if trusted
func (rl *IPRateLimiter) getIP(r *http.Request) string {
	// If proxy headers are trusted, check X-Forwarded-For first
	if rl.trustProxy {
		// X-Forwarded-For contains a list of IPs in the format: client, proxy1, proxy2, ...
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// Get the first (client) IP from the list
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				clientIP := strings.TrimSpace(ips[0])
				if net.ParseIP(clientIP) != nil {
					return clientIP
				}
			}
		}

		// Fallback to X-Real-IP if X-Forwarded-For is not valid
		if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
			if net.ParseIP(xrip) != nil {
				return xrip
			}
		}
	}

	// If no proxy headers or they're not trusted, get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, try using RemoteAddr directly
		ip = r.RemoteAddr
	}
	return ip
}

// RateLimitMiddleware creates a new middleware handler for rate limiting
func RateLimitMiddleware(limiter *rate.Limiter, trustProxy bool) func(http.Handler) http.Handler {
	if limiter == nil {
		log.Debug("Rate limiting disabled (nil limiter)")
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	ipLimiter := NewIPRateLimiter(limiter.Limit(), limiter.Burst(), trustProxy)
	log.Debugf("Rate limiting enabled with limit: %v, burst: %d, trust proxy: %v",
		limiter.Limit(), limiter.Burst(), trustProxy)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ipLimiter.getIP(r)
			limiter := ipLimiter.GetLimiter(ip)

			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
