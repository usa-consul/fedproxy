package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed.
	// Use ["*"] to allow all origins.
	AllowedOrigins []string

	// AllowedMethods specifies the HTTP methods allowed for CORS requests.
	AllowedMethods []string

	// AllowedHeaders specifies the request headers allowed for CORS requests.
	AllowedHeaders []string

	// AllowCredentials indicates whether the request can include credentials.
	AllowCredentials bool

	// MaxAge sets the Access-Control-Max-Age header in seconds.
	MaxAge string
}

// DefaultCORSConfig returns a conservative CORS configuration suitable
// for federal proxy deployments.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           "600",
	}
}

// CORS returns a middleware that handles Cross-Origin Resource Sharing headers.
// Preflight OPTIONS requests are responded to immediately with 204.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	allowedSet := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		allowedSet[o] = struct{}{}
	}

	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			_, wildcard := allowedSet["*"]
			_, exact := allowedSet[origin]

			if !wildcard && !exact {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			allowOrigin := origin
			if wildcard {
				allowOrigin = "*"
			}

			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if cfg.MaxAge != "" {
				w.Header().Set("Access-Control-Max-Age", cfg.MaxAge)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
