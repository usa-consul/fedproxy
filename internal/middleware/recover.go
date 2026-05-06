package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// RecoverConfig holds configuration for the panic recovery middleware.
type RecoverConfig struct {
	// Logger is used to emit panic details. If nil, slog.Default() is used.
	Logger *slog.Logger
	// PrintStack controls whether the goroutine stack trace is logged.
	PrintStack bool
}

// DefaultRecoverConfig returns a RecoverConfig with sensible defaults.
func DefaultRecoverConfig() RecoverConfig {
	return RecoverConfig{
		Logger:     slog.Default(),
		PrintStack: true,
	}
}

// Recover returns middleware that catches panics, logs them, and responds with
// 500 Internal Server Error so the proxy process stays alive.
func Recover(cfg RecoverConfig) func(http.Handler) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					attrs := []any{
						slog.Any("panic", rec),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					}
					if cfg.PrintStack {
						attrs = append(attrs, slog.String("stack", string(debug.Stack())))
					}
					cfg.Logger.Error("recovered from panic", attrs...)

					// Only write the header if nothing has been sent yet.
					if rw, ok := w.(*ResponseRecorder); ok && !rw.Written() {
						http.Error(w, "internal server error", http.StatusInternalServerError)
						return
					}
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
