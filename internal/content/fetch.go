package content

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// DefaultReleaseBase is where GitHub serves release assets from.
const DefaultReleaseBase = "https://github.com"

// CVRepository is the repository whose release carries the assets.
const CVRepository = "ostendorf-cloud/cv"

// Size caps for one download. The real JSON is a few kilobytes.
const (
	maxJSON = 8 << 20
	maxPDF  = 32 << 20
)

// maxRedirects is how many redirects the client follows.
const maxRedirects = 5

// ReleaseRedirectHosts are the hosts a release asset download may be redirected
// to, besides the host the download started at. GitHub answers a release asset
// request with a redirect to its asset CDN. These names were not verified live
// for this change: if GitHub moves the CDN, the build fails with a refused
// redirect naming the host, and this list is the one place to extend after
// checking it.
var ReleaseRedirectHosts = []string{
	"github.com",
	"objects.githubusercontent.com",
	"release-assets.githubusercontent.com",
}

var errRedirectRefused = errors.New("redirect refused")

// Asset is one file of the pinned release this site reads.
type Asset struct {
	// Name is the release asset name, for example cv.de.json.
	Name string
	// Env is the cv-release.env variable holding its sha256.
	Env string
	// Lang is the language the asset belongs to.
	Lang string
	// PDF is true for a PDF, false for the JSON artefact.
	PDF bool
}

// Assets lists every asset of a release: the JSON artefact and the PDF of each
// language.
func Assets() []Asset {
	var out []Asset
	for _, lang := range Languages {
		up := strings.ToUpper(lang)
		out = append(out,
			Asset{Name: ArtefactFileName(lang), Env: "CV_RELEASE_SHA256_" + up, Lang: lang},
			Asset{Name: PDFFileName(lang), Env: "CV_RELEASE_SHA256_PDF_" + up, Lang: lang, PDF: true},
		)
	}
	return out
}

// tagPattern accepts a semantic version tag only. A moving reference such as
// latest, main or a branch name does not match, which keeps the pin a pin, and
// the pattern keeps the tag safe to put into a URL path.
var tagPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$`)

// ValidateTag reports whether tag names a release rather than a moving reference.
func ValidateTag(tag string) error {
	if tag == "" {
		return errors.New("no cv release is pinned: set CV_RELEASE_TAG in cv-release.env to a release tag such as v0.1.0")
	}
	if !tagPattern.MatchString(tag) {
		return fmt.Errorf("cv release tag %q is not a version tag such as v0.1.0: a moving reference is not a pin", tag)
	}
	return nil
}

// CheckDigest verifies that raw, the content of asset, has the sha256 pinned
// in cv-release.env. A release tag is not immutable, so this pin is what makes
// a moved tag or a replaced asset fail the build instead of changing it
// silently. An empty pin is an error too.
func CheckDigest(a Asset, tag string, raw []byte, want string) error {
	if want == "" {
		return fmt.Errorf("cv-release.env: %s is empty: with %s pinned, set it to the sha256 of %s (sha256sum)", a.Env, tag, a.Name)
	}
	wantLower := strings.ToLower(want)
	if b, err := hex.DecodeString(wantLower); err != nil || len(b) != sha256.Size {
		return fmt.Errorf("cv-release.env: %s is not a sha256, 64 hexadecimal digits are expected", a.Env)
	}
	sum := sha256.Sum256(raw)
	if got := hex.EncodeToString(sum[:]); got != wantLower {
		return fmt.Errorf("cv asset %s of release %s does not match %s in cv-release.env: got %s, pinned %s",
			a.Name, tag, a.Env, got, wantLower)
	}
	return nil
}

func releaseURL(base, tag, file string) (string, error) {
	owner, repo, _ := strings.Cut(CVRepository, "/")
	u, err := url.JoinPath(base, url.PathEscape(owner), url.PathEscape(repo), "releases", "download", url.PathEscape(tag), url.PathEscape(file))
	if err != nil {
		return "", fmt.Errorf("release url: %w", err)
	}
	return u, nil
}

// redirectPolicy follows at most maxRedirects hops, only to https URLs on the
// host the download started at or a name in ReleaseRedirectHosts, matched
// whole, never as a suffix.
func redirectPolicy(baseHost string) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		switch {
		case len(via) > maxRedirects:
			return fmt.Errorf("%w: more than %d redirects", errRedirectRefused, maxRedirects)
		case req.URL.Scheme != "https" && req.URL.Host != baseHost:
			return fmt.Errorf("%w: to a %s URL on %s, only https is followed", errRedirectRefused, req.URL.Scheme, req.URL.Host)
		}
		if req.URL.Host == baseHost {
			return nil
		}
		for _, h := range ReleaseRedirectHosts {
			if req.URL.Hostname() == h {
				return nil
			}
		}
		return fmt.Errorf("%w: to host %s, which is not allowed", errRedirectRefused, req.URL.Host)
	}
}

// DownloadAssets fetches every asset of the pinned release, by asset name. The
// assets are public, so the request carries no credential. base is
// DefaultReleaseBase when empty.
func DownloadAssets(ctx context.Context, hc *http.Client, base, tag string) (map[string][]byte, error) {
	if err := ValidateTag(tag); err != nil {
		return nil, err
	}
	if base == "" {
		base = DefaultReleaseBase
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("release base: %w", err)
	}
	client := http.Client{Timeout: 60 * time.Second}
	if hc != nil {
		client = *hc
	}
	client.CheckRedirect = redirectPolicy(baseURL.Host)

	out := map[string][]byte{}
	for _, a := range Assets() {
		u, err := releaseURL(base, tag, a.Name)
		if err != nil {
			return nil, err
		}
		limit := int64(maxJSON)
		if a.PDF {
			limit = maxPDF
		}
		data, err := download(ctx, &client, u, limit)
		if err != nil {
			return nil, err
		}
		out[a.Name] = data
	}
	return out, nil
}

func download(ctx context.Context, hc *http.Client, target string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", target, err)
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", target, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %d", target, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", target, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("download %s: larger than %d bytes", target, limit)
	}
	return data, nil
}

// ReadAssets reads every asset from a local directory, for an offline build and
// for the tests.
func ReadAssets(dir string) (map[string][]byte, error) {
	out := map[string][]byte{}
	for _, a := range Assets() {
		data, err := os.ReadFile(filepath.Join(dir, a.Name))
		if err != nil {
			return nil, fmt.Errorf("read asset: %w", err)
		}
		out[a.Name] = data
	}
	return out, nil
}
