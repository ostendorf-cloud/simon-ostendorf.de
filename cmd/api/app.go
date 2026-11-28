package main

import (
	"log/slog"
	"net/http"

	"github.com/ostendorf-cloud/govue/api"
	"github.com/ostendorf-cloud/govue/clientip"
	"github.com/ostendorf-cloud/govue/health"
	"github.com/ostendorf-cloud/govue/metrics"
	"github.com/ostendorf-cloud/govue/middleware"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/apiconfig"
)

// app is the API: the public handler, the admin handler and the readiness
// registry main drains before shutdown.
type app struct {
	handler http.Handler
	admin   http.Handler
	health  *health.Registry
}

// newApp builds both handlers from cfg. The chains follow govue's example:
// the public one is Recover, RequestID, the security headers, the metrics and
// the access log around the mux, the admin one only Recover and RequestID.
// There is no rate limit and no session: the API has no endpoint yet.
func newApp(cfg apiconfig.Config, logger *slog.Logger) *app {
	resolver := clientip.NewResolver(cfg.TrustedProxies)

	registry := metrics.NewRegistry()
	httpMetrics := metrics.NewHTTP(registry)

	security := middleware.DefaultSecurity()
	if cfg.AllowPlainHTTP {
		// A browser that reached this service over plain HTTP must never
		// be told to upgrade every future visit to HTTPS.
		security.HSTS = 0
	}
	accessLog := middleware.AccessLog{ClientIP: resolver}

	// Every path is a JSON 404: under /api an unknown route, and any other
	// path one the Ingress never sends here.
	mux := http.NewServeMux()
	mux.Handle("/", api.NotFoundHandler())

	h := health.New()
	adminMux := http.NewServeMux()
	h.Handle(adminMux)
	adminMux.Handle("GET /metrics", metrics.Handler(registry))

	return &app{
		health: h,
		handler: middleware.Chain(mux,
			middleware.Recover(logger),
			middleware.RequestID(logger),
			security.Middleware,
			httpMetrics.Middleware,
			accessLog.Middleware,
		),
		admin: middleware.Chain(adminMux,
			middleware.Recover(logger),
			middleware.RequestID(logger),
		),
	}
}
