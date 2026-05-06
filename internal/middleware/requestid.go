package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// RequestIDHeader is the HTTP header used to propagate request IDs.
const RequestIDHeader = "X-Request-ID"

// RequestID is a middleware that assigns a unique ID to every incoming request.
// If the upstream client already supplies an X-Request-ID header its value is
// reused; otherwise a new random 16-byte hex ID is generated.
// The ID is stored on the request context and echoed back in the response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = generateID()
		}

		// Propagate to upstream via request header.
		r = r.WithContext(context.WithValue(r.Context(), RequestIDKey, id))
		r.Header.Set(RequestIDHeader, id)

		// Echo back to the caller.
		w.Header().Set(RequestIDHeader, id)

		next.ServeHTTP(w, r)
	})
}

// FromContext retrieves the request ID stored on the context, or an empty
// string if none is present.
func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback: return a fixed sentinel so callers can detect the failure.
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b)
}
