package main

import (
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/config"
	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/webapp"
)

var pdfLink = regexp.MustCompile(`<a[^>]*href="(/assets/cv\.(de|en)-[^"]+\.pdf)"[^>]*>`)

// TestPrerenderedCvPageOffersServedPDFs proves, over the real build output,
// that the CV page of each locale is static HTML with its content and a PDF
// link per language, and that every linked PDF is served by the real handler
// as a PDF. The page needs no JavaScript to show the CV or to reach the files.
func TestPrerenderedCvPageOffersServedPDFs(t *testing.T) {
	dist, err := fs.Sub(distFS, "dist")
	if err != nil {
		t.Fatal(err)
	}
	prerendered, err := fs.Sub(prerenderedFS, "prerendered")
	if err != nil {
		t.Fatal(err)
	}
	siteURL, err := url.Parse(config.DefaultSiteURL)
	if err != nil {
		t.Fatal(err)
	}
	app, err := webapp.New(config.Config{SiteURL: siteURL}, slog.New(slog.DiscardHandler), dist, prerendered)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = app.Close() }()
	handler := app.Handler()

	for _, locale := range []string{"de", "en"} {
		t.Run(locale, func(t *testing.T) {
			page := get(t, handler, "/"+locale+"/cv/")
			if page.Code != http.StatusOK {
				t.Fatalf("status = %d", page.Code)
			}
			html, _ := io.ReadAll(page.Body)
			if !strings.Contains(string(html), "<h2") {
				t.Fatal("the CV page carries no section in its static HTML")
			}
			links := pdfLink.FindAllStringSubmatch(string(html), -1)
			if len(links) != 2 {
				t.Fatalf("PDF links = %d, want one per language", len(links))
			}
			seen := map[string]bool{}
			for _, l := range links {
				seen[l[2]] = true
				rec := get(t, handler, l[1])
				if rec.Code != http.StatusOK {
					t.Fatalf("%s: status = %d", l[1], rec.Code)
				}
				body, _ := io.ReadAll(rec.Body)
				if !strings.HasPrefix(string(body), "%PDF-") {
					t.Fatalf("%s is not a PDF", l[1])
				}
			}
			if !seen["de"] || !seen["en"] {
				t.Fatalf("languages linked: %v", seen)
			}
		})
	}
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
	return rec
}
