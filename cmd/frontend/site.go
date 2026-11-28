package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/ostendorf-cloud/govue/analytics"
	"github.com/ostendorf-cloud/govue/clientip"
	"github.com/ostendorf-cloud/govue/frontend"
	"github.com/ostendorf-cloud/govue/language"
	"github.com/ostendorf-cloud/govue/middleware"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/frontendconfig"
)

// The browser bar colours, the two ends of the theme's --color-bg in
// web/src/theme/simon-ostendorf-de.css: light first, dark second. govue
// writes one theme-color tag per mode from them. The site follows the system
// colour mode, so it passes no theme boot option.
const (
	themeColorLight = "#d2d1cd"
	themeColorDark  = "#070604"
)

// languages are the site's two, English first as the default.
var languages = []string{"en", "de"}

// options are what a test replaces: the client that reaches Umami and the
// transport that reaches the API in development.
type options struct {
	umamiHTTP       *http.Client
	devAPITransport http.RoundTripper
}

// newServer builds govue's frontend server from the build and cfg. It counts
// page views when Umami is configured and forwards /api/ to the API
// container only with the development switch, FRONTEND_DEV_API_PROXY. The
// pages, the sitemap and robots.txt carry the origin the build was made
// with. cfg.SiteURL only names the host page views are reported under.
func newServer(cfg frontendconfig.Config, build fs.FS, logger *slog.Logger, o options) (*frontend.Server, error) {
	resolver := clientip.NewResolver(cfg.TrustedProxies)

	security := middleware.DefaultSecurity()
	if cfg.AllowPlainHTTP {
		// A browser that reached this service over plain HTTP must never
		// be told to upgrade every future visit to HTTPS.
		security.HSTS = 0
	}

	// The whole client address is never sent: FullIP stays off, so govue
	// masks it. frontend registers the drop counter and closes the client.
	views, err := analytics.New(cfg.Umami, analytics.Options{
		Hostname:   cfg.SiteURL.Hostname(),
		Logger:     logger,
		HTTPClient: o.umamiHTTP,
		ClientIP:   resolver,
	})
	if err != nil {
		return nil, fmt.Errorf("frontend: analytics: %w", err)
	}
	logger.Info("page view counting", "enabled", views.Enabled())
	if views.Enabled() && len(cfg.TrustedProxies) == 0 {
		logger.Warn("page view counting is on but TRUSTED_PROXIES is empty: behind a proxy no view is counted")
	}

	s, err := frontend.New(frontend.Config{
		Build:          build,
		Language:       language.Config{Languages: languages, Default: languages[0]},
		ThemeColor:     themeColorLight,
		ThemeColorDark: themeColorDark,
		Security:       &security,
		ClientIP:       resolver,
		Analytics:      views,
		Logger:         logger,

		Development:     cfg.Development(),
		DevAPIProxy:     cfg.DevAPIProxy,
		DevAPITransport: o.devAPITransport,
	})
	if err != nil {
		// Nothing was served, so nothing is queued.
		_ = views.Close(context.Background())
		return nil, err
	}
	return s, nil
}
