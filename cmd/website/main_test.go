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

	if _, err := fs.Stat(dist, "index.html"); err != nil {
		t.Fatalf("embedded dist has no index.html: %v", err)
	}
}
