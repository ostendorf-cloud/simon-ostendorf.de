package webapp

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
)

// testApp builds an App with a fake SPA bundle, a fake prerendered output
// and a logger that discards everything, so a test asserts on behaviour,
// not log noise.
func testApp(t *testing.T) *App {
	t.Helper()

	dist := fstest.MapFS{
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('app')")},
	}
	prerendered := fstest.MapFS{
		"de/index.html": &fstest.MapFile{Data: []byte("<!doctype html><title>test</title>")},
		"de/404.html":   &fstest.MapFile{Data: []byte("<!doctype html><title>404</title>")},
		"en/index.html": &fstest.MapFile{Data: []byte("<!doctype html><title>test en</title>")},
		"en/404.html":   &fstest.MapFile{Data: []byte("<!doctype html><title>404 en</title>")},
	}

	siteURL, err := url.Parse("https://simon-ostendorf.de")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	logger := slog.New(slog.DiscardHandler)
	app, err := New(config.Config{SiteURL: siteURL}, logger, dist, prerendered)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return app
}

// TestHandlerServesRootRedirectAndKnownRoute proves the public handler
// answers "/" with the locale redirect and a known route with its
// prerendered page, through the real middleware chain.
func TestHandlerServesRootRedirectAndKnownRoute(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("GET / status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "/de/" {
		t.Fatalf("GET / Location = %q, want %q", got, "/de/")
	}

	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/de/", nil)
	rec = httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /de/ status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "<title>test</title>") {
		t.Fatalf("GET /de/ body = %q, want it to contain the embedded prerendered page", rec.Body.String())
	}
}

// TestHandlerRegistersRobotsAndSitemapAlongsideTheRoutes proves the mux
// built from registerRoutes' per-locale patterns, the asset handler and
// the catch-all document handler does not conflict with /robots.txt and
// /sitemap.xml: both are reachable through the real middleware chain, not
// swallowed by a more general pattern.
func TestHandlerRegistersRobotsAndSitemapAlongsideTheRoutes(t *testing.T) {
	app := testApp(t)

	for _, path := range []string{"/robots.txt", "/sitemap.xml"} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		app.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, rec.Code, http.StatusOK)
		}
	}
}

// TestAdminHandlerServesHealthReadinessAndMetrics proves the admin listener
// answers all three operational endpoints, independently of the public
// handler above.
func TestAdminHandlerServesHealthReadinessAndMetrics(t *testing.T) {
	app := testApp(t)
	admin := app.AdminHandler()

	for _, path := range []string{"/healthz", "/readyz"} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		admin.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", path, rec.Code, http.StatusOK)
		}
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	admin.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want %d", rec.Code, http.StatusOK)
	}
	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read /metrics body: %v", err)
	}
	if !strings.Contains(string(body), "# HELP") {
		t.Fatalf("GET /metrics body does not look like a Prometheus exposition: %q", string(body))
	}
}
