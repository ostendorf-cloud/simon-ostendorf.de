package embedfs

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// TestSubtreeFailsWhenTheDirectoryIsMissing reproduces the defect fs.Sub
// itself cannot catch: it builds a filesystem rooted at a given directory
// lazily, with no check that the directory exists, for every fs.FS that
// does not implement fs.SubFS itself (embed.FS and fstest.MapFS included).
// Subtree exists to catch this at the call site instead.
func TestSubtreeFailsWhenTheDirectoryIsMissing(t *testing.T) {
	t.Parallel()

	tests := map[string]fs.FS{
		"another directory only": fstest.MapFS{
			"other/index.html": &fstest.MapFile{Data: []byte("<html></html>")},
		},
		"fsys entirely empty": fstest.MapFS{},
	}

	for name, fsys := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := Subtree(fsys, "dist")
			if !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("Subtree() error = %v, want fs.ErrNotExist", err)
			}
			if !strings.Contains(err.Error(), "dist") {
				t.Errorf("error = %q, want it to name the directory", err.Error())
			}
		})
	}
}

// TestSubtreeSucceedsForAnExistingDirectory proves the check adds no false
// positive: a directory that is there resolves to a usable subtree.
func TestSubtreeSucceedsForAnExistingDirectory(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"dist/assets/app.js": &fstest.MapFile{Data: []byte("console.log(1)")},
	}
	sub, err := Subtree(fsys, "dist")
	if err != nil {
		t.Fatalf("Subtree() error = %v, want nil", err)
	}
	if _, err := fs.Stat(sub, "assets/app.js"); err != nil {
		t.Errorf("fs.Stat(sub, \"assets/app.js\") error = %v, want nil", err)
	}
}
