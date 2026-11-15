// Command website serves simon-ostendorf.de: the prerendered HTML and the
// built Vue SPA on its public listener, and the operational endpoints
// /healthz, /readyz and /metrics on a second, administrative listener,
// reached only from inside the cluster.
package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"github.com/ostendorf-cloud/govue/httpserver"
	"github.com/ostendorf-cloud/govue/logging"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/webapp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	logger := logging.New(os.Stdout, cfg.LogLevel)

	dist, err := fs.Sub(distFS, "dist")
	if err != nil {
		return fmt.Errorf("website: %w", err)
	}
	prerendered, err := fs.Sub(prerenderedFS, "prerendered")
	if err != nil {
		return fmt.Errorf("website: %w", err)
	}

	app, err := webapp.New(cfg, logger, dist, prerendered)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return httpserver.RunAdmin(ctx, httpserver.Config{
		Addr:    cfg.Addr,
		Handler: app.Handler(),
		Logger:  logger,
		// Failing the readiness probe before the drain delay starts is
		// what lets a load balancer stop sending new requests before the
		// server stops accepting connections at all. This runs once, for
		// both listeners, however many are running.
		BeforeShutdown: app.Health().Drain,
	}, httpserver.AdminConfig{
		Addr:    cfg.AdminAddr,
		Handler: app.AdminHandler(),
	})
}
