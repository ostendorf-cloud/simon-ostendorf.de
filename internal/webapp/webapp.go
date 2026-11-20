// Package webapp builds the two HTTP handlers this website serves: the
// public one, serving the prerendered HTML and the built SPA bundle, and
// the administrative one govue's second listener serves, health,
// readiness and metrics, reached only from inside the cluster.
//
// pages.go is the catch-all document handler: the locale redirect at "/",
// trailing slash canonicalisation, and serving the matching prerendered
// file or a locale's own 404 page. Its own routeFiles is built from
// internal/routes' embedded table, not hand written, so the page handler,
// the per-page mux registration and sitemap.go's /robots.txt and
// /sitemap.xml (which read that same routeFiles/locales) all answer from
// one list. That a page added to routes.json actually gets prerendered,
// rather than merely agreeing with Go's own copy of a list that no
// longer exists, is what cmd/website's
// TestPrerenderedOutputMatchesRouteTable checks against the real output
// "task prerender" wrote.
package webapp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/ostendorf-cloud/govue/analytics"
	"github.com/ostendorf-cloud/govue/clientip"
	"github.com/ostendorf-cloud/govue/health"
	"github.com/ostendorf-cloud/govue/language"
	"github.com/ostendorf-cloud/govue/metrics"
	"github.com/ostendorf-cloud/govue/middleware"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
)

// App holds every dependency the two handlers below are built from.
type App struct {
	logger  *slog.Logger
	pages   *pageServer
	siteURL *url.URL

	// analytics is never nil after New; it is inert when counting is off.
	analytics *analytics.Client
	// accessLog writes one line per request: method, route pattern, status,
	// bytes, duration and the resolved client address. No user agent, no
	// referrer, never the raw path or query. The privacy text names these
	// fields: change both together.
	accessLog middleware.AccessLog
	// closeTimeout bounds Close; a test shortens it.
	closeTimeout time.Duration

	health         *health.Registry
	metricsHandler http.Handler
}

// New constructs an App from cfg. dist is the built SPA and prerendered is
// the static HTML per route and locale, both already resolved to their
// own root (the caller has already called fs.Sub on the embedded
// filesystem).
//
// There is nothing to register against health yet: this site has no
// database, no external API and no cache to ping. Readiness therefore
// answers on the draining flag alone, until a dependency worth checking
// exists.
func New(cfg config.Config, logger *slog.Logger, dist, prerendered fs.FS) (*App, error) {
	pages, err := newPageServer(dist, prerendered)
	if err != nil {
		return nil, err
	}

	// Handler registers the legal redirects and cannot return an error, so a
	// bad legalConfig is reported here, once, on a mux of its own.
	if err := language.HandleLegal(http.NewServeMux(), legalConfig, nil); err != nil {
		return nil, fmt.Errorf("webapp: %w", err)
	}

	registry := metrics.NewRegistry()

	// Page view counting and the access log share one resolver. It trusts
	// X-Forwarded-For only from the configured proxies, so a visitor behind
	// the ingress is not mistaken for the ingress. The hostname reported is
	// the public host from SITE_URL, never the request's Host header. A zero
	// Config gives an inert client.
	clientIPs := clientip.NewResolver(cfg.TrustedProxies)
	views, err := analytics.New(cfg.Umami, analytics.Options{
		Hostname: cfg.SiteURL.Hostname(),
		Logger:   logger,
		ClientIP: clientIPs,
		OnDrop: metrics.NewCounter(registry, "analytics_views_dropped_total",
			"Page views lost to a full queue or a shutdown before they reached Umami."),
	})
	if err != nil {
		return nil, fmt.Errorf("webapp: %w", err)
	}
	// govue logs nothing when counting is off, so say it once here. The line
	// never carries the Umami URL.
	if views.Enabled() {
		logger.Info("page view counting is on")
	} else {
		logger.Info("page view counting is off: UMAMI_URL and UMAMI_WEBSITE_ID are not set")
	}

	return &App{
		logger:         logger,
		pages:          pages,
		siteURL:        cfg.SiteURL,
		analytics:      views,
		accessLog:      middleware.AccessLog{ClientIP: clientIPs},
		closeTimeout:   analyticsCloseTimeout,
		health:         health.New(),
		metricsHandler: metrics.Handler(registry),
	}, nil
}

