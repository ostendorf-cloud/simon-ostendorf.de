// Package apiconfig loads and validates the API image's environment once,
// in cmd/api's main, aggregating every problem into one error. The API holds
// no secret yet: it has no endpoint, so no variable here is a key, a token or
// a password.
package apiconfig

import (
	"errors"
	"fmt"
	"log/slog"
	"net/netip"

	"github.com/ostendorf-cloud/govue/clientip"
	"github.com/ostendorf-cloud/govue/envconf"
	"github.com/ostendorf-cloud/govue/logging"
)

// Config is every setting the API reads from the environment.
type Config struct {
	// Addr is the public HTTP listener's address.
	Addr string
	// AdminAddr is the admin listener's address: /healthz, /readyz and
	// /metrics, reached only from inside the cluster, never through the
	// chart's own Ingress.
	AdminAddr string
	LogLevel  slog.Level
	// TrustedProxies are the reverse proxies whose X-Forwarded-For header is
	// believed when resolving a client's address for the access log. Empty
	// trusts nothing, so a client behind a proxy then logs as the proxy.
	TrustedProxies []netip.Prefix
	// AllowPlainHTTP drops the Strict-Transport-Security header, for a plain
	// HTTP development run.
	AllowPlainHTTP bool
}

// Load reads every setting through envconf, so a misconfigured deployment
// reports every mistake on the first start instead of one per restart.
func Load(getenv func(string) string) (Config, error) {
	l := envconf.New(getenv)

	cfg := Config{
		Addr:           l.String("ADDR", ":8080"),
		AdminAddr:      l.String("ADMIN_ADDR", ":8081"),
		AllowPlainHTTP: l.Bool("ALLOW_PLAIN_HTTP", false),
	}

	var problems []error
	level, err := logging.Level(l.String("LOG_LEVEL", "info"))
	if err != nil {
		problems = append(problems, err)
	}
	cfg.LogLevel = level

	proxies, err := clientip.ParseCIDRs(l.String("TRUSTED_PROXIES", ""))
	if err != nil {
		problems = append(problems, err)
	}
	cfg.TrustedProxies = proxies

	if cfg.AdminAddr == "" {
		problems = append(problems, errors.New("ADMIN_ADDR must not be empty, the probes and the metrics live there"))
	}

	if err := errors.Join(append(problems, l.Err())...); err != nil {
		return Config{}, fmt.Errorf("apiconfig: %w", err)
	}
	return cfg, nil
}
