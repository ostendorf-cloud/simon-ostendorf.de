package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ostendorf-cloud/govue/frontend"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/frontendconfig"
	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/routes"
)

const (
	testWebsiteID = "0f1e2d3c-4b5a-4968-8776-655443322110"
	visitorUA     = "Mozilla/5.0 (X11; Linux x86_64; rv:130.0) Gecko/20100101 Firefox/130.0"
	// visitorAddr is a documentation address that counts as public.
	visitorAddr = "203.0.113.7:4000"
)

func loadConfig(t *testing.T, env map[string]string) frontendconfig.Config {
	t.Helper()
	cfg, err := frontendconfig.Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("frontendconfig.Load() error = %v", err)
	}
	return cfg
}

func serverWithLogger(t *testing.T, env map[string]string, logger *slog.Logger, o options) *frontend.Server {
	t.Helper()
	s, err := newServer(loadConfig(t, env), build(t), logger, o)
	if err != nil {
		t.Fatalf("newServer() error = %v", err)
	}
	return s
}

func server(t *testing.T, env map[string]string, o options) *frontend.Server {
	t.Helper()
	return serverWithLogger(t, env, slog.New(slog.DiscardHandler), o)
}

func visit(t *testing.T, h http.Handler, target string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	req.RemoteAddr = visitorAddr
	req.Header.Set("User-Agent", visitorUA)
	for k, v := range header {
		req.Header[k] = v
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// pagePaths is the canonical path of every route in every language, as
// govue's frontend serves it: the home page with a trailing slash, every
// other page without.
func pagePaths(t *testing.T) []string {
	t.Helper()
	table, err := routes.Load()
	if err != nil {
		t.Fatalf("routes.Load() error = %v", err)
	}
	var out []string
	for _, code := range languages {
		for _, r := range table {
			path := "/" + code + "/"
			if r.Slug != "" {
				path += r.Slug
			}
			out = append(out, path)
		}
	}
	return out
}

// TestSecurityHeadersAndPolicy proves pages and the 404 page carry govue's
// default security headers, the content security policy included.
func TestSecurityHeadersAndPolicy(t *testing.T) {
	t.Parallel()
	h := server(t, nil, options{}).Handler()
	for _, tc := range []struct {
		name   string
		path   string
		status int
	}{
		{"home", "/en/", http.StatusOK},
		{"legal page", "/de/imprint", http.StatusOK},
		{"404 page", "/en/nope", http.StatusNotFound},
		{"path outside any language", "/nope", http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rec := visit(t, h, tc.path, nil)
			if rec.Code != tc.status {
				t.Fatalf("GET %s status = %d, want %d", tc.path, rec.Code, tc.status)
			}
			header := rec.Header()
			if got := header.Get("X-Frame-Options"); got != "DENY" {
				t.Errorf("X-Frame-Options = %q, want DENY", got)
			}
			if got := header.Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
			}
			if got := header.Get("Strict-Transport-Security"); !strings.HasPrefix(got, "max-age=") {
				t.Errorf("Strict-Transport-Security = %q, want a max-age", got)
			}
			if got := header.Get("Content-Security-Policy"); !strings.Contains(got, "default-src") {
				t.Errorf("Content-Security-Policy = %q, want a policy", got)
			}
		})
	}
}

// TestPlainHTTPDropsOnlyHSTS proves ALLOW_PLAIN_HTTP removes only
// Strict-Transport-Security.
func TestPlainHTTPDropsOnlyHSTS(t *testing.T) {
	t.Parallel()
	h := server(t, map[string]string{"ALLOW_PLAIN_HTTP": "true"}, options{}).Handler()
	rec := visit(t, h, "/en/", nil)
	if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("Strict-Transport-Security = %q with ALLOW_PLAIN_HTTP, want none", got)
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" || rec.Header().Get("Content-Security-Policy") == "" {
		t.Error("other security headers are missing")
	}
}

