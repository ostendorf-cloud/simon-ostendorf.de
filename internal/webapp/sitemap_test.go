package webapp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRobotsHandlerAllowsEverythingAndNamesTheSitemap(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/robots.txt", nil)
	rec := httptest.NewRecorder()
	app.robotsHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Allow: /") {
		t.Errorf("body = %q, want it to allow everything", body)
	}
	if !strings.Contains(body, "Sitemap: https://simon-ostendorf.de/sitemap.xml") {
		t.Errorf("body = %q, want it to name the sitemap", body)
	}
}

func TestSitemapHandlerListsEveryRouteAndLocaleWithAlternates(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sitemap.xml", nil)
	rec := httptest.NewRecorder()
	app.sitemapHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()

	wantURLCount := len(routeFiles) * len(locales)
	if got := strings.Count(body, "<url>"); got != wantURLCount {
		t.Errorf("sitemap carries %d <url> entries, want %d (one per route per locale)", got, wantURLCount)
	}

	for _, want := range []string{
		"<loc>https://simon-ostendorf.de/de/</loc>",
		"<loc>https://simon-ostendorf.de/en/</loc>",
		"<loc>https://simon-ostendorf.de/de/imprint</loc>",
		"<loc>https://simon-ostendorf.de/en/imprint</loc>",
		"<loc>https://simon-ostendorf.de/de/cv/</loc>",
		`hreflang="de"`,
		`hreflang="en"`,
		`hreflang="x-default"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("sitemap body does not contain %q:\n%s", want, body)
		}
	}
}

func TestRouteURL(t *testing.T) {
	tests := map[string]struct {
		slug string
		want string
	}{
		"home has no slug":                   {"", "https://simon-ostendorf.de/de/"},
		"a page has a slug":                  {"cv", "https://simon-ostendorf.de/de/cv/"},
		"a legal page has no trailing slash": {"imprint", "https://simon-ostendorf.de/de/imprint"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := routeURL("https://simon-ostendorf.de", "de", tc.slug); got != tc.want {
				t.Errorf("routeURL(..., %q) = %q, want %q", tc.slug, got, tc.want)
			}
		})
	}
}
