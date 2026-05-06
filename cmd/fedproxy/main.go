package main

import (
	"log"
	"net/http"
	"os"

	"github.com/agency/fedproxy/internal/config"
	"github.com/agency/fedproxy/internal/middleware"
	"github.com/agency/fedproxy/internal/proxy"
)

func main() {
	cfgPath := "fedproxy.yaml"
	if v := os.Getenv("FEDPROXY_CONFIG"); v != "" {
		cfgPath = v
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	rp, err := proxy.New(cfg)
	if err != nil {
		log.Fatalf("proxy: %v", err)
	}

	// Build security headers middleware from config.
	secCfg := middleware.DefaultSecurityHeadersConfig()
	if cfg.Security.HSTSMaxAge != 0 {
		secCfg.HSTSMaxAge = cfg.Security.HSTSMaxAge
	}
	if cfg.Security.ContentSecurityPolicy != "" {
		secCfg.ContentSecurityPolicy = cfg.Security.ContentSecurityPolicy
	}
	if len(cfg.Security.ExtraHeaders) > 0 {
		secCfg.ExtraHeaders = cfg.Security.ExtraHeaders
	}

	var handler http.Handler = rp
	handler = middleware.RequireAuth(cfg, handler)
	if cfg.RateLimit.Enabled {
		rl := middleware.NewRateLimiter(cfg.RateLimit.Requests, cfg.RateLimit.WindowSecs)
		handler = middleware.RateLimit(rl)(handler)
	}
	handler = middleware.SecurityHeaders(secCfg)(handler)
	handler = middleware.RequestLogger(log.Default())(handler)

	log.Printf("fedproxy listening on %s -> %s", cfg.Addr, cfg.Upstream)
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
