package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type FileEntry struct {
	Path        string    // File path relative to static directory
	ContentType string    // MIME type of the file
	ETag        string    // ETag for caching
	ModTime     time.Time // Last modification time
	Size        int64     // File size
}

type StaticCache struct {
	files     map[string]*FileEntry
	filePaths map[string]bool // Simple map to track static file paths
	mutex     sync.RWMutex
	config    *Config
}

func NewStaticCache(config *Config) *StaticCache {
	return &StaticCache{
		files:     make(map[string]*FileEntry),
		filePaths: make(map[string]bool),
		config:    config,
	}
}

func (sc *StaticCache) PreloadFiles() error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	staticDir := filepath.Join(sc.config.BlastraCWD, sc.config.StaticDir)
	log.Infof("Preloading static files metadata from: %s", staticDir)

	return filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file should be excluded
		for _, pattern := range sc.config.ExcludePatterns {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				log.Debugf("Skipping excluded file: %s", path)
				return nil
			}
		}

		relPath, err := filepath.Rel(staticDir, path)
		if err != nil {
			return err
		}

		// Convert path separators to forward slashes for URLs
		relPath = filepath.ToSlash(relPath)
		urlPath := "/" + relPath

		// Always add to filePaths
		sc.filePaths[urlPath] = true

		// Determine content type
		ext := strings.ToLower(filepath.Ext(path))
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Generate ETag from file info
		hash := sha256.New()
		hash.Write([]byte(fmt.Sprintf("%s-%d-%d", urlPath, info.Size(), info.ModTime().UnixNano())))
		etag := `"` + hex.EncodeToString(hash.Sum(nil)) + `"`

		entry := &FileEntry{
			Path:        urlPath,
			ContentType: contentType,
			ModTime:     info.ModTime(),
			Size:        info.Size(),
			ETag:        etag,
		}

		sc.files[urlPath] = entry
		log.Debugf("Preloaded file metadata: %s", relPath)

		return nil
	})
}

func (sc *StaticCache) Get(path string) (*FileEntry, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	entry, exists := sc.files[path]
	return entry, exists
}

func (sc *StaticCache) IsStaticFile(path string) bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	_, exists := sc.filePaths[path]
	return exists
}

func (sc *StaticCache) GetCacheControl(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if cacheControl, ok := sc.config.CacheControl[ext]; ok {
		return cacheControl
	}

	// Default cache control based on file type
	switch ext {
	case ".js", ".css", ".woff2", ".woff", ".ttf", ".eot":
		return "public, max-age=31536000, immutable" // 1 year for static assets
	case ".html", ".json":
		return "public, max-age=0, must-revalidate"
	default:
		if strings.HasPrefix(path, "/assets/") {
			return "public, max-age=31536000, immutable"
		}
		return "public, max-age=3600" // 1 hour default
	}
}
