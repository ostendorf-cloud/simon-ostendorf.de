package webapp

import (
	"bytes"
	"io/fs"
	"log/slog"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
)

// panicFS panics when the named file is opened, to drive a handler panic
// through the real chain.
type panicFS struct {
	fs.FS
	name string
}

func (p panicFS) Open(name string) (fs.File, error) {
	if name == p.name {
		panic("boom")
	}
	return p.FS.Open(name)
}

// logApp builds an App over the real routes, logging to the returned buffer.
func logApp(t *testing.T, proxies []netip.Prefix, panicOn string) (*App, *bytes.Buffer) {
	t.Helper()
	page := &fstest.MapFile{Data: []byte("<!doctype html><title>page</title>")}
	prerendered := fstest.MapFS{}
	for _, locale := range locales {
		prerendered[locale+"/404.html"] = &fstest.MapFile{Data: []byte("<!doctype html><title>404</title>")}
		for _, file := range routeFiles {
			prerendered[locale+"/"+file] = page
		}
	}
	var pre fs.FS = prerendered
	if panicOn != "" {
		pre = panicFS{FS: prerendered, name: panicOn}
	}
	dist := fstest.MapFS{"assets/app.js": &fstest.MapFile{Data: []byte("x")}}
	site, err := url.Parse("https://simon-ostendorf.de")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	var logs bytes.Buffer
	app, err := New(config.Config{SiteURL: site, TrustedProxies: proxies},
		slog.New(slog.NewTextHandler(&logs, nil)), dist, pre)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })
	return app, &logs
}

// requestLines returns the access log lines of a buffer.
func requestLines(logs *bytes.Buffer) []string {
	var out []string
	for _, line := range strings.Split(logs.String(), "\n") {
		if strings.Contains(line, "request handled") {
			out = append(out, line)
		}
	}
	return out
}

// TestAccessLogWritesOneLinePerRequestWithThePatternOnly proves, through the
// real chain, one line per request with the route pattern, never the raw path
// or query, no user agent and no referrer.
func TestAccessLogWritesOneLinePerRequestWithThePatternOnly(t *testing.T) {
	app, logs := logApp(t, nil, "")

	req := visitor(t, "/de/cv/?token=s3cr3t-token")
	req.Header.Set("Referer", "https://other.example.test/s3cr3t-path")
	if rec := serve(app, req); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	lines := requestLines(logs)
	if len(lines) != 1 {
		t.Fatalf("got %d access log lines, want 1: %q", len(lines), logs.String())
	}
	line := lines[0]
	for _, want := range []string{"method=GET", "status=200", "bytes=", "duration_ms=", "client=203.0.113.77", "request_id="} {
		if !strings.Contains(line, want) {
			t.Errorf("line lacks %q: %s", want, line)
		}
	}
	if !strings.Contains(line, "route=") || strings.Contains(line, "route=unmatched") {
		t.Errorf("line carries no route pattern: %s", line)
	}
	for _, leak := range []string{"s3cr3t", "token", "?", "Firefox", "Mozilla", "user_agent", "other.example", "referrer", "referer"} {
		if strings.Contains(line, leak) {
			t.Errorf("line leaks %q: %s", leak, line)
		}
	}
}

// TestAccessLogResolvesTheClientBehindATrustedProxy proves the visitor
// address, not the proxy's, is logged when the peer is a trusted proxy, and
// that a public peer cannot spoof it.
func TestAccessLogResolvesTheClientBehindATrustedProxy(t *testing.T) {
	proxies := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}

	t.Run("trusted proxy", func(t *testing.T) {
		app, logs := logApp(t, proxies, "")
		req := visitor(t, "/de/")
		req.RemoteAddr = "10.1.2.3:4000"
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		serve(app, req)
		lines := requestLines(logs)
		if len(lines) != 1 || !strings.Contains(lines[0], "client=198.51.100.7") ||
			strings.Contains(lines[0], "10.1.2.3") {
			t.Fatalf("lines = %q, want one with client=198.51.100.7", lines)
		}
	})

	t.Run("public peer spoofing the header", func(t *testing.T) {
		app, logs := logApp(t, proxies, "")
		req := visitor(t, "/de/")
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		serve(app, req)
		lines := requestLines(logs)
		if len(lines) != 1 || !strings.Contains(lines[0], "client=203.0.113.77") ||
			strings.Contains(lines[0], "198.51.100.7") {
			t.Fatalf("lines = %q, want one with the peer's own address", lines)
		}
	})
}

// TestAccessLogLogsTheRouteOfUnknownPaths proves a 404 on a path carrying an
// id is logged without the path.
func TestAccessLogNeverLogsAnUnknownPath(t *testing.T) {
	app, logs := logApp(t, nil, "")
	serve(app, visitor(t, "/de/reset/s3cr3t-token/"))
	lines := requestLines(logs)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1: %q", len(lines), logs.String())
	}
	if strings.Contains(lines[0], "s3cr3t") || strings.Contains(lines[0], "reset") {
		t.Errorf("line leaks the path: %s", lines[0])
	}
}

// TestAPanicIsStillRecoveredAndLoggedOnce proves the recovery line is
// unchanged (method, request id, no path, no client) and the access log
// records a 500.
func TestAPanicIsStillRecoveredAndLoggedOnce(t *testing.T) {
	app, logs := logApp(t, nil, "de/cv/index.html")

	rec := serve(app, visitor(t, "/de/cv/?token=s3cr3t-token"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	lines := requestLines(logs)
	if len(lines) != 1 || !strings.Contains(lines[0], "status=500") {
		t.Fatalf("access lines = %q, want one with status=500", lines)
	}
	var recovered []string
	for _, line := range strings.Split(logs.String(), "\n") {
		if strings.Contains(line, "panic=") {
			recovered = append(recovered, line)
		}
	}
	if len(recovered) != 1 {
		t.Fatalf("got %d recovery lines, want 1: %q", len(recovered), logs.String())
	}
	for _, want := range []string{"method=GET", "request_id="} {
		if !strings.Contains(recovered[0], want) {
			t.Errorf("recovery line lacks %q: %s", want, recovered[0])
		}
	}
	for _, leak := range []string{"s3cr3t", "203.0.113.77", "/de/cv"} {
		if strings.Contains(recovered[0], leak) {
			t.Errorf("recovery line leaks %q: %s", leak, recovered[0])
		}
	}
}
