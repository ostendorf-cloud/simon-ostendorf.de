package webapp

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ostendorf-cloud/govue/analytics"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
)

const (
	testWebsiteID = "123e4567-e89b-42d3-a456-426614174000"
	browserAgent  = "Mozilla/5.0 (X11; Linux x86_64) Firefox/130.0"
)

// umamiStub records what the site posts to Umami's send endpoint.
type umamiStub struct {
	server *httptest.Server

	mu     sync.Mutex
	bodies []string
}

// newUmamiStub starts a stub on a loopback address, which govue accepts over
// plain HTTP. block, when not nil, holds every post until it is closed.
func newUmamiStub(t *testing.T, block <-chan struct{}) *umamiStub {
	t.Helper()
	s := &umamiStub{}
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		s.mu.Lock()
		s.bodies = append(s.bodies, string(raw))
		s.mu.Unlock()
		if block != nil {
			<-block
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(s.server.Close)
	return s
}

type sentPayload struct {
	Website   string `json:"website"`
	Hostname  string `json:"hostname"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Referrer  string `json:"referrer"`
	Language  string `json:"language"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
}

func (s *umamiStub) rawBodies() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.bodies...)
}

func (s *umamiStub) payloads(t *testing.T) []sentPayload {
	t.Helper()
	var out []sentPayload
	for _, raw := range s.rawBodies() {
		var ev struct {
			Payload sentPayload `json:"payload"`
		}
		if err := json.Unmarshal([]byte(raw), &ev); err != nil {
			t.Fatalf("decode %q: %v", raw, err)
		}
		out = append(out, ev.Payload)
	}
	return out
}

// countingApp builds an App over every route of both locales, with counting
// pointed at stub, or off when stub is nil.
func countingApp(t *testing.T, stub *umamiStub, proxies []netip.Prefix, logger *slog.Logger) *App {
	t.Helper()

	page := &fstest.MapFile{Data: []byte("<!doctype html><title>page</title>")}
	prerendered := fstest.MapFS{}
	for _, locale := range locales {
		prerendered[locale+"/404.html"] = &fstest.MapFile{Data: []byte("<!doctype html><title>404</title>")}
		for _, file := range routeFiles {
			prerendered[locale+"/"+file] = page
		}
	}
	dist := fstest.MapFS{"assets/app.js": &fstest.MapFile{Data: []byte("console.log('app')")}}

	siteURL, err := url.Parse("https://simon-ostendorf.de")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	cfg := config.Config{SiteURL: siteURL, TrustedProxies: proxies}
	if stub != nil {
		cfg.Umami = analytics.Config{URL: stub.server.URL, WebsiteID: testWebsiteID}
	}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	app, err := New(cfg, logger, dist, prerendered)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return app
}

// visitor is a GET from a public address with a browser user agent.
func visitor(t *testing.T, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	req.Header.Set("User-Agent", browserAgent)
	req.RemoteAddr = "203.0.113.77:51000"
	return req
}

func serve(app *App, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	return rec
}

func closeApp(t *testing.T, app *App) {
	t.Helper()
	if err := app.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

// TestEveryPageIsCountedUnderItsRoutePattern proves each route of each locale
// is reported with its pattern path, through the real handler chain.
func TestEveryPageIsCountedUnderItsRoutePattern(t *testing.T) {
	stub := newUmamiStub(t, nil)
	app := countingApp(t, stub, nil, nil)

	var want []string
	for _, locale := range locales {
		for _, slug := range routeSlugs() {
			path := "/" + locale + "/"
			if slug != "" {
				path += slug + "/"
			}
			if rec := serve(app, visitor(t, path)); rec.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", path, rec.Code)
			}
			want = append(want, path)
		}
	}
	closeApp(t, app)

	got := map[string]bool{}
	for _, p := range stub.payloads(t) {
		got[p.URL] = true
		if p.Website != testWebsiteID {
			t.Errorf("website = %q, want %q", p.Website, testWebsiteID)
		}
		if p.Hostname != "simon-ostendorf.de" {
			t.Errorf("hostname = %q, want the SITE_URL host", p.Hostname)
		}
		if p.IP != "203.0.113.0" {
			t.Errorf("ip = %q, want the /24 mask 203.0.113.0", p.IP)
		}
	}
	for _, path := range want {
		if !got[path] {
			t.Errorf("no view reported for %s, got %v", path, got)
		}
	}
	if len(stub.rawBodies()) != len(want) {
		t.Errorf("reported %d views, want %d", len(stub.rawBodies()), len(want))
	}
}

