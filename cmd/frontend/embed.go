package main

import "embed"

// distFS is the frontend build: Vite's dist, written by "task web:build",
// with the prerendered pages, the 404 page per language and sitemap.xml
// written into it by "task prerender". It is govue's frontend build layout.
// embed resolves relative to this file, which is why the directory sits
// under cmd/ rather than internal/. Vite's .vite manifest directory starts
// with a dot and is not embedded.
//
//go:embed dist
var distFS embed.FS
