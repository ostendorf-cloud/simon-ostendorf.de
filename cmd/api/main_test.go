package main

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestAPIEmbedsNoFrontend proves the API binary carries no frontend build:
// no file of this package imports embed or holds a go:embed directive. The
// pages are cmd/frontend's, in an image of their own.
func TestAPIEmbedsNoFrontend(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range file.Imports {
			if imp.Path.Value == `"embed"` {
				t.Errorf("%s imports embed", e.Name())
			}
		}
		for _, group := range file.Comments {
			for _, c := range group.List {
				if strings.HasPrefix(c.Text, "//go:embed") {
					t.Errorf("%s holds a go:embed directive", e.Name())
				}
			}
		}
	}
}
