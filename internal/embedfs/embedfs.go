// Package embedfs guards the directory lookup cmd/frontend runs against
// its go:embed filesystem, so a missing build fails at startup.
package embedfs

import (
	"fmt"
	"io/fs"
)

// Subtree returns the subtree of fsys rooted at dir, failing when dir
// does not exist.
//
// fs.Sub alone would not catch a missing dir: it builds a filesystem
// rooted there lazily, with no check that it exists, for every fs.FS
// that does not implement fs.SubFS itself, embed.FS included. It
// returns no error at all for a directory that is not there and only
// fails on the first read against the result, which for an embedded
// build means the first request for a page or an asset, far from where
// the real problem is. fs.Stat against the original filesystem is what
// actually does the validating.
func Subtree(fsys fs.FS, dir string) (fs.FS, error) {
	if _, err := fs.Stat(fsys, dir); err != nil {
		return nil, fmt.Errorf("embedded %s: %w", dir, err)
	}
	return fs.Sub(fsys, dir)
}