// legalConfig is the legal redirect set: German and English, German default,
// no terms.
var legalConfig = language.LegalConfig{Config: language.Config{Languages: locales, Default: defaultLocale}}

// analyticsCloseTimeout bounds how long Close waits for the queued page
// views to be posted.
const analyticsCloseTimeout = 5 * time.Second

// Close sends the page views still queued, within analyticsCloseTimeout. Call
// it after the public listener has stopped. Views that did not make it in
// time are lost, which is logged at warn level and is no failure: a slow Umami
// must not turn a normal SIGTERM into a failed exit. Any other error is
// returned.
func (a *App) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.closeTimeout)
	defer cancel()
	err := a.analytics.Close(ctx)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		a.logger.Warn("page views still queued at shutdown were not sent", "error", err)
		return nil
	}
	return err
}

// Health exposes the health registry so main can pass its Drain method as
// httpserver.Config's BeforeShutdown.
func (a *App) Health() *health.Registry {
	return a.health
}

// Handler is the public handler: the SPA's static assets, the prerendered
// page for every known route and locale, /robots.txt and /sitemap.xml,
// and the catch-all document handler for the locale redirect, trailing
// slash canonicalisation and the 404 page, wrapped in panic recovery, a
// request id and govue's access log (one line per request, route pattern
// only, client address resolved behind the trusted proxies, no user agent).
// Everything else govue's middleware package offers (security headers, CORS,
// rate limiting) needs a dependency or a decision this site does not have
// yet, see AGENTS.md.
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/assets/", a.pages.assetHandler())
	mux.Handle("GET /robots.txt", a.robotsHandler())
	mux.Handle("GET /sitemap.xml", a.sitemapHandler())
	registerRoutes(mux, a.pages, func(h http.Handler) http.Handler { return h })
	// govue's legal redirects: /imprint, /privacy and the German aliases
	// /impressum, /datenschutz and /datenschutzerklaerung. They answer 302,
	// which analytics never counts. This site has no terms, so /terms and
	// /agb stay unregistered and fall to the 404 below. The catch-all must be
	// method-less, see language.HandleLegal.
	if err := language.HandleLegal(mux, legalConfig, nil); err != nil {
		panic("webapp: legal routes: " + err.Error())
	}
	mux.Handle("/", a.pages.pageHandler())

	// Analytics is innermost, directly around the mux: it reads the route
	// pattern the mux sets, and RequestID replaces the request with
	// WithContext, so any middleware between would hide the pattern. It is
	// the mux itself when counting is off. The access log sits outside it
	// and inside RequestID, so its line carries the request id and, as
	// analytics never replaces the request, still sees the route pattern.
	return middleware.Chain(mux,
		middleware.Recover(a.logger),
		middleware.RequestID(a.logger),
		a.accessLog.Middleware,
		a.analytics.Middleware,
	)
}

// AdminHandler is the second, smaller handler govue's admin listener serves:
// health, readiness and metrics, never reached through this chart's own
// Ingress. The middleware here is the same short list AdminHandler's model
// in govue's own example package uses: a scrape or a probe is not browser
// traffic, so there is nothing for security headers, CORS or a rate limiter
// to protect here, but a panic is still worth catching and still worth a
// correlatable log line.
func (a *App) AdminHandler() http.Handler {
	mux := http.NewServeMux()
	a.health.Handle(mux)
	mux.Handle("GET /metrics", a.metricsHandler)

	return middleware.Chain(mux,
		middleware.Recover(a.logger),
		middleware.RequestID(a.logger),
	)
}
