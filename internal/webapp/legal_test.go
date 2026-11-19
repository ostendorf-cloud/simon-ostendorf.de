package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// legalApp is an App whose prerendered output holds every route's page, so a
// redirect can be followed to a page that answers 200.
func legalApp(t *testing.T) *App {
	t.Helper()
	return countingApp(t, nil, nil, nil)
}

func legalGet(t *testing.T, app *App, method, target string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), method, target, nil)
	for k, v := range header {
		req.Header[k] = v
	}
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	return rec
}

// TestLegalPathsRedirectToTheLocalePage pins the govue redirect set through
// the real handler chain: the short paths follow the visitor, the German
// aliases always go to German, and every target answers 200.
func TestLegalPathsRedirectToTheLocalePage(t *testing.T) {
	app := legalApp(t)

	for name, tc := range map[string]struct {
		target string
		header http.Header
		want   string
	}{
		"imprint without a header":     {"/imprint", nil, "/de/imprint"},
		"imprint english browser":      {"/imprint", http.Header{"Accept-Language": {"en-US,en;q=0.9"}}, "/en/imprint"},
		"imprint trailing slash":       {"/imprint/", nil, "/de/imprint"},
		"privacy english browser":      {"/privacy", http.Header{"Accept-Language": {"en"}}, "/en/privacy"},
		"privacy german browser":       {"/privacy", http.Header{"Accept-Language": {"de"}}, "/de/privacy"},
		"impressum is always german":   {"/impressum", http.Header{"Accept-Language": {"en"}}, "/de/imprint"},
		"datenschutz is always german": {"/datenschutz", http.Header{"Accept-Language": {"en"}}, "/de/privacy"},
		"datenschutzerklaerung":        {"/datenschutzerklaerung", http.Header{"Accept-Language": {"en"}}, "/de/privacy"},
		"datenschutzerklaerung slash":  {"/datenschutzerklaerung/", nil, "/de/privacy"},
		"query is dropped":             {"/impressum?next=//evil.example", nil, "/de/imprint"},
	} {
		t.Run(name, func(t *testing.T) {
			rec := legalGet(t, app, http.MethodGet, tc.target, tc.header)
			if rec.Code != http.StatusFound {
				t.Fatalf("status = %d, want 302", rec.Code)
			}
			if got := rec.Header().Get("Location"); got != tc.want {
				t.Fatalf("Location = %q, want %q", got, tc.want)
			}
			if got := rec.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			if page := legalGet(t, app, http.MethodGet, tc.want, nil); page.Code != http.StatusOK {
				t.Errorf("the redirect target %s answers %d, want 200", tc.want, page.Code)
			}
		})
	}
}

// TestSiteWithoutTermsHasNoTermsRoutes pins that this site has no terms page:
// /terms and /agb are not registered and fall to the 404 page.
func TestSiteWithoutTermsHasNoTermsRoutes(t *testing.T) {
	app := legalApp(t)
	for _, target := range []string{"/terms", "/agb", "/de/terms/", "/en/terms/"} {
		if rec := legalGet(t, app, http.MethodGet, target, nil); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", target, rec.Code)
		}
	}
}

// TestLegalPagesAreServedWithoutATrailingSlash pins the canonical path govue
// links and checks, and the redirect from the slash form.
func TestLegalPagesAreServedWithoutATrailingSlash(t *testing.T) {
	app := legalApp(t)
	for _, target := range []string{"/de/imprint", "/en/imprint", "/de/privacy", "/en/privacy"} {
		if rec := legalGet(t, app, http.MethodGet, target, nil); rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", target, rec.Code)
		}
		rec := legalGet(t, app, http.MethodGet, target+"/", nil)
		if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != target {
			t.Errorf("GET %s/ = %d to %q, want 301 to %s", target, rec.Code, rec.Header().Get("Location"), target)
		}
		// A bare 301 is cached forever, and this one reverses the old direction.
		if got := rec.Header().Get("Cache-Control"); got != "public, max-age=3600" {
			t.Errorf("GET %s/ Cache-Control = %q, want public, max-age=3600", target, got)
		}
	}
}

func TestLegalPathsRefuseOtherMethods(t *testing.T) {
	app := legalApp(t)
	rec := legalGet(t, app, http.MethodPost, "/impressum", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /impressum = %d, want 405", rec.Code)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q, want GET, HEAD", got)
	}
}
