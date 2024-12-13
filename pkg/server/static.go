package server

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type staticFileHandler struct {
	root        string
	staticCache *StaticCache
	config      *Config
}

func (h *staticFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Static file request received: %s", r.URL.Path)

	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(r.URL.Path)
	filePath := filepath.Join(h.root, cleanPath)

	// Get file info before opening
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If it's a directory, return 404
	if stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	// Open file
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set content type
	ext := filepath.Ext(cleanPath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set all headers before serving content
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Keep-Alive", "timeout=5, max=1000")
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	// Set cache control
	if h.staticCache != nil {
		w.Header().Set("Cache-Control", h.staticCache.GetCacheControl(cleanPath))
	} else if h.config.StaticMaxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", h.config.StaticMaxAge))
	}

	// Use http.ServeContent for efficient serving with range support
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func CreateFileServer(config *Config) http.Handler {
	log.Debugf("Creating file server for directory: %s", config.StaticDir)

	handler := &staticFileHandler{
		root:   filepath.Join(config.BlastraCWD, config.StaticDir),
		config: config,
	}

	if config.ShouldPreloadFileList() {
		staticCache := NewStaticCache(config)
		if err := staticCache.PreloadFiles(); err != nil {
			log.Errorf("Failed to preload static files: %v", err)
		} else {
			handler.staticCache = staticCache
		}
	}

	return handler
}
