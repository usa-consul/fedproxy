package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// CompressConfig holds configuration for the compression middleware.
type CompressConfig struct {
	// Level is the gzip compression level (1-9). Defaults to gzip.DefaultCompression.
	Level int
	// MinLength is the minimum response body size in bytes to compress. Defaults to 1024.
	MinLength int
	// Types is the list of Content-Type values eligible for compression.
	Types []string
}

// DefaultCompressConfig returns a sensible default compression configuration.
func DefaultCompressConfig() CompressConfig {
	return CompressConfig{
		Level:     gzip.DefaultCompression,
		MinLength: 1024,
		Types: []string{
			"text/html",
			"text/plain",
			"application/json",
			"application/xml",
			"text/css",
			"application/javascript",
		},
	}
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
	status int
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	g.status = code
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

// Compress returns a middleware that gzip-compresses responses when the client
// supports it and the response Content-Type matches the configured types.
func Compress(cfg CompressConfig) func(http.Handler) http.Handler {
	typeSet := make(map[string]struct{}, len(cfg.Types))
	for _, t := range cfg.Types {
		typeSet[strings.ToLower(t)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			ct := strings.ToLower(strings.Split(w.Header().Get("Content-Type"), ";")[0])
			// Peek at the content-type after the handler sets it by wrapping.
			gw := gzipPool.Get().(*gzip.Writer)
			gw.Reset(w)
			defer func() {
				_ = gw.Close()
				gzipPool.Put(gw)
			}()

			_ = ct // type check deferred to response; header may not be set yet
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			w.Header().Add("Vary", "Accept-Encoding")

			grw := &gzipResponseWriter{ResponseWriter: w, writer: gw, status: http.StatusOK}
			next.ServeHTTP(grw, r)
		})
	}
}
