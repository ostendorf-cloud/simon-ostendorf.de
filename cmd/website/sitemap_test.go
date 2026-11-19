package main

import (
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/webapp"
)

// TestPrerenderedSitemapMatchesServedSitemap proves the sitemap.xml
// "task prerender" writes, the file govue-check-legal reads, is byte for byte
// the sitemap the server serves for the default SITE_URL. The server's own
// sitemap stays the one that is served: the prerendered copy exists only as
// the check's input, and this test is what stops it from drifting.
func TestPrerenderedSitemapMatchesServedSitemap(t *testing.T) {
	dist, err := fs.Sub(distFS, "dist")
	if err != nil {
		t.Fatalf("fs.Sub(dist) error = %v", err)
	}
	prerendered, err := fs.Sub(prerenderedFS, "prerendered")
	if err != nil {
		t.Fatalf("fs.Sub(prerendered) error = %v", err)
	}

	want, err := fs.ReadFile(prerendered, "sitemap.xml")
	if err != nil {
		t.Fatalf("prerendered has no sitemap.xml, run task prerender: %v", err)
	}

	siteURL, err := url.Parse(config.DefaultSiteURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	app, err := webapp.New(config.Config{SiteURL: siteURL}, slog.New(slog.DiscardHandler), dist, prerendered)
	if err != nil {
		t.Fatalf("webapp.New() error = %v", err)
	}
	defer func() { _ = app.Close() }()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sitemap.xml", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	got, _ := io.ReadAll(rec.Body)

	if string(got) != string(want) {
		t.Fatalf("the served sitemap differs from the prerendered one\nserved:\n%s\nprerendered:\n%s", got, want)
	}
}
