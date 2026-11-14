package webapp

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
)

// testApp builds an App with a fake SPA bundle and a logger that discards
// everything, so a test asserts on behaviour, not log noise.
func testApp(t *testing.T) *App {
	t.Helper()

	dist := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><title>test</title>")},
	}

	logger := slog.New(slog.DiscardHandler)
	app, err := New(config.Config{}, logger, dist)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return app
}

// TestHandlerServesEmbeddedSPA proves the public handler answers "/" with
// the embedded bundle's index.html.
func TestHandlerServesEmbeddedSPA(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "<title>test</title>") {
		t.Fatalf("GET / body = %q, want it to contain the embedded index.html", rec.Body.String())
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
