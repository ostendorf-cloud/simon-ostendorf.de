package main

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/embedfs"
	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/routes"
)

// build is the real embedded build: the output "task web:build" and "task
// prerender" wrote into dist. This package does not compile without a dist
// directory, since go:embed refuses an empty pattern, so every test here
// runs against the real build and none skips.
func build(t *testing.T) fs.FS {
	t.Helper()
	b, err := embedfs.Subtree(distFS, "dist")
	if err != nil {
		t.Fatalf("embedded build: %v; run task web:build and task prerender first", err)
	}
	return b
}

// TestPrerenderedOutputMatchesRouteTable checks the real prerender output
// against internal/routes' table, which web/scripts/prerender.ts reads as
// well. It catches the two ways the sides could drift apart:
//
//   - a route the table lists with no prerendered page under some language:
//     a crawler following the sitemap would hit a 404;
//   - a prerendered page that answers no route of the table: a page
//     prerender.ts writes on its own pattern, which nothing mounts.
func TestPrerenderedOutputMatchesRouteTable(t *testing.T) {
	table, err := routes.Load()
	if err != nil {
		t.Fatalf("routes.Load() error = %v", err)
	}
	want := make(map[string]bool, len(table))
	for _, r := range table {
		want[r.File] = true
	}

	b := build(t)
	for _, code := range languages {
		for _, r := range table {
			if _, err := fs.Stat(b, code+"/"+r.File); err != nil {
				t.Errorf("language %q: route %q (slug %q) has no page %s/%s: %v",
					code, r.Name, r.Slug, code, r.File, err)
			}
		}

		n := 0
		err := fs.WalkDir(b, code, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			n++
			file := strings.TrimPrefix(p, code+"/")
			// The language's own fallback page is no route.
			if file != "404.html" && !want[file] {
				t.Errorf("language %q: page %q answers no route in routes.json", code, file)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("fs.WalkDir(%q) error = %v", code, err)
		}
		if n == 0 {
			t.Errorf("language %q has no prerendered pages; run task prerender first", code)
		}
	}
}

// TestRealBuildStartsTheFrontend feeds the real prerender output to govue's
// frontend.New: an incomplete build, a missing 404 page or legal page, or a
// bad favicon file would make it refuse to start.
func TestRealBuildStartsTheFrontend(t *testing.T) {
	server(t, nil, options{})
}
