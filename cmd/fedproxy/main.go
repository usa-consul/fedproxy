package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/agency/fedproxy/internal/config"
	"github.com/agency/fedproxy/internal/middleware"
	"github.com/agency/fedproxy/internal/proxy"
)

func main() {
	cfgPath := flag.String("config", "fedproxy.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	rp, err := proxy.New(cfg)
	if err != nil {
		log.Fatalf("failed to create reverse proxy: %v", err)
	}

	logger := log.Default()
	var handler http.Handler = rp

	handler = middleware.RequireAuth(cfg, handler)

	if cfg.Rate.Enabled {
		window := time.Minute
		limiter := middleware.NewRateLimiter(cfg.Rate.RequestsPerMin, window)
		handler = middleware.RateLimit(limiter)(handler)
	}

	handler = middleware.RequestLogger(logger)(handler)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("fedproxy listening on %s -> %s", cfg.Addr, cfg.Upstream)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
