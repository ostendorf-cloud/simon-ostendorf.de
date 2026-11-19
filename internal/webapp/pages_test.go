package webapp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testPageServer(t *testing.T) *pageServer {
	t.Helper()
	dist := fstest.MapFS{
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('app')")},
	}
	prerendered := fstest.MapFS{
		"de/index.html":         &fstest.MapFile{Data: []byte("<html><body>Startseite</body></html>")},
		"de/cv/index.html":      &fstest.MapFile{Data: []byte("<html><body>Lebenslauf</body></html>")},
		"de/contact/index.html": &fstest.MapFile{Data: []byte("<html><body>Kontakt</body></html>")},
		"de/imprint/index.html": &fstest.MapFile{Data: []byte("<html><body>Impressum</body></html>")},
		"de/privacy/index.html": &fstest.MapFile{Data: []byte("<html><body>Datenschutz</body></html>")},
		"de/404.html":           &fstest.MapFile{Data: []byte("<html><body>Nicht gefunden</body></html>")},
		"en/index.html":         &fstest.MapFile{Data: []byte("<html><body>Home</body></html>")},
		"en/cv/index.html":      &fstest.MapFile{Data: []byte("<html><body>CV</body></html>")},
		"en/contact/index.html": &fstest.MapFile{Data: []byte("<html><body>Contact</body></html>")},
		"en/imprint/index.html": &fstest.MapFile{Data: []byte("<html><body>Imprint</body></html>")},
		"en/privacy/index.html": &fstest.MapFile{Data: []byte("<html><body>Privacy</body></html>")},
		"en/404.html":           &fstest.MapFile{Data: []byte("<html><body>Not found</body></html>")},
	}
	p, err := newPageServer(dist, prerendered)
	if err != nil {
		t.Fatalf("newPageServer() error = %v", err)
	}
	return p
}

// testMux builds a real http.ServeMux the same way production code does:
// one pattern per locale and route slug via registerRoutes, plus the
// catch-all pageHandler under "/". Tests dispatch through it rather than
// calling pageHandler or routeHandler directly, so they observe the
// actual matched pattern and the actual route interception a bare
// handler call cannot show.
func testMux(t *testing.T, p *pageServer) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	identity := func(h http.Handler) http.Handler { return h }
	registerRoutes(mux, p, identity)
	mux.Handle("/", p.pageHandler())
	return mux
}

func TestPageHandlerServesKnownRoutes(t *testing.T) {
	p := testPageServer(t)
	mux := testMux(t, p)

	tests := map[string]struct {
		path        string
		wantStatus  int
		wantBody    string
		wantPattern string
	}{
		"german home":     {"/de/", http.StatusOK, "Startseite", "GET /de/{$}"},
		"german cv":       {"/de/cv/", http.StatusOK, "Lebenslauf", "GET /de/cv/{$}"},
		"german contact":  {"/de/contact/", http.StatusOK, "Kontakt", "GET /de/contact/{$}"},
		"german imprint":  {"/de/imprint", http.StatusOK, "Impressum", "GET /de/imprint"},
		"german privacy":  {"/de/privacy", http.StatusOK, "Datenschutz", "GET /de/privacy"},
		"english home":    {"/en/", http.StatusOK, "Home", "GET /en/{$}"},
		"english cv":      {"/en/cv/", http.StatusOK, "CV", "GET /en/cv/{$}"},
		"english contact": {"/en/contact/", http.StatusOK, "Contact", "GET /en/contact/{$}"},
		"english imprint": {"/en/imprint", http.StatusOK, "Imprint", "GET /en/imprint"},
		"english privacy": {"/en/privacy", http.StatusOK, "Privacy", "GET /en/privacy"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			body, _ := io.ReadAll(rec.Body)
			if !strings.Contains(string(body), tc.wantBody) {
				t.Errorf("body = %q, want it to contain %q", body, tc.wantBody)
			}
			if req.Pattern != tc.wantPattern {
				t.Errorf("matched pattern = %q, want %q", req.Pattern, tc.wantPattern)
			}
		})
	}
}

func TestPageHandlerUnknownRouteThroughMuxServesLocalNotFound(t *testing.T) {
	p := testPageServer(t)
	mux := testMux(t, p)

	tests := map[string]struct {
		path     string
		wantBody string
	}{
		"unknown german route":  {"/de/portfolio/", "Nicht gefunden"},
		"unknown english route": {"/en/portfolio/", "Not found"},
		"unknown locale":        {"/fr/", "Nicht gefunden"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
			body, _ := io.ReadAll(rec.Body)
			if !strings.Contains(string(body), tc.wantBody) {
				t.Errorf("body = %q, want it to contain %q", body, tc.wantBody)
			}
			if req.Pattern != "/" {
				t.Errorf("matched pattern = %q, want the catch-all %q", req.Pattern, "/")
			}
		})
	}
}

