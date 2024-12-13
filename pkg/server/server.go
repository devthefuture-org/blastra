package server

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/devthefuture-org/blastra/pkg/cache"
	"github.com/devthefuture-org/blastra/pkg/health"
	"github.com/devthefuture-org/blastra/pkg/worker"
)

type Config struct {
	BlastraCWD            string
	StaticDir             string
	SSRHandler            http.HandlerFunc
	HealthChecker         *health.HealthChecker
	PreloadStaticFileList *bool             // Whether to preload static file list for routing (default: true)
	PreloadStaticContent  *bool             // Whether to preload static files into memory (default: true)
	StaticMaxAge          int               // Cache duration for static files in seconds
	ExcludePatterns       []string          // Patterns to exclude from preloading
	CacheControl          map[string]string // Custom cache control headers for different file types
}

// Helper function to get PreloadStaticFileList with default value
func (c *Config) ShouldPreloadFileList() bool {
	if c.PreloadStaticFileList == nil {
		return true // default enabled
	}
	return *c.PreloadStaticFileList
}

// Helper function to get PreloadStaticContent with default value
func (c *Config) ShouldPreloadContent() bool {
	if c.PreloadStaticContent == nil {
		return true // default enabled
	}
	return *c.PreloadStaticContent
}

// Helper function to generate ETag for file content
func generateETag(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])[:32] // Use first 32 chars of hash
}

// Helper function to preload static content
func preloadStaticContent(staticDir string, excludePatterns []string) (map[string]*cache.CacheEntry, error) {
	contentCache := make(map[string]*cache.CacheEntry)
	err := filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file matches exclude patterns
		relPath, err := filepath.Rel(staticDir, path)
		if err != nil {
			return err
		}
		for _, pattern := range excludePatterns {
			if matched, _ := filepath.Match(pattern, relPath); matched {
				return nil
			}
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Generate ETag
		etag := generateETag(content)

		// Create cache entry
		contentCache["/"+relPath] = &cache.CacheEntry{
			Content:     content,
			ETag:        etag,
			LastUpdated: info.ModTime(),
		}

		return nil
	})

	return contentCache, err
}

func SSRHandler(ssrCache *cache.CacheProvider, notFoundCache *cache.CacheProvider, ssrCommand []string, maxAge int, cwd string, wp worker.IWorkerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Received SSR request: %s", r.URL.Path)
		cacheKey := r.URL.Path

		// Try to serve from cache first if caching is enabled
		if ssrCache != nil {
			if entry, found := ssrCache.Get(cacheKey); found {
				// Add ETag support
				w.Header().Set("ETag", entry.ETag)

				// Check If-None-Match
				if match := r.Header.Get("If-None-Match"); match != "" && match == entry.ETag {
					w.WriteHeader(http.StatusNotModified)
					return
				}

				// Check If-Modified-Since
				ifModifiedSince := r.Header.Get("If-Modified-Since")
				if ifModifiedSince != "" {
					t, err := time.Parse(http.TimeFormat, ifModifiedSince)
					if err == nil && !entry.LastUpdated.After(t) {
						log.Debugf("Returning 304 Not Modified for: %s", r.URL.Path)
						w.WriteHeader(http.StatusNotModified)
						return
					}
				}

				log.Debugf("Serving cached SSR response for: %s", r.URL.Path)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
				w.Header().Set("Last-Modified", entry.LastUpdated.UTC().Format(http.TimeFormat))
				w.Write(entry.Content)
				return
			}
		}

		// Check NotFoundCache for 404 responses if caching is enabled
		if notFoundCache != nil {
			if entry, found := notFoundCache.Get(cacheKey); found {
				// Check If-None-Match for 404 response
				if match := r.Header.Get("If-None-Match"); match != "" && match == entry.ETag {
					w.WriteHeader(http.StatusNotModified)
					return
				}

				log.Debugf("Serving cached 404 response for: %s", r.URL.Path)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
				w.Header().Set("ETag", entry.ETag)
				w.Header().Set("Last-Modified", entry.LastUpdated.UTC().Format(http.TimeFormat))
				w.WriteHeader(http.StatusNotFound)
				w.Write(entry.Content)
				return
			}
		}

		// Try worker-based SSR first
		if handled := handleWorkerSSR(w, r, wp, ssrCache, notFoundCache, cacheKey); handled {
			return
		}

		// Fall back to direct command execution
		handleDirectSSR(w, r, ssrCommand, cwd, ssrCache, notFoundCache, cacheKey, maxAge)
	}
}

func StaticHandler(staticDir string, maxAge int, preloadContent bool, excludePatterns []string) http.HandlerFunc {
	var contentCache map[string]*cache.CacheEntry
	if preloadContent {
		var err error
		contentCache, err = preloadStaticContent(staticDir, excludePatterns)
		if err != nil {
			log.Errorf("Failed to preload static content: %v", err)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticDir, r.URL.Path)

		// Try to serve from preloaded cache first
		if preloadContent && contentCache != nil {
			if entry, found := contentCache[r.URL.Path]; found {
				// Add ETag support
				w.Header().Set("ETag", entry.ETag)

				// Check If-None-Match
				if match := r.Header.Get("If-None-Match"); match != "" && match == entry.ETag {
					w.WriteHeader(http.StatusNotModified)
					return
				}

				// Check If-Modified-Since
				ifModifiedSince := r.Header.Get("If-Modified-Since")
				if ifModifiedSince != "" {
					t, err := time.Parse(http.TimeFormat, ifModifiedSince)
					if err == nil && !entry.LastUpdated.After(t) {
						w.WriteHeader(http.StatusNotModified)
						return
					}
				}

				// Set headers
				w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
				w.Header().Set("Last-Modified", entry.LastUpdated.UTC().Format(http.TimeFormat))
				w.Write(entry.Content)
				return
			}
		}

		// Serve file directly if not in cache
		file, err := os.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Read file content for ETag generation
		content, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Generate ETag
		etag := generateETag(content)
		w.Header().Set("ETag", etag)

		// Check If-None-Match
		if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Check If-Modified-Since
		if ifModifiedSince := r.Header.Get("If-Modified-Since"); ifModifiedSince != "" {
			if t, err := time.Parse(http.TimeFormat, ifModifiedSince); err == nil {
				if !info.ModTime().After(t) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}

		// Set headers
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
		w.Header().Set("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))
		w.Write(content)
	}
}
