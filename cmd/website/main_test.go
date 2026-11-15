package main

import (
	"io/fs"
	"testing"
)

// TestEmbeddedDistBuildsAWorkingHandler proves the embedded dist filesystem is
// shaped the way main's fs.Sub call expects, without starting a server or
// binding a port: a build with an empty or mis-shaped frontend build fails
// here instead of at a request.
func TestEmbeddedDistBuildsAWorkingHandler(t *testing.T) {
	dist, err := fs.Sub(distFS, "dist")
	if err != nil {
		t.Fatalf(`fs.Sub(distFS, "dist") error = %v`, err)
	}

	if _, err := fs.Stat(dist, "assets"); err != nil {
		t.Fatalf("embedded dist has no assets directory: %v", err)
	}
}

// TestEmbeddedPrerenderedBuildsAWorkingHandler is prerenderedFS' own
// equivalent of the test above: a build with an empty or mis-shaped
// "task prerender" output fails here instead of at a request.
func TestEmbeddedPrerenderedBuildsAWorkingHandler(t *testing.T) {
	prerendered, err := fs.Sub(prerenderedFS, "prerendered")
	if err != nil {
		t.Fatalf(`fs.Sub(prerenderedFS, "prerendered") error = %v`, err)
	}

	if _, err := fs.Stat(prerendered, "de/index.html"); err != nil {
		t.Fatalf("embedded prerendered has no de/index.html: %v", err)
	}
}
