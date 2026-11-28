package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/apiconfig"
)

// newTestApp builds the API with a logger writing into the returned buffer.
func newTestApp(t *testing.T, cfg apiconfig.Config) (*app, *bytes.Buffer) {
	t.Helper()
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	return newApp(cfg, logger), &logs
}

func get(t *testing.T, h http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), method, target, nil)
	req.RemoteAddr = "203.0.113.7:4000"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestEveryPublicPathAnswersAJSON404(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t, apiconfig.Config{})

	for _, path := range []string{"/", "/api/v1/anything", "/de/", "/healthz", "/metrics", "/nope"} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			t.Run(method+" "+path, func(t *testing.T) {
				t.Parallel()

				rec := get(t, a.handler, method, path)
				if rec.Code != http.StatusNotFound {
					t.Fatalf("status = %d, want 404", rec.Code)
				}
				if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
					t.Errorf("Content-Type = %q, want JSON", ct)
				}
				var body struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("body %q is not JSON: %v", rec.Body.String(), err)
				}
				if body.Error.Code != "not_found" {
					t.Errorf("error code = %q, want not_found", body.Error.Code)
				}
			})
		}
	}
}

func TestPublicAnswerCarriesTheSecurityHeaders(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t, apiconfig.Config{})
	rec := get(t, a.handler, http.MethodGet, "/api/v1/x")

	for _, name := range []string{"Strict-Transport-Security", "X-Content-Type-Options", "Content-Security-Policy"} {
		if rec.Header().Get(name) == "" {
			t.Errorf("%s is missing", name)
		}
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Error("X-Request-Id is missing")
	}
}

func TestPlainHTTPDropsOnlyHSTS(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t, apiconfig.Config{AllowPlainHTTP: true})
	rec := get(t, a.handler, http.MethodGet, "/")

	if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("Strict-Transport-Security = %q, want none", got)
	}
	if rec.Header().Get("X-Content-Type-Options") == "" {
		t.Error("X-Content-Type-Options is missing")
	}
}

func TestAdminListenerServesHealthAndMetricsOnly(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t, apiconfig.Config{})

	for _, path := range []string{"/healthz", "/readyz"} {
		if rec := get(t, a.admin, http.MethodGet, path); rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, rec.Code)
		}
	}

	// The public chain has counted one request by now.
	get(t, a.handler, http.MethodGet, "/nope")
	rec := get(t, a.admin, http.MethodGet, "/metrics")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "http_requests_total") {
		t.Errorf("metrics lack the request counter:\n%s", rec.Body.String())
	}

	if rec := get(t, a.admin, http.MethodGet, "/nope"); rec.Code != http.StatusNotFound {
		t.Errorf("GET /nope on admin = %d, want 404", rec.Code)
	}
}

func TestPublicListenerDoesNotServeProbesOrMetrics(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t, apiconfig.Config{})
	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		if rec := get(t, a.handler, http.MethodGet, path); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s on the public listener = %d, want 404", path, rec.Code)
		}
	}
}

func TestDrainFailsReadinessButNotLiveness(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t, apiconfig.Config{})
	a.health.Drain()

	if rec := get(t, a.admin, http.MethodGet, "/readyz"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("GET /readyz while draining = %d, want 503", rec.Code)
	}
	if rec := get(t, a.admin, http.MethodGet, "/healthz"); rec.Code != http.StatusOK {
		t.Errorf("GET /healthz while draining = %d, want 200", rec.Code)
	}
}

func TestAccessLogWritesOneLineWithThePatternAndNeverTheQuery(t *testing.T) {
	t.Parallel()

	a, logs := newTestApp(t, apiconfig.Config{})
	get(t, a.handler, http.MethodGet, "/api/v1/secret-path?token=s3cr3t")

	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("logged %d lines, want 1:\n%s", len(lines), logs.String())
	}
	line := lines[0]
	if !strings.Contains(line, "status=404") {
		t.Errorf("line lacks the status: %s", line)
	}
	for _, leaked := range []string{"s3cr3t", "secret-path", "token="} {
		if strings.Contains(line, leaked) {
			t.Errorf("line leaks %q: %s", leaked, line)
		}
	}
}

func TestAccessLogResolvesTheClientBehindATrustedProxy(t *testing.T) {
	t.Parallel()

	a, logs := newTestApp(t, apiconfig.Config{TrustedProxies: []netip.Prefix{netip.MustParsePrefix("203.0.113.0/24")}})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.7:4000"
	req.Header.Set("X-Forwarded-For", "198.51.100.9")
	a.handler.ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(logs.String(), "client=198.51.100.9") {
		t.Errorf("log does not name the forwarded client:\n%s", logs.String())
	}
}

func TestAccessLogSkipsTheProbes(t *testing.T) {
	t.Parallel()

	a, logs := newTestApp(t, apiconfig.Config{})
	get(t, a.admin, http.MethodGet, "/healthz")
	get(t, a.admin, http.MethodGet, "/readyz")

	if logs.Len() != 0 {
		t.Errorf("probes were logged:\n%s", logs.String())
	}
}
