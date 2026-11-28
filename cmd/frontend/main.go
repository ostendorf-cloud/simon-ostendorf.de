// Command frontend serves simon-ostendorf.de: the embedded Vite build with
// the prerendered pages written into it, through govue's frontend package,
// on its public listener, and /healthz, /readyz and /metrics on a second,
// administrative listener. It holds no secret and serves no API: the
// Ingress routes /api to cmd/api, and only a development stack forwards
// /api/ through this server (FRONTEND_DEV_API_PROXY).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ostendorf-cloud/govue/logging"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/embedfs"
	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/frontendconfig"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := frontendconfig.Load(os.Getenv)
	if err != nil {
		return err
	}
	logger := logging.New(os.Stdout, cfg.LogLevel)

	build, err := embedfs.Subtree(distFS, "dist")
	if err != nil {
		return fmt.Errorf("frontend: %w", err)
	}
	s, err := newServer(cfg, build, logger, options{})
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Run drains readiness before shutdown and sends the queued page views
	// once both listeners stopped.
	return s.Run(ctx, cfg.Addr, cfg.AdminAddr)
}
