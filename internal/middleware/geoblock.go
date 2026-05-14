package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

// GeoBlockConfig controls country-based IP filtering.
// CountryCodes is a list of ISO 3166-1 alpha-2 codes.
// Block=true denies listed countries; Block=false allows only listed countries.
type GeoBlockConfig struct {
	CountryCodes []string
	Block        bool // true = blocklist, false = allowlist
	Lookup       func(ip string) (string, error)
}

// DefaultGeoBlockConfig returns a permissive no-op configuration.
func DefaultGeoBlockConfig() GeoBlockConfig {
	return GeoBlockConfig{
		Block:  true,
		Lookup: stubLookup,
	}
}

// stubLookup always returns an empty country code (unknown).
func stubLookup(_ string) (string, error) {
	return "", nil
}

func buildCountrySet(codes []string) map[string]struct{} {
	set := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		set[strings.ToUpper(c)] = struct{}{}
	}
	return set
}

// GeoBlock returns middleware that allows or denies requests based on the
// resolved country of the client IP.
func GeoBlock(cfg GeoBlockConfig) func(http.Handler) http.Handler {
	countries := buildCountrySet(cfg.CountryCodes)
	lookup := cfg.Lookup
	if lookup == nil {
		lookup = stubLookup
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := geoClientIP(r)
			country, err := lookup(ip)
			if err != nil {
				// On lookup failure, fail open (allow request).
				next.ServeHTTP(w, r)
				return
			}
			country = strings.ToUpper(country)
			_, listed := countries[country]

			blocked := (cfg.Block && listed) || (!cfg.Block && len(countries) > 0 && !listed)
			if blocked {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "forbidden",
					"country": country,
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func geoClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
