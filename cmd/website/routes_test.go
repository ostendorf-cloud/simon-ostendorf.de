package main

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/routes"
)

// TestPrerenderedOutputMatchesRouteTable is the check that proves
// internal/webapp's page handler, sitemap and mux registration, and
// web/scripts/prerender.ts's static renderer, cannot drift apart any
// more. Both now read the same embedded route table (see
// internal/routes), but reading the same file does not by itself prove
// prerender.ts still writes what that table says: this test checks the
// real, generated prerendered output "task prerender" wrote against that
// same table, catching exactly the two failure modes a table restated by
// hand on either side would allow silently:
//
//   - a route the table lists with no prerendered file for it under some
//     locale (the "Go only" case: a route would register and the sitemap
//     would advertise it, but a crawler following that link would hit a
//     404);
//   - a prerendered file that answers no route the table lists (the
//     "TypeScript only" case: a page prerender.ts would write under its
//     own pattern, with nothing in internal/webapp ever serving it
//     there).
func TestPrerenderedOutputMatchesRouteTable(t *testing.T) {
	table, err := routes.Load()
	if err != nil {
		t.Fatalf("routes.Load() error = %v", err)
	}
	wantFiles := make(map[string]bool, len(table))
	for _, r := range table {
		wantFiles[r.File] = true
	}

	sub, err := fs.Sub(prerenderedFS, "prerendered")
	if err != nil {
		t.Fatalf(`fs.Sub(prerenderedFS, "prerendered") error = %v`, err)
	}

	localeDirs, err := fs.ReadDir(sub, ".")
	if err != nil {
		t.Fatalf("fs.ReadDir(prerendered) error = %v", err)
	}
	if len(localeDirs) == 0 {
		t.Fatal("prerendered has no locale directories; run task prerender first")
	}

	for _, entry := range localeDirs {
		if !entry.IsDir() {
			continue
		}
		locale := entry.Name()

		for _, r := range table {
			want := locale + "/" + r.File
			if _, err := fs.Stat(sub, want); err != nil {
				t.Errorf("locale %q: route %q (slug %q) has no prerendered file at %q: %v",
					locale, r.Name, r.Slug, want, err)
			}
		}

		for file := range foundFiles(t, sub, locale) {
			if file == "404.html" {
				// Not a route: the locale's own fallback page, served
				// directly by serveNotFound rather than through the
				// route table.
				continue
			}
			if !wantFiles[file] {
				t.Errorf("locale %q: prerendered file %q answers no route in routes.json", locale, file)
			}
		}
	}
}

// foundFiles lists every file actually prerendered under locale,
// relative to that locale's own directory.
func foundFiles(t *testing.T, sub fs.FS, locale string) map[string]bool {
	t.Helper()
	found := make(map[string]bool)
	err := fs.WalkDir(sub, locale, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		found[strings.TrimPrefix(p, locale+"/")] = true
		return nil
	})
	if err != nil {
		t.Fatalf("fs.WalkDir(%q) error = %v", locale, err)
	}
	return found
}