// TestLegalPathsRedirectToTheLanguagePage pins govue's redirect set through
// the real handler: the short paths follow the visitor, the German aliases
// always go to German, and every target answers 200.
func TestLegalPathsRedirectToTheLanguagePage(t *testing.T) {
	t.Parallel()
	h := server(t, nil, options{}).Handler()
	for _, tc := range []struct {
		name   string
		target string
		header http.Header
		want   string
	}{
		{"imprint without a header", "/imprint", nil, "/en/imprint"},
		{"imprint german browser", "/imprint", http.Header{"Accept-Language": {"de-DE,de;q=0.9"}}, "/de/imprint"},
		{"imprint trailing slash", "/imprint/", nil, "/en/imprint"},
		{"privacy english browser", "/privacy", http.Header{"Accept-Language": {"en"}}, "/en/privacy"},
		{"privacy german browser", "/privacy", http.Header{"Accept-Language": {"de"}}, "/de/privacy"},
		{"impressum is always german", "/impressum", http.Header{"Accept-Language": {"en"}}, "/de/imprint"},
		{"datenschutz is always german", "/datenschutz", http.Header{"Accept-Language": {"en"}}, "/de/privacy"},
		{"datenschutzerklaerung", "/datenschutzerklaerung", http.Header{"Accept-Language": {"en"}}, "/de/privacy"},
		{"datenschutzerklaerung slash", "/datenschutzerklaerung/", nil, "/de/privacy"},
		{"query is dropped", "/impressum?next=//evil.example", nil, "/de/imprint"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rec := visit(t, h, tc.target, tc.header)
			if rec.Code != http.StatusFound {
				t.Fatalf("status = %d, want 302", rec.Code)
			}
			if got := rec.Header().Get("Location"); got != tc.want {
				t.Fatalf("Location = %q, want %q", got, tc.want)
			}
			if got := rec.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			if page := visit(t, h, tc.want, nil); page.Code != http.StatusOK {
				t.Errorf("the redirect target %s answers %d, want 200", tc.want, page.Code)
			}
		})
	}
}

// TestSiteWithoutTermsHasNoTermsRoutes pins that this site has no terms
// page: /terms and /agb are not registered and fall to the 404 page.
func TestSiteWithoutTermsHasNoTermsRoutes(t *testing.T) {
	t.Parallel()
	h := server(t, nil, options{}).Handler()
	for _, target := range []string{"/terms", "/agb", "/de/terms", "/en/terms"} {
		if rec := visit(t, h, target, nil); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", target, rec.Code)
		}
	}
}

// TestPagesAreServedWithoutATrailingSlash pins the canonical path of every
// page but the home page and the redirect from its slash form.
func TestPagesAreServedWithoutATrailingSlash(t *testing.T) {
	t.Parallel()
	h := server(t, nil, options{}).Handler()
	for _, target := range []string{
		"/de/cv", "/en/cv", "/de/contact", "/en/contact",
		"/de/imprint", "/en/imprint", "/de/privacy", "/en/privacy",
	} {
		if rec := visit(t, h, target, nil); rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", target, rec.Code)
		}
		rec := visit(t, h, target+"/", nil)
		if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != target {
			t.Errorf("GET %s/ = %d to %q, want 301 to %s", target, rec.Code, rec.Header().Get("Location"), target)
		}
	}
}

// TestRootFollowsTheVisitorsLanguage proves English is the default and an
// Accept-Language of German sends / to the German home page.
func TestRootFollowsTheVisitorsLanguage(t *testing.T) {
	t.Parallel()
	h := server(t, nil, options{}).Handler()
	for name, tc := range map[string]struct {
		header http.Header
		want   string
	}{
		"no header":        {nil, "/en/"},
		"german":           {http.Header{"Accept-Language": {"de"}}, "/de/"},
		"unsupported":      {http.Header{"Accept-Language": {"fr"}}, "/en/"},
		"english with q":   {http.Header{"Accept-Language": {"en-GB,en;q=0.9"}}, "/en/"},
		"german over many": {http.Header{"Accept-Language": {"fr,de;q=0.8,en;q=0.5"}}, "/de/"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rec := visit(t, h, "/", tc.header)
			if rec.Code != http.StatusFound || rec.Header().Get("Location") != tc.want {
				t.Errorf("GET / = %d %q, want 302 %q", rec.Code, rec.Header().Get("Location"), tc.want)
			}
		})
	}
}

