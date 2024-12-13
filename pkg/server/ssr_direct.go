package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/devthefuture-org/blastra/pkg/cache"
	"github.com/devthefuture-org/blastra/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func handleDirectSSR(w http.ResponseWriter, r *http.Request, ssrCommand []string, cwd string, ssrCache *cache.CacheProvider, notFoundCache *cache.CacheProvider, cacheKey string, maxAge int) {
	log.Debugf("Worker pool not active, executing SSR command directly for: %s", r.URL.Path)

	fullSsrCommand := append(ssrCommand, r.URL.Path)
	cmd := exec.Command(fullSsrCommand[0], fullSsrCommand[1:]...)
	cmd.Dir = cwd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		log.Errorf("Error starting SSR command: %v, stderr: %s", err, stderr.String())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Errorf("Error waiting for SSR command: %v, stderr: %s", err, stderr.String())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var ssrResponse utils.SSRResponse
	if err := json.Unmarshal(stdout.Bytes(), &ssrResponse); err != nil {
		log.Errorf("Failed to parse SSR JSON response: %v, stdout: %s", err, stdout.String())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if ssrResponse.Error != "" {
		log.Errorf("SSR returned error: %s, code: %d", ssrResponse.Error, ssrResponse.Code)
		statusCode := ssrResponse.Code
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}
		http.Error(w, ssrResponse.Error, statusCode)
		return
	}

	renderedContent := []byte(ssrResponse.HTML)
	currentTime := time.Now()

	// Handle 404 responses
	if ssrResponse.Code == http.StatusNotFound {
		// Cache 404 response only if caching is enabled
		if notFoundCache != nil {
			notFoundCache.Set(cacheKey, renderedContent)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
		w.Header().Set("Last-Modified", currentTime.UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusNotFound)
		w.Write(renderedContent)
		return
	}

	// Cache successful response only if caching is enabled
	if ssrCache != nil {
		ssrCache.Set(cacheKey, renderedContent)
	}

	ifModifiedSince := r.Header.Get("If-Modified-Since")
	if ifModifiedSince != "" {
		t, err := time.Parse(http.TimeFormat, ifModifiedSince)
		if err == nil && !currentTime.After(t) {
			log.Debugf("Returning 304 Not Modified for: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
	w.Header().Set("Last-Modified", currentTime.UTC().Format(http.TimeFormat))
	w.Write(renderedContent)
}
