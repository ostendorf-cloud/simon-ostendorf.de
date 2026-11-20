// Command content is the build step that fetches the pinned cv release and
// writes the CV page's dataset and PDFs. It runs in CI and on a developer
// machine, never in the deployed service.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ostendorf-cloud/simon-ostendorf.de/internal/content"
)

const defaultOut = "web/src/generated"

func main() {
	os.Exit(realMain())
}

func realMain() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Getenv, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "content:", err)
		return 1
	}
	return 0
}

// run is main without the process around it, so a test can drive it.
//
// Environment: CV_RELEASE_TAG is the pinned cv release and the
// CV_RELEASE_SHA256_* variables are the digests of its four assets, all from
// cv-release.env. SOURCE_DATE_EPOCH fixes the clock for a reproducible build.
func run(ctx context.Context, args []string, getenv func(string) string, stderr io.Writer) error {
	fs := flag.NewFlagSet("content", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", defaultOut, "directory the dataset and the PDFs are written to")
	assetDir := fs.String("asset-dir", "", "read the four assets from this directory instead of downloading the release")
	base := fs.String("cv-base", content.DefaultReleaseBase, "base URL release assets are downloaded from")
	if err := fs.Parse(args); err != nil {
		return err
	}

	now := time.Now
	if v := getenv("SOURCE_DATE_EPOCH"); v != "" {
		secs, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("SOURCE_DATE_EPOCH: %w", err)
		}
		now = func() time.Time { return time.Unix(secs, 0) }
	}

	local := *assetDir != ""
	tag := getenv("CV_RELEASE_TAG")
	digests := map[string]string{}
	for _, a := range content.Assets() {
		digests[a.Env] = getenv(a.Env)
	}

	var assets map[string][]byte
	var err error
	if local {
		tag = "local"
		assets, err = content.ReadAssets(*assetDir)
	} else {
		assets, err = content.DownloadAssets(ctx, nil, *base, tag)
	}
	if err != nil {
		return err
	}
	res, err := content.Generate(content.Options{
		Tag: tag, Digests: digests, Local: local, Assets: assets, Now: now,
	})
	if err != nil {
		return err
	}
	return content.WriteOutput(*out, res)
}