// TestPublicFiles proves the favicon files, robots.txt and the sitemap are
// served, and that both name the origin the build was made with.
func TestPublicFiles(t *testing.T) {
	t.Parallel()
	h := server(t, nil, options{}).Handler()
	for path, contentType := range map[string]string{
		"/favicon.svg":          "image/svg+xml",
		"/favicon-32x32.png":    "image/png",
		"/apple-touch-icon.png": "image/png",
		"/robots.txt":           "text/plain; charset=utf-8",
		"/sitemap.xml":          "application/xml; charset=utf-8",
	} {
		rec := visit(t, h, path, nil)
		if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != contentType {
			t.Errorf("GET %s = %d %q, want 200 %q", path, rec.Code, rec.Header().Get("Content-Type"), contentType)
		}
	}
	robots := visit(t, h, "/robots.txt", nil).Body.String()
	if !strings.Contains(robots, "Sitemap: "+frontendconfig.DefaultSiteURL+"/sitemap.xml") {
		t.Errorf("robots.txt = %q, want the production sitemap", robots)
	}
}

// TestSiteURLDoesNotChangeTheSitemap proves SITE_URL only feeds the
// analytics hostname: the sitemap keeps the origin of the build.
func TestSiteURLDoesNotChangeTheSitemap(t *testing.T) {
	t.Parallel()
	h := server(t, map[string]string{"SITE_URL": "https://other.example.test"}, options{}).Handler()
	body := visit(t, h, "/sitemap.xml", nil).Body.String()
	if strings.Contains(body, "other.example.test") {
		t.Error("the sitemap carries SITE_URL, want the build origin")
	}
	if !strings.Contains(body, "<loc>"+frontendconfig.DefaultSiteURL+"/en/</loc>") {
		t.Errorf("the sitemap lacks the build origin: %s", body)
	}
}

// TestThemeColorsMatchTheTheme proves the two browser bar colours are the
// two ends of --color-bg in the theme, and that every page carries one tag
// per mode.
func TestThemeColorsMatchTheTheme(t *testing.T) {
	t.Parallel()
	css, err := os.ReadFile("../../web/src/theme/simon-ostendorf-de.css")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`--color-bg:\s*light-dark\(\s*(#[0-9a-fA-F]{6})\s*,\s*(#[0-9a-fA-F]{6})\s*\)`).FindSubmatch(css)
	if m == nil {
		t.Fatal("simon-ostendorf-de.css declares no --color-bg as light-dark(<hex>, <hex>)")
	}
	if string(m[1]) != themeColorLight || string(m[2]) != themeColorDark {
		t.Errorf("theme colours %s/%s, want the theme's %s/%s", themeColorLight, themeColorDark, m[1], m[2])
	}
	page := visit(t, server(t, nil, options{}).Handler(), "/en/", nil).Body.String()
	for _, tag := range []string{
		`<meta name="theme-color" content="` + themeColorLight + `" media="(prefers-color-scheme: light)">`,
		`<meta name="theme-color" content="` + themeColorDark + `" media="(prefers-color-scheme: dark)">`,
	} {
		if !strings.Contains(page, tag) {
			t.Errorf("page lacks %s", tag)
		}
	}
}

// fakeUmami records every body posted to it and opens no socket.
type fakeUmami struct {
	mu    sync.Mutex
	posts []string
	// block, when not nil, holds every post until it is closed.
	block <-chan struct{}
}

func (f *fakeUmami) RoundTrip(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	f.mu.Lock()
	f.posts = append(f.posts, string(body))
	f.mu.Unlock()
	if f.block != nil {
		select {
		case <-f.block:
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
	return &http.Response{
		StatusCode: http.StatusOK, Header: http.Header{},
		Body: io.NopCloser(strings.NewReader("")), Request: req,
	}, nil
}

type sentPayload struct {
	Website  string `json:"website"`
	Hostname string `json:"hostname"`
	URL      string `json:"url"`
	Referrer string `json:"referrer"`
	IP       string `json:"ip"`
}

func (f *fakeUmami) payloads(t *testing.T) []sentPayload {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []sentPayload
	for _, raw := range f.posts {
		var ev struct {
			Payload sentPayload `json:"payload"`
		}
		if err := json.Unmarshal([]byte(raw), &ev); err != nil {
			t.Fatalf("decode view %q: %v", raw, err)
		}
		out = append(out, ev.Payload)
	}
	return out
}

func (f *fakeUmami) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.posts)
}

func (f *fakeUmami) urls(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, p := range f.payloads(t) {
		out = append(out, p.URL)
	}
	slices.Sort(out)
	return out
}

// countingServer is a server with page view counting on and posting to f.
func countingServer(t *testing.T, f *fakeUmami, env map[string]string) *frontend.Server {
	t.Helper()
	merged := map[string]string{
		"UMAMI_URL":        "https://umami.example.test",
		"UMAMI_WEBSITE_ID": testWebsiteID,
	}
	for k, v := range env {
		merged[k] = v
	}
	return server(t, merged, options{umamiHTTP: &http.Client{Transport: f}})
}

