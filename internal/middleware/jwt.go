package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type jwtContextKey struct{}

// JWTClaims holds the parsed claims from a validated JWT.
type JWTClaims struct {
	Subject  string
	Expiry   int64
	Issuer   string
	Audience string
	Raw      map[string]interface{}
}

// DefaultJWTConfig returns a JWTConfig with sensible defaults.
func DefaultJWTConfig(secret string) JWTConfig {
	return JWTConfig{
		Secret:       secret,
		HeaderName:   "Authorization",
		ExemptPaths:  []string{"/health", "/metrics"},
		ClockSkewSec: 30,
	}
}

// JWTConfig configures the JWT validation middleware.
type JWTConfig struct {
	Secret       string
	HeaderName   string
	ExemptPaths  []string
	ClockSkewSec int64
}

// ClaimsFromContext retrieves parsed JWT claims stored in the request context.
func ClaimsFromContext(ctx context.Context) (*JWTClaims, bool) {
	v, ok := ctx.Value(jwtContextKey{}).(*JWTClaims)
	return v, ok
}

// JWT validates a HS256-signed JWT bearer token on each request.
func JWT(cfg JWTConfig, next http.Handler) http.Handler {
	exempt := buildExemptSet(cfg.ExemptPaths)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, skip := exempt[r.URL.Path]; skip {
			next.ServeHTTP(w, r)
			return
		}
		raw := r.Header.Get(cfg.HeaderName)
		token := strings.TrimPrefix(raw, "Bearer ")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		claims, err := validateJWT(token, cfg.Secret, cfg.ClockSkewSec)
		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), jwtContextKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateJWT(token, secret string, skewSec int64) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errorf("malformed token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errorf("bad signature encoding")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(mac.Sum(nil), sig) {
		return nil, errorf("signature mismatch")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errorf("bad payload encoding")
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, errorf("payload not JSON")
	}
	claims := &JWTClaims{Raw: raw}
	if sub, ok := raw["sub"].(string); ok {
		claims.Subject = sub
	}
	if iss, ok := raw["iss"].(string); ok {
		claims.Issuer = iss
	}
	if aud, ok := raw["aud"].(string); ok {
		claims.Audience = aud
	}
	if exp, ok := raw["exp"].(float64); ok {
		claims.Expiry = int64(exp)
		if time.Now().Unix() > claims.Expiry+skewSec {
			return nil, errorf("token expired")
		}
	}
	return claims, nil
}

func errorf(msg string) error {
	return &jwtError{msg}
}

type jwtError struct{ msg string }

func (e *jwtError) Error() string { return e.msg }
