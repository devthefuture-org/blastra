package server

import (
	"io"
	"net/http"
	"time"

	"github.com/devthefuture-org/blastra/pkg/cache"
	"github.com/devthefuture-org/blastra/pkg/worker"
	log "github.com/sirupsen/logrus"
)

func handleWorkerSSR(w http.ResponseWriter, r *http.Request, wp worker.IWorkerPool, ssrCache *cache.CacheProvider, notFoundCache *cache.CacheProvider, cacheKey string) bool {
	// Handle nil worker pool
	if wp == nil {
		return false
	}

	endpoint := wp.GetWorkerEndpoint()
	if endpoint == "" {
		return false
	}

	log.Debugf("Attempting SSR via worker pool for: %s", r.URL.Path)
	ssrURL := endpoint + r.URL.Path

	req, err := http.NewRequest("GET", ssrURL, nil)
	if err != nil {
		log.Errorf("Failed to create worker request: %v", err)
		return false
	}

	// Copy relevant headers from original request
	for _, header := range []string{"Accept", "Accept-Language", "Cookie", "User-Agent"} {
		if value := r.Header.Get(header); value != "" {
			req.Header.Set(header, value)
		}
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Worker request failed: %v", err)
		return false // Fall back to direct SSR
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read worker response: %v", err)
		return false // Fall back to direct SSR
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Ensure content type is set
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	// Cache responses only if caching is enabled and response is cacheable
	if resp.StatusCode == http.StatusNotFound && notFoundCache != nil {
		notFoundCache.Set(cacheKey, body)
	} else if resp.StatusCode == http.StatusOK && ssrCache != nil {
		ssrCache.Set(cacheKey, body)
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
	return true
}
