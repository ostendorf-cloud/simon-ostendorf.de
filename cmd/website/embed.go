package main

import "embed"

// dist is the built SPA, written by "task web:build" (vite build, with
// outDir pointed at this directory). embed resolves relative to this file,
// which is why the directory sits under cmd/ rather than internal/.
//
//go:embed dist
var distFS embed.FS

// prerendered is the static HTML for every route of every locale, written
// by "task prerender" after dist exists.
//
//go:embed prerendered
var prerenderedFS embed.FS
