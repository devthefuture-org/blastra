package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return w
	},
}

type GzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipMiddleware(enabled bool) func(http.Handler) http.Handler {
	if !enabled {
		log.Debug("Gzip disabled")
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	log.Debug("Gzip enabled")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Skip compression if client doesn't accept gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gz := gzipWriterPool.Get().(*gzip.Writer)
			gz.Reset(w)
			defer func() {
				gz.Close()
				gzipWriterPool.Put(gz)
			}()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			w.Header().Add("Vary", "Accept-Encoding")

			gzw := &GzipResponseWriter{
				Writer:         gz,
				ResponseWriter: w,
			}
			next.ServeHTTP(gzw, r)
		})
	}
}
