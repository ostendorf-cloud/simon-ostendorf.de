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
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ostendorf-cloud/govue/health"
	"github.com/ostendorf-cloud/govue/metrics"
	"github.com/ostendorf-cloud/govue/middleware"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
)

// App holds every dependency the two handlers below are built from.
type App struct {
	logger  *slog.Logger
	pages   *pageServer
	siteURL *url.URL

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

	registry := metrics.NewRegistry()

	return &App{
		logger:         logger,
		pages:          pages,
		siteURL:        cfg.SiteURL,
		health:         health.New(),
		metricsHandler: metrics.Handler(registry),
	}, nil
}

// Health exposes the health registry so main can pass its Drain method as
// httpserver.Config's BeforeShutdown.
func (a *App) Health() *health.Registry {
	return a.health
}

// Handler is the public handler: the SPA's static assets, the prerendered
// page for every known route and locale, /robots.txt and /sitemap.xml,
// and the catch-all document handler for the locale redirect, trailing
// slash canonicalisation and the 404 page, wrapped in the two middlewares
// worth keeping even for a handler with no session or form of its own. A
// panic serving static files is still worth catching and still worth a
// correlatable log line; everything else govue's middleware package
// offers (security headers, access log, CORS, rate limiting) needs a
// dependency or a decision this site does not have yet, see AGENTS.md.
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/assets/", a.pages.assetHandler())
	mux.Handle("GET /robots.txt", a.robotsHandler())
	mux.Handle("GET /sitemap.xml", a.sitemapHandler())
	registerRoutes(mux, a.pages, func(h http.Handler) http.Handler { return h })
	mux.Handle("/", a.pages.pageHandler())

	return middleware.Chain(mux,
		middleware.Recover(a.logger),
		middleware.RequestID(a.logger),
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
