package middleware

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type traceKey struct{}

// TraceInfo holds distributed tracing identifiers for a request.
type TraceInfo struct {
	TraceID string
	SpanID  string
	Parent  string
}

// TraceFromContext retrieves TraceInfo from the request context.
func TraceFromContext(ctx context.Context) (TraceInfo, bool) {
	v, ok := ctx.Value(traceKey{}).(TraceInfo)
	return v, ok
}

// Tracing injects or propagates W3C traceparent tracing headers.
// If an incoming traceparent header is present it is parsed and a child
// span is created; otherwise a new root trace is started.
func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var info TraceInfo

		if parent := r.Header.Get("Traceparent"); parent != "" {
			info = parseTraceparent(parent)
			info.SpanID = newID(8)
		} else {
			info = TraceInfo{
				TraceID: newID(16),
				SpanID:  newID(8),
			}
		}

		traceparent := fmt.Sprintf("00-%s-%s-01", info.TraceID, info.SpanID)
		r.Header.Set("Traceparent", traceparent)
		w.Header().Set("Traceparent", traceparent)

		ctx := context.WithValue(r.Context(), traceKey{}, info)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// parseTraceparent extracts the trace-id and span-id from a W3C traceparent
// header value (version 00 only). On parse failure a new root trace is returned.
func parseTraceparent(header string) TraceInfo {
	// format: 00-<traceID>-<parentSpanID>-<flags>
	var version, traceID, spanID, flags string
	n, _ := fmt.Sscanf(header, "%2s-%32s-%16s-%2s", &version, &traceID, &spanID, &flags)
	if n != 4 || version != "00" {
		return TraceInfo{TraceID: newID(16), SpanID: newID(8)}
	}
	return TraceInfo{
		TraceID: traceID,
		SpanID:  spanID,
		Parent:  spanID,
	}
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

func newID(bytes int) string {
	b := make([]byte, bytes)
	_, _ = rng.Read(b)
	return fmt.Sprintf("%x", b)
}