func closeServer(t *testing.T, s *frontend.Server) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if err := s.Close(ctx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

// TestEveryPageIsCountedUnderItsRoutePattern proves each route of each
// language is reported by its pattern path, through the real handler, with
// the site's own host name and the masked client address.
func TestEveryPageIsCountedUnderItsRoutePattern(t *testing.T) {
	t.Parallel()
	fake := &fakeUmami{}
	s := countingServer(t, fake, nil)
	h := s.Handler()

	want := pagePaths(t)
	for _, path := range want {
		if rec := visit(t, h, path, nil); rec.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200", path, rec.Code)
		}
	}
	closeServer(t, s)

	for _, p := range fake.payloads(t) {
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
	slices.Sort(want)
	if got := fake.urls(t); !slices.Equal(got, want) {
		t.Errorf("counted %q, want %q", got, want)
	}
}

// TestAnalyticsCountsOnlyPages proves a page answered with 200 or revalidated
// with 304 is counted by its pattern and without the query, and a redirect,
// an alias, a 404, an asset, a public file, robots.txt, the sitemap, a HEAD
// request, Do Not Track and a bot count nothing.
func TestAnalyticsCountsOnlyPages(t *testing.T) {
	t.Parallel()
	fake := &fakeUmami{}
	s := countingServer(t, fake, nil)
	h := s.Handler()

	page := visit(t, h, "/de/imprint?utm_source=x", nil)
	visit(t, h, "/en/", nil)
	visit(t, h, "/de/imprint", http.Header{"If-None-Match": {page.Header().Get("ETag")}})

	for _, path := range []string{
		"/", "/de", "/de/cv/", "/imprint", "/impressum", "/datenschutz", "/datenschutzerklaerung",
		"/de/imprint/", "/nope", "/de/nope", "/assets/nope.js", "/robots.txt", "/sitemap.xml",
		"/favicon.svg", "/api/v1/anything", "/terms", "/agb",
	} {
		visit(t, h, path, http.Header{"Accept-Language": {"de"}})
	}
	// Do Not Track, Global Privacy Control, a prefetch and a bot.
	for _, header := range []http.Header{
		{"Dnt": {"1"}},
		{"Sec-Gpc": {"1"}},
		{"Sec-Purpose": {"prefetch"}},
		{"User-Agent": {"Googlebot/2.1"}},
	} {
		visit(t, h, "/de/cv", header)
	}
	head := httptest.NewRequestWithContext(t.Context(), http.MethodHead, "/de/cv", nil)
	head.RemoteAddr = visitorAddr
	head.Header.Set("User-Agent", visitorUA)
	h.ServeHTTP(httptest.NewRecorder(), head)

	closeServer(t, s)
	want := []string{"/de/imprint", "/de/imprint", "/en/"}
	if got := fake.urls(t); !slices.Equal(got, want) {
		t.Errorf("counted %q, want %q", got, want)
	}
}

// TestQueryStringNeverLeaves proves a token in the query and a referrer path
// never reach Umami.
func TestQueryStringNeverLeaves(t *testing.T) {
	t.Parallel()
	fake := &fakeUmami{}
	s := countingServer(t, fake, nil)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/de/cv?token=s3cr3t-token&id=4711", nil)
	req.RemoteAddr = visitorAddr
	req.Header.Set("User-Agent", visitorUA)
	req.Header.Set("Referer", "https://other.example.test/reset/s3cr3t-path?k=s3cr3t-ref")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	closeServer(t, s)

	if fake.count() != 1 {
		t.Fatalf("reported %d views, want 1", fake.count())
	}
	for _, raw := range fake.posts {
		for _, secret := range []string{"s3cr3t", "4711", "token", "?"} {
			if strings.Contains(raw, secret) {
				t.Errorf("payload %s contains %q", raw, secret)
			}
		}
	}
	p := fake.payloads(t)[0]
	if p.URL != "/de/cv" {
		t.Errorf("url = %q, want the pattern path /de/cv", p.URL)
	}
	if p.Referrer != "https://other.example.test/" {
		t.Errorf("referrer = %q, want the origin only", p.Referrer)
	}
}

// TestSlowUmamiNeverDelaysAResponse proves a client that holds every post
// does not hold a page.
func TestSlowUmamiNeverDelaysAResponse(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	fake := &fakeUmami{block: release}
	s := countingServer(t, fake, nil)

	done := make(chan int, 1)
	go func() { done <- visit(t, s.Handler(), "/de/", nil).Code }()
	select {
	case code := <-done:
		if code != http.StatusOK {
			t.Fatalf("status = %d, want 200", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the response waited for Umami")
	}
	close(release)
	closeServer(t, s)
}

// TestForwardedClientIsCountedOnlyBehindATrustedProxy proves the visitor
// behind the ingress is counted with their own masked address when the proxy
// is trusted, and not counted at all when it is not.
func TestForwardedClientIsCountedOnlyBehindATrustedProxy(t *testing.T) {
	t.Parallel()
	behindProxy := http.Header{"X-Forwarded-For": {"198.51.100.7"}}
	// visit sets the public peer address, so the proxy case builds its own.
	fromProxy := func(t *testing.T, h http.Handler) {
		t.Helper()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/de/", nil)
		req.RemoteAddr = "10.1.2.3:4000"
		req.Header.Set("User-Agent", visitorUA)
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		h.ServeHTTP(httptest.NewRecorder(), req)
	}

	t.Run("trusted", func(t *testing.T) {
		t.Parallel()
		fake := &fakeUmami{}
		s := countingServer(t, fake, map[string]string{"TRUSTED_PROXIES": "10.0.0.0/8"})
		fromProxy(t, s.Handler())
		closeServer(t, s)
		got := fake.payloads(t)
		if len(got) != 1 || got[0].IP != "198.51.100.0" {
			t.Fatalf("reported %+v, want one view with ip 198.51.100.0", got)
		}
	})

	t.Run("public peer spoofing X-Forwarded-For keeps its own address", func(t *testing.T) {
		t.Parallel()
		fake := &fakeUmami{}
		s := countingServer(t, fake, map[string]string{"TRUSTED_PROXIES": "10.0.0.0/8"})
		visit(t, s.Handler(), "/de/", behindProxy)
		closeServer(t, s)
		got := fake.payloads(t)
		if len(got) != 1 || got[0].IP != "203.0.113.0" {
			t.Fatalf("reported %+v, want one view with the peer's own /24 203.0.113.0", got)
		}
	})

	t.Run("untrusted", func(t *testing.T) {
		t.Parallel()
		fake := &fakeUmami{}
		s := countingServer(t, fake, nil)
		fromProxy(t, s.Handler())
		closeServer(t, s)
		if fake.count() != 0 {
			t.Fatalf("reported %d views, want nothing from a private peer", fake.count())
		}
	})
}

// TestHostnameHasNoPort proves the reported hostname is the host of SITE_URL
// without its port.
func TestHostnameHasNoPort(t *testing.T) {
	t.Parallel()
	fake := &fakeUmami{}
	s := countingServer(t, fake, map[string]string{"SITE_URL": "https://simon-ostendorf.test:8443"})
	visit(t, s.Handler(), "/de/", nil)
	closeServer(t, s)

	got := fake.payloads(t)
	if len(got) != 1 || got[0].Hostname != "simon-ostendorf.test" {
		t.Fatalf("reported %+v, want hostname simon-ostendorf.test without a port", got)
	}
}

// TestCountingIsOffWithoutConfiguration proves the site serves normally and
// says once that counting is off.
func TestCountingIsOffWithoutConfiguration(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	s := serverWithLogger(t, nil, slog.New(slog.NewTextHandler(&logs, nil)), options{})
	if rec := visit(t, s.Handler(), "/de/", nil); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if n := strings.Count(logs.String(), `msg="page view counting" enabled=false`); n != 1 {
		t.Errorf("off line logged %d times, want once: %q", n, logs.String())
	}
}

// requestLines returns the access log lines of a buffer.
func requestLines(logs *bytes.Buffer) []string {
	var out []string
	for line := range strings.SplitSeq(logs.String(), "\n") {
		if strings.Contains(line, "request handled") {
			out = append(out, line)
		}
	}
	return out
}

// TestAccessLogWritesOneLinePerRequestWithThePatternOnly proves, through the
// real handler, one line per request with the route pattern, never the raw
// path or query, no user agent and no referrer.
func TestAccessLogWritesOneLinePerRequestWithThePatternOnly(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	s := serverWithLogger(t, nil, slog.New(slog.NewTextHandler(&logs, nil)), options{})

	rec := visit(t, s.Handler(), "/de/cv?token=s3cr3t-token", http.Header{
		"Referer": {"https://other.example.test/s3cr3t-path"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	lines := requestLines(&logs)
	if len(lines) != 1 {
		t.Fatalf("got %d access log lines, want 1: %q", len(lines), logs.String())
	}
	line := lines[0]
	for _, want := range []string{"method=GET", "status=200", "bytes=", "duration_ms=", "client=203.0.113.7", "request_id="} {
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
	t.Parallel()
	env := map[string]string{"TRUSTED_PROXIES": "10.0.0.0/8"}

	t.Run("trusted proxy", func(t *testing.T) {
		t.Parallel()
		var logs bytes.Buffer
		s := serverWithLogger(t, env, slog.New(slog.NewTextHandler(&logs, nil)), options{})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/de/", nil)
		req.RemoteAddr = "10.1.2.3:4000"
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		s.Handler().ServeHTTP(httptest.NewRecorder(), req)
		lines := requestLines(&logs)
		if len(lines) != 1 || !strings.Contains(lines[0], "client=198.51.100.7") || strings.Contains(lines[0], "10.1.2.3") {
			t.Fatalf("lines = %q, want one with client=198.51.100.7", lines)
		}
	})

	t.Run("public peer spoofing the header", func(t *testing.T) {
		t.Parallel()
		var logs bytes.Buffer
		s := serverWithLogger(t, env, slog.New(slog.NewTextHandler(&logs, nil)), options{})
		visit(t, s.Handler(), "/de/", http.Header{"X-Forwarded-For": {"198.51.100.7"}})
		lines := requestLines(&logs)
		if len(lines) != 1 || !strings.Contains(lines[0], "client=203.0.113.7") || strings.Contains(lines[0], "198.51.100.7") {
			t.Fatalf("lines = %q, want one with the peer's own address", lines)
		}
	})
}

// TestAccessLogNeverLogsAnUnknownPath proves a 404 on a path carrying an id
// is logged without the path.
func TestAccessLogNeverLogsAnUnknownPath(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	s := serverWithLogger(t, nil, slog.New(slog.NewTextHandler(&logs, nil)), options{})
	visit(t, s.Handler(), "/de/reset/s3cr3t-token/", nil)
	lines := requestLines(&logs)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1: %q", len(lines), logs.String())
	}
	if strings.Contains(lines[0], "s3cr3t") || strings.Contains(lines[0], "reset") {
		t.Errorf("line leaks the path: %s", lines[0])
	}
}

// TestDevProxyOffByDefault proves /api/ is the 404 of production without
// FRONTEND_DEV_API_PROXY: nothing is forwarded.
func TestDevProxyOffByDefault(t *testing.T) {
	t.Parallel()
	transport := &recordingTransport{}
	h := server(t, map[string]string{"ALLOW_PLAIN_HTTP": "true"}, options{devAPITransport: transport}).Handler()
	rec := visit(t, h, "/api/v1/anything", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /api/v1/anything status = %d, want 404", rec.Code)
	}
	if transport.calls() != 0 {
		t.Errorf("the API was called %d times, want none", transport.calls())
	}
}

// TestDevProxyForwardsTheAPI proves the development switch forwards /api/ to
// the API container, so a compose stack has one browser origin.
func TestDevProxyForwardsTheAPI(t *testing.T) {
	t.Parallel()
	transport := &recordingTransport{}
	const upstream = "http://backend:8080"
	h := server(t, map[string]string{"FRONTEND_DEV_API_PROXY": upstream},
		options{devAPITransport: transport}).Handler()
	rec := visit(t, h, "/api/v1/anything", nil)
	if rec.Code != http.StatusOK || transport.calls() != 1 {
		t.Fatalf("GET /api/v1/anything = %d with %d calls, want 200 and one", rec.Code, transport.calls())
	}
	if got, want := transport.lastURL(), upstream+"/api/v1/anything"; got != want {
		t.Errorf("forwarded to %q, want %q", got, want)
	}
}

type recordingTransport struct {
	mu   sync.Mutex
	urls []string
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.urls = append(r.urls, req.URL.String())
	r.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{}`)), Request: req,
	}, nil
}

func (r *recordingTransport) calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.urls)
}

func (r *recordingTransport) lastURL() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.urls[len(r.urls)-1]
}