func TestPageHandlerRedirectsRoot(t *testing.T) {
	tests := map[string]struct {
		acceptLanguage string
		wantLocation   string
	}{
		"no header defaults to german":    {"", "/de/"},
		"german preference":               {"de-DE,de;q=0.9", "/de/"},
		"english preference":              {"en-US,en;q=0.9", "/en/"},
		"english ranked above german":     {"de;q=0.5,en;q=0.9", "/en/"},
		"unsupported language falls back": {"fr-FR,fr;q=0.9", "/de/"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p := testPageServer(t)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			if tc.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tc.acceptLanguage)
			}
			rec := httptest.NewRecorder()
			p.pageHandler().ServeHTTP(rec, req)

			if rec.Code != http.StatusFound {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
			}
			if got := rec.Header().Get("Location"); got != tc.wantLocation {
				t.Errorf("Location = %q, want %q", got, tc.wantLocation)
			}
		})
	}
}

func TestPageHandlerCanonicalizesTrailingSlash(t *testing.T) {
	p := testPageServer(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/de/cv", nil)
	rec := httptest.NewRecorder()
	p.pageHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMovedPermanently)
	}
	if got := rec.Header().Get("Location"); got != "/de/cv/" {
		t.Errorf("Location = %q, want /de/cv/", got)
	}
}

func TestPageHandlerCanonicalizesTrailingSlashPreservesQuery(t *testing.T) {
	p := testPageServer(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/de/cv?utm_source=x", nil)
	rec := httptest.NewRecorder()
	p.pageHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMovedPermanently)
	}
	if got := rec.Header().Get("Location"); got != "/de/cv/?utm_source=x" {
		t.Errorf("Location = %q, want /de/cv/?utm_source=x", got)
	}
}

// TestPageHandlerUnknownPathAnswersNotFoundWithoutRedirecting guards
// against a redirect issued before a path is known to be a route this
// site serves. ostendorf.dev's own page handler has exactly that bug: it
// redirects first and only checks the route afterwards, so an admin only
// path like /metrics answers 301 on its public listener too. This site's
// locale check runs first, so the same request here must answer 404 with
// no Location header at all, never a redirect to anywhere.
func TestPageHandlerUnknownPathAnswersNotFoundWithoutRedirecting(t *testing.T) {
	p := testPageServer(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	p.pageHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Header().Get("Location"); got != "" {
		t.Errorf("Location = %q, want no redirect at all", got)
	}
}

func TestPageHandlerUnknownRouteServesLocalNotFound(t *testing.T) {
	p := testPageServer(t)

	tests := map[string]struct {
		path     string
		wantBody string
	}{
		"unknown german route":  {"/de/portfolio/", "Nicht gefunden"},
		"unknown english route": {"/en/portfolio/", "Not found"},
		"unknown locale":        {"/fr/", "Nicht gefunden"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			p.pageHandler().ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
			body, _ := io.ReadAll(rec.Body)
			if !strings.Contains(string(body), tc.wantBody) {
				t.Errorf("body = %q, want it to contain %q", body, tc.wantBody)
			}
		})
	}
}

func TestPageHandlerRejectsUnsafeMethods(t *testing.T) {
	p := testPageServer(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/de/", nil)
	rec := httptest.NewRecorder()
	p.pageHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q, want %q", got, "GET, HEAD")
	}
}

func TestAssetHandlerServesAndCachesImmutably(t *testing.T) {
	p := testPageServer(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()
	p.assetHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control = %q, want immutable", got)
	}
}

func TestSplitLocale(t *testing.T) {
	tests := map[string]struct {
		path       string
		wantLocale string
		wantRest   string
		wantOK     bool
	}{
		"bare locale":        {"/de", "de", "", true},
		"locale with slash":  {"/de/", "de", "", true},
		"locale with route":  {"/en/imprint/", "en", "imprint/", true},
		"unknown locale":     {"/fr/", "", "", false},
		"root":               {"/", "", "", false},
		"locale-like prefix": {"/development/", "", "", false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			locale, rest, ok := splitLocale(tc.path)
			if ok != tc.wantOK || locale != tc.wantLocale || rest != tc.wantRest {
				t.Errorf("splitLocale(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tc.path, locale, rest, ok, tc.wantLocale, tc.wantRest, tc.wantOK)
			}
		})
	}
}

func TestPreferredLocale(t *testing.T) {
	tests := map[string]struct {
		header string
		want   string
	}{
		"empty header":                {"", "de"},
		"german only":                 {"de", "de"},
		"english only":                {"en", "en"},
		"english with region":         {"en-US", "en"},
		"german with region":          {"de-AT", "de"},
		"german ranked above english": {"en;q=0.5,de;q=0.9", "de"},
		"unsupported falls back":      {"fr-FR", "de"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := preferredLocale(tc.header); got != tc.want {
				t.Errorf("preferredLocale(%q) = %q, want %q", tc.header, got, tc.want)
			}
		})
	}
}