// TestOnlyPagesAreCounted proves redirects, 404s, assets, the sitemap, robots,
// HEAD, Do Not Track, Global Privacy Control, a prefetch and a bot count
// nothing. One real page at the end proves the stub works.
func TestOnlyPagesAreCounted(t *testing.T) {
	stub := newUmamiStub(t, nil)
	app := countingApp(t, stub, nil, nil)

	cases := []struct {
		name   string
		path   string
		method string
		header map[string]string
	}{
		{name: "root redirect", path: "/"},
		{name: "missing trailing slash redirect", path: "/de/cv"},
		{name: "unknown path 404", path: "/de/does-not-exist/"},
		{name: "path outside any locale", path: "/nope"},
		{name: "static asset", path: "/assets/app.js"},
		{name: "robots", path: "/robots.txt"},
		{name: "sitemap", path: "/sitemap.xml"},
		{name: "HEAD", path: "/de/cv/", method: http.MethodHead},
		{name: "Do Not Track", path: "/de/cv/", header: map[string]string{"DNT": "1"}},
		{name: "Global Privacy Control", path: "/de/cv/", header: map[string]string{"Sec-GPC": "1"}},
		{name: "prefetch", path: "/de/cv/", header: map[string]string{"Sec-Purpose": "prefetch"}},
		{name: "bot", path: "/de/cv/", header: map[string]string{"User-Agent": "Googlebot/2.1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := visitor(t, tc.path)
			if tc.method != "" {
				req.Method = tc.method
			}
			for k, v := range tc.header {
				req.Header.Set(k, v)
			}
			serve(app, req)
		})
	}
	serve(app, visitor(t, "/en/privacy/"))
	closeApp(t, app)

	got := stub.payloads(t)
	if len(got) != 1 || got[0].URL != "/en/privacy/" {
		t.Fatalf("reported %+v, want exactly one view for /en/privacy/", got)
	}
}

// TestQueryStringNeverLeaves proves a token in the query, a fragment-like
// value and a referrer path never reach Umami.
func TestQueryStringNeverLeaves(t *testing.T) {
	stub := newUmamiStub(t, nil)
	app := countingApp(t, stub, nil, nil)

	req := visitor(t, "/de/cv/?token=s3cr3t-token&id=4711")
	req.Header.Set("Referer", "https://other.example.test/reset/s3cr3t-path?k=s3cr3t-ref")
	if rec := serve(app, req); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	closeApp(t, app)

	raws := stub.rawBodies()
	if len(raws) != 1 {
		t.Fatalf("reported %d views, want 1", len(raws))
	}
	for _, secret := range []string{"s3cr3t", "4711", "token", "?"} {
		if strings.Contains(raws[0], secret) {
			t.Errorf("payload %s contains %q", raws[0], secret)
		}
	}
	p := stub.payloads(t)[0]
	if p.URL != "/de/cv/" {
		t.Errorf("url = %q, want the pattern path /de/cv/", p.URL)
	}
	if p.Referrer != "https://other.example.test/" {
		t.Errorf("referrer = %q, want the origin only", p.Referrer)
	}
}

// TestSlowUmamiNeverDelaysAResponse proves a stub that holds every post does
// not hold a page.
func TestSlowUmamiNeverDelaysAResponse(t *testing.T) {
	release := make(chan struct{})
	stub := newUmamiStub(t, release)
	app := countingApp(t, stub, nil, nil)

	done := make(chan int, 1)
	go func() { done <- serve(app, visitor(t, "/de/")).Code }()
	select {
	case code := <-done:
		if code != http.StatusOK {
			t.Fatalf("status = %d, want 200", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the response waited for Umami")
	}
	close(release)
	closeApp(t, app)
}

// TestForwardedClientIsCountedOnlyBehindATrustedProxy proves the visitor behind
// the ingress is counted with their own masked address when the proxy is
// trusted, and not counted at all when it is not.
func TestForwardedClientIsCountedOnlyBehindATrustedProxy(t *testing.T) {
	behindProxy := func(t *testing.T) *http.Request {
		req := visitor(t, "/de/")
		req.RemoteAddr = "10.1.2.3:4000"
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		return req
	}

	t.Run("trusted", func(t *testing.T) {
		stub := newUmamiStub(t, nil)
		app := countingApp(t, stub, []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}, nil)
		serve(app, behindProxy(t))
		closeApp(t, app)
		got := stub.payloads(t)
		if len(got) != 1 || got[0].IP != "198.51.100.0" {
			t.Fatalf("reported %+v, want one view with ip 198.51.100.0", got)
		}
	})

	t.Run("untrusted", func(t *testing.T) {
		stub := newUmamiStub(t, nil)
		app := countingApp(t, stub, nil, nil)
		serve(app, behindProxy(t))
		closeApp(t, app)
		if got := stub.rawBodies(); len(got) != 0 {
			t.Fatalf("reported %v, want nothing from a private peer", got)
		}
	})
}

// TestCountingIsOffWithoutConfiguration proves the site serves normally and
// says once that counting is off, without naming any URL.
func TestCountingIsOffWithoutConfiguration(t *testing.T) {
	var logs bytes.Buffer
	app := countingApp(t, nil, nil, slog.New(slog.NewTextHandler(&logs, nil)))

	if rec := serve(app, visitor(t, "/de/")); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	closeApp(t, app)

	if n := strings.Count(logs.String(), "page view counting is off"); n != 1 {
		t.Errorf("off line logged %d times, want once: %q", n, logs.String())
	}
}
