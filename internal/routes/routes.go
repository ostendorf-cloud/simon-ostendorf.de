// Package routes is the website's one route table. internal/webapp's page
// handler, sitemap and per-page mux registration, web/scripts/prerender.ts's
// static renderer, and the frontend's own router (web/src/router/index.ts)
// and SEO metadata (web/src/lib/seo.ts), through the shared
// web/src/lib/routeTable.ts, each build their own list from routes.json,
// embedded here, instead of restating it: a page added to one side and not
// the other is no longer possible, because there is only one list left to
// edit. See AGENTS.md for how each side reads it, and cmd/website's own test
// for the check that the prerendered output this table describes actually
// exists.
package routes

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed routes.json
var tableJSON []byte

// Route is one page the site serves under every locale.
type Route struct {
	// Name identifies the route for a caller that needs a stable key
	// distinct from Slug, such as the frontend router's route name or its
	// SEO metadata lookup: Slug alone cannot serve that purpose, because
	// the home page's slug is empty.
	Name string `json:"name"`
	// Slug is the path segment after the locale; the home page's slug is
	// empty.
	Slug string `json:"slug"`
	// File is the prerendered file that answers Slug, relative to one
	// locale's own directory under cmd/website/prerendered.
	File string `json:"file"`
}

// Load parses the embedded route table.
func Load() ([]Route, error) {
	return Parse(tableJSON)
}

// Parse decodes data into the route table, failing on anything that would
// make the table ambiguous for any of its five consumers (the mux
// registration, the sitemap, the prerenderer, the Vue router and the SEO
// metadata; see the package comment): an entry missing a name or a file,
// or two entries sharing a name, a slug or a file.
//
//   - Name is required and must be unique: it is the key both
//     web/src/router/index.ts (the route name) and web/src/lib/seo.ts (the
//     slugs and metadata lookups) use. A repeated name corrupts seo.ts's
//     name-keyed map silently (the later entry wins) and makes vue-router
//     drop the earlier of the two routes outright.
//   - File is required and must be unique: two entries sharing one would
//     make web/scripts/prerender.ts's render loop overwrite one route's
//     prerendered output with the other's, the same "later entry wins"
//     failure as a duplicate name, just inside the generated HTML on disk
//     instead of a map.
//   - Slug must be unique, but may be empty: the home page's own
//     convention. An empty slug is rejected only once a second entry
//     repeats it, exactly like any other duplicate value; emptiness
//     itself is not the problem.
//
// A caller that only ever parses the checked-in routes.json should treat a
// failure here as a build bug, not a runtime one.
//
// internal/routes/testdata holds the fixtures this file's own tests check
// against; web/src/lib/routeTable.ts's validateRouteTable enforces the
// same three rules for the frontend, which reads routes.json directly
// through Vite's JSON import rather than through this function, and its
// own test reads the exact same testdata fixtures, so the two validators
// are proven to agree on every case either side has written down.
func Parse(data []byte) ([]Route, error) {
	var table []Route
	if err := json.Unmarshal(data, &table); err != nil {
		return nil, fmt.Errorf("routes: parse route table: %w", err)
	}

	names := make(map[string]bool, len(table))
	slugs := make(map[string]bool, len(table))
	files := make(map[string]bool, len(table))
	for _, r := range table {
		if r.Name == "" {
			return nil, fmt.Errorf("routes: route table: entry with slug %q has no name", r.Slug)
		}
		if r.File == "" {
			return nil, fmt.Errorf("routes: route table: route %q has no file", r.Name)
		}
		if names[r.Name] {
			return nil, fmt.Errorf("routes: route table: duplicate name %q", r.Name)
		}
		names[r.Name] = true
		if slugs[r.Slug] {
			return nil, fmt.Errorf("routes: route table: duplicate slug %q", r.Slug)
		}
		slugs[r.Slug] = true
		if files[r.File] {
			return nil, fmt.Errorf("routes: route table: duplicate file %q", r.File)
		}
		files[r.File] = true
	}
	return table, nil
}
