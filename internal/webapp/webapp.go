// Package webapp builds the two HTTP handlers this website serves: the
// public one, serving the built SPA bundle, and the administrative one
// govue's second listener serves, health, readiness and metrics, reached
// only from inside the cluster.
package webapp

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/ostendorf-cloud/govue/health"
	"github.com/ostendorf-cloud/govue/metrics"
	"github.com/ostendorf-cloud/govue/middleware"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
)

// App holds every dependency the two handlers below are built from.
type App struct {
	logger *slog.Logger
	dist   fs.FS

	health         *health.Registry
	metricsHandler http.Handler
}

// New constructs an App from cfg. dist is the built SPA, already resolved to
// its own root (the caller has already called fs.Sub on the embedded
// filesystem).
//
// There is nothing to register against health yet: this site has no
// database, no external API and no cache to ping. Readiness therefore
// answers on the draining flag alone, until a dependency worth checking
// exists.
func New(_ config.Config, logger *slog.Logger, dist fs.FS) (*App, error) {
	registry := metrics.NewRegistry()

	return &App{
		logger:         logger,
		dist:           dist,
		health:         health.New(),
		metricsHandler: metrics.Handler(registry),
	}, nil
}

// Health exposes the health registry so main can pass its Drain method as
// httpserver.Config's BeforeShutdown.
func (a *App) Health() *health.Registry {
	return a.health
}

// Handler is the public handler: the embedded SPA bundle, served at "/",
// wrapped in the two middlewares worth keeping even for a handler with no
// routes of its own. A panic serving static files is still worth catching
// and still worth a correlatable log line; everything else govue's
// middleware package offers (security headers, access log, CORS, rate
// limiting) needs a dependency or a decision this site does not have yet,
// see AGENTS.md.
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(a.dist))

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
