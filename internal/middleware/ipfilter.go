package middleware

import (
	"net"
	"net/http"
)

// IPFilterConfig defines allowed and blocked CIDR ranges.
type IPFilterConfig struct {
	// AllowCIDRs is an explicit allowlist; if non-empty, only matching IPs are permitted.
	AllowCIDRs []string
	// BlockCIDRs is an explicit blocklist; matching IPs are always denied.
	BlockCIDRs []string
	// TrustProxy controls whether X-Forwarded-For is used to determine client IP.
	TrustProxy bool
}

// DefaultIPFilterConfig returns a permissive config with no rules.
func DefaultIPFilterConfig() IPFilterConfig {
	return IPFilterConfig{TrustProxy: false}
}

type ipFilter struct {
	allowNets []*net.IPNet
	blockNets []*net.IPNet
	cfg       IPFilterConfig
}

func parseCIDRs(cidrs []string) ([]*net.IPNet, error) {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		nets = append(nets, ipNet)
	}
	return nets, nil
}

func containsIP(nets []*net.IPNet, ip net.IP) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// IPFilter returns middleware that enforces IP-based allow/block rules.
func IPFilter(cfg IPFilterConfig) (func(http.Handler) http.Handler, error) {
	allowNets, err := parseCIDRs(cfg.AllowCIDRs)
	if err != nil {
		return nil, err
	}
	blockNets, err := parseCIDRs(cfg.BlockCIDRs)
	if err != nil {
		return nil, err
	}

	f := &ipFilter{allowNets: allowNets, blockNets: blockNets, cfg: cfg}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := net.ParseIP(clientIP(r, cfg.TrustProxy))
			if ip == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if len(f.blockNets) > 0 && containsIP(f.blockNets, ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if len(f.allowNets) > 0 && !containsIP(f.allowNets, ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}
