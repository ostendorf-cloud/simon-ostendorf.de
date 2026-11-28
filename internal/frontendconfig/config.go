// Package frontendconfig loads and validates the frontend image's
// environment once, in cmd/frontend's main, aggregating every problem into
// one error. The frontend holds no secret: no variable here is a key, a
// token or a password.
package frontendconfig

import (
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"

	"github.com/ostendorf-cloud/govue/analytics"
	"github.com/ostendorf-cloud/govue/clientip"
	"github.com/ostendorf-cloud/govue/envconf"
	"github.com/ostendorf-cloud/govue/logging"
)

// DefaultSiteURL is this site's own production origin. The frontend reads
// only its host name, which page views are reported under. The pages, the
// sitemap and robots.txt carry the origin the build was made with.
const DefaultSiteURL = "https://simon-ostendorf.de"

// EnvDevAPIProxy is the development switch: the URL of the API container
// govue's frontend forwards /api/ to, so a compose stack has one browser
// origin as the Ingress gives production. Set, it also turns on govue's
// Config.Development, which allows the proxy at all. The Helm chart never
// sets it.
const EnvDevAPIProxy = "FRONTEND_DEV_API_PROXY"

// Config is every setting the frontend reads from the environment. The
// Umami website id and URL are not secrets: the id is public in Umami's own
// script snippet and the URL is an endpoint name.
type Config struct {
	// Addr is the public HTTP listener's address.
	Addr string
	// AdminAddr is the admin listener's address: /healthz, /readyz and
	// /metrics, reached only from inside the cluster, never through the
	// chart's own Ingress.
	AdminAddr string
	// SiteURL is this site's public origin. Its host name is the hostname
	// page views are reported under.
	SiteURL  *url.URL
	LogLevel slog.Level
	// TrustedProxies are the reverse proxies whose X-Forwarded-For header is
	// believed when resolving a visitor's address. Empty trusts nothing, so a
	// visitor behind a proxy then looks like the proxy and is not counted.
	TrustedProxies []netip.Prefix
	// AllowPlainHTTP drops the Strict-Transport-Security header, for a plain
	// HTTP development run.
	AllowPlainHTTP bool
	// Umami is the connection to Umami, the zero value when page view counting
	// is off. It is on only when UMAMI_URL and UMAMI_WEBSITE_ID are both set.
	Umami analytics.Config
	// DevAPIProxy is FRONTEND_DEV_API_PROXY, empty in production.
	DevAPIProxy string
}

// Development reports whether the development switch is set. It is true
// exactly when FRONTEND_DEV_API_PROXY is set, never from another variable.
func (c Config) Development() bool {
	return c.DevAPIProxy != ""
}

// Load reads every setting through envconf, so a misconfigured deployment
// reports every mistake on the first start instead of one per restart.
func Load(getenv func(string) string) (Config, error) {
	l := envconf.New(getenv)

	cfg := Config{
		Addr:           l.String("ADDR", ":8080"),
		AdminAddr:      l.String("ADMIN_ADDR", ":8081"),
		SiteURL:        l.URL("SITE_URL", DefaultSiteURL),
		AllowPlainHTTP: l.Bool("ALLOW_PLAIN_HTTP", false),
		// The literal name, so a search for the variable finds it here.
		DevAPIProxy: l.String("FRONTEND_DEV_API_PROXY", ""),
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

	// LoadConfig reports every problem with the four UMAMI_ variables at
	// once, and a half set pair is one of them.
	cfg.Umami, err = analytics.LoadConfig(getenv)
	if err != nil {
		problems = append(problems, err)
	}

	if err := errors.Join(append(problems, l.Err())...); err != nil {
		return Config{}, fmt.Errorf("frontendconfig: %w", err)
	}
	return cfg, nil
}
