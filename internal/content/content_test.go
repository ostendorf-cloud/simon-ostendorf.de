package content

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sample(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "assets", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func sum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func TestParseArtefactSample(t *testing.T) {
	doc, err := ParseArtefact(sample(t, "cv.en.json"), "en")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Experience) != 2 || len(doc.Certificates) != 2 || len(doc.Engagements) != 1 {
		t.Fatalf("unexpected sections: %+v", doc)
	}
	if doc.Certificates[1].ValidUntil != nil {
		t.Fatal("a certificate without expiry must keep a null valid_until")
	}
	if doc.Engagements[0].Experience != "exp-1" {
		t.Fatalf("engagement lost its position: %+v", doc.Engagements[0])
	}
}

func TestParseArtefactSchemaVersion(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"future version", `{"schema_version": 2, "language": "en"}`, "unknown schema_version 2"},
		{"version zero", `{"schema_version": 0, "language": "en"}`, "unknown schema_version 0"},
		{"missing version", `{"language": "en"}`, "schema_version is missing"},
		{"not json", `nope`, "not valid JSON"},
		{"wrong language", `{"schema_version": 1, "language": "de"}`, `language is "de", want "en"`},
		{"trailing data", `{"schema_version": 1, "language": "en"} {}`, "not valid JSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArtefact([]byte(tt.body), "en")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestUnknownVersionMessageIsActionable(t *testing.T) {
	_, err := ParseArtefact([]byte(`{"schema_version": 7, "language": "en"}`), "en")
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"cv artefact en", "7", "version 1 only", "pin an older cv release"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message %q lacks %q", err, want)
		}
	}
}

func TestParseArtefactEmptySectionsBecomeEmptyLists(t *testing.T) {
	doc, err := ParseArtefact([]byte(`{"schema_version": 1, "language": "en", "skills": null}`), "en")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Skills == nil || doc.Engagements == nil || doc.Languages == nil {
		t.Fatal("a missing or null list must become an empty list")
	}
}

// TestNoNonPublicEntryCanAppear feeds the parser an artefact that wrongly
// contains an entry marked other than public, in several places and shapes.
// The cv repository's model carries no such marker today: the guard is the
// fail closed answer to an artefact that does.
func TestNoNonPublicEntryCanAppear(t *testing.T) {
	tests := []struct {
		name  string
		entry string
	}{
		{"experience private", `"experience": [{"id": "x", "employer": "Secret", "title": "T", "start": "2020-01", "visibility": "private"}]`},
		{"engagement internal", `"engagements": [{"id": "x", "experience": "e", "title": "Secret", "summary": "S", "visibility": "internal"}]`},
		{"certificate not public flag", `"certificates": [{"id": "x", "name": "Secret", "issuer": "I", "issued": "2020-01", "public": false}]`},
		{"nested in evidence", `"skills": [{"id": "x", "name": "N", "level": "expert", "evidence": [{"type": "project", "ref": "p", "visibility": "private"}]}]`},
		{"contact address", `"contact": {"email": "e", "address": {"street": "Secret", "visibility": "private"}}`},
		{"visibility not a string", `"experience": [{"id": "x", "employer": "Secret", "title": "T", "start": "2020-01", "visibility": null}]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"schema_version": 1, "language": "en", ` + tt.entry + `}`
			doc, err := ParseArtefact([]byte(body), "en")
			if err == nil {
				t.Fatalf("parsed without error, document: %+v", doc)
			}
			if !strings.Contains(err.Error(), "only public entries") {
				t.Fatalf("message does not explain the refusal: %v", err)
			}
		})
	}
}

func TestPublicMarkerIsAccepted(t *testing.T) {
	body := `{"schema_version": 1, "language": "en", "skills": [{"id": "x", "name": "N", "level": "expert", "visibility": "public"}]}`
	if _, err := ParseArtefact([]byte(body), "en"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTag(t *testing.T) {
	for _, tag := range []string{"v0.1.0", "v10.20.30", "v1.0.0-rc.1"} {
		if err := ValidateTag(tag); err != nil {
			t.Errorf("%q: %v", tag, err)
		}
	}
	for _, tag := range []string{"", "latest", "main", "0.1.0", "v1", "v1.0.0/../x", "v1.0.0?x=1", "v1.0.0 "} {
		if err := ValidateTag(tag); err == nil {
			t.Errorf("%q must be refused", tag)
		}
	}
}

func TestCheckDigest(t *testing.T) {
	a := Assets()[0]
	raw := []byte("content")
	good := sum(raw)
	if err := CheckDigest(a, "v0.1.0", raw, good); err != nil {
		t.Fatal(err)
	}
	if err := CheckDigest(a, "v0.1.0", raw, strings.ToUpper(good)); err != nil {
		t.Fatalf("upper case digest: %v", err)
	}
	other := sum([]byte("other"))
	tests := map[string]string{"": "is empty", "xyz": "not a sha256", other: "does not match"}
	for want, msg := range tests {
		err := CheckDigest(a, "v0.1.0", raw, want)
		if err == nil || !strings.Contains(err.Error(), msg) || !strings.Contains(err.Error(), "cv-release.env") {
			t.Errorf("digest %q: error = %v, want %q", want, err, msg)
		}
	}
}

func fixtureOptions(t *testing.T) Options {
	t.Helper()
	assets := map[string][]byte{}
	digests := map[string]string{}
	for _, a := range Assets() {
		assets[a.Name] = sample(t, a.Name)
		digests[a.Env] = sum(assets[a.Name])
	}
	return Options{
		Tag: "v0.1.0", Digests: digests, Assets: assets,
		Now: func() time.Time { return time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC) },
	}
}

func TestGenerate(t *testing.T) {
	out, err := Generate(fixtureOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	ds := string(out.Dataset)
	for _, want := range []string{`"release": "v0.1.0"`, `"generated_at": "2026-11-20T12:00:00Z"`, "Example Engagement", "Beispiel Einsatz"} {
		if !strings.Contains(ds, want) {
			t.Errorf("dataset lacks %q", want)
		}
	}
	// Whatever the page has no place for stays out of the dataset.
	for _, banned := range []string{"Dropped", "Entfallene", "dropped@example.com", "example/repo", "evidence"} {
		if strings.Contains(ds, banned) {
			t.Errorf("dataset carries %q", banned)
		}
	}
	if len(out.PDFs) != 2 {
		t.Fatalf("PDFs: %d", len(out.PDFs))
	}
}

func TestGenerateRefusesWrongDigestAndNamesIt(t *testing.T) {
	opts := fixtureOptions(t)
	opts.Assets["cv.en.pdf"] = append([]byte("%PDF-1.4\n"), "tampered"...)
	_, err := Generate(opts)
	if err == nil || !strings.Contains(err.Error(), "CV_RELEASE_SHA256_PDF_EN") {
		t.Fatalf("error = %v", err)
	}
}

func TestGenerateRefusesUnpinnedRelease(t *testing.T) {
	opts := fixtureOptions(t)
	opts.Digests = map[string]string{}
	if _, err := Generate(opts); err == nil || !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("error = %v", err)
	}
}

func TestGenerateRefusesUnknownSchemaVersion(t *testing.T) {
	opts := fixtureOptions(t)
	raw := []byte(strings.Replace(string(opts.Assets["cv.de.json"]), `"schema_version": 1`, `"schema_version": 2`, 1))
	opts.Assets["cv.de.json"] = raw
	opts.Digests["CV_RELEASE_SHA256_DE"] = sum(raw)
	if _, err := Generate(opts); err == nil || !strings.Contains(err.Error(), "unknown schema_version 2") {
		t.Fatalf("error = %v", err)
	}
}

func TestGenerateRefusesNonPDF(t *testing.T) {
	opts := fixtureOptions(t)
	raw := []byte("<html>not a pdf</html>")
	opts.Assets["cv.de.pdf"] = raw
	opts.Digests["CV_RELEASE_SHA256_PDF_DE"] = sum(raw)
	if _, err := Generate(opts); err == nil || !strings.Contains(err.Error(), "is not a PDF") {
		t.Fatalf("error = %v", err)
	}
}

func TestGenerateLocalSkipsThePin(t *testing.T) {
	opts := fixtureOptions(t)
	opts.Local = true
	opts.Tag = "local"
	opts.Digests = nil
	if _, err := Generate(opts); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	a, err := Generate(fixtureOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate(fixtureOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	if string(a.Dataset) != string(b.Dataset) {
		t.Fatal("same input, different bytes")
	}
}

func releaseServer(t *testing.T, tag string) *httptest.Server {
	t.Helper()
	prefix := "/ostendorf-cloud/cv/releases/download/" + tag + "/"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, ok := strings.CutPrefix(r.URL.Path, prefix)
		if !ok {
			http.NotFound(w, r)
			return
		}
		data, err := os.ReadFile(filepath.Join("testdata", "assets", name))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("a download must never carry a credential")
		}
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestDownloadAssets(t *testing.T) {
	srv := releaseServer(t, "v0.1.0")
	got, err := DownloadAssets(context.Background(), nil, srv.URL, "v0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("assets: %d", len(got))
	}
	if string(got["cv.de.json"]) != string(sample(t, "cv.de.json")) {
		t.Fatal("content differs")
	}
}

func TestDownloadAssetsRefusesMovingReference(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("no request may be made for a moving reference")
	}))
	defer srv.Close()
	for _, tag := range []string{"", "latest", "main"} {
		if _, err := DownloadAssets(context.Background(), nil, srv.URL, tag); err == nil {
			t.Errorf("%q must be refused", tag)
		}
	}
}

func TestDownloadAssetsFailsOnMissingAsset(t *testing.T) {
	srv := releaseServer(t, "v0.9.9")
	_, err := DownloadAssets(context.Background(), nil, srv.URL, "v0.1.0")
	if err == nil || !strings.Contains(err.Error(), "status 404") {
		t.Fatalf("error = %v", err)
	}
}

func TestDownloadRefusesRedirectToOtherHost(t *testing.T) {
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("the redirect target must not be contacted")
	}))
	defer evil.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, evil.URL+"/x", http.StatusFound)
	}))
	defer srv.Close()
	_, err := DownloadAssets(context.Background(), nil, srv.URL, "v0.1.0")
	if err == nil || !strings.Contains(err.Error(), "redirect refused") {
		t.Fatalf("error = %v", err)
	}
}

func TestDownloadRejectsOversizedAsset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", 100)))
	}))
	defer srv.Close()
	hc := &http.Client{}
	if _, err := download(context.Background(), hc, srv.URL, 10); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("error = %v", err)
	}
}

func TestWriteOutput(t *testing.T) {
	out, err := Generate(fixtureOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "generated")
	if err := WriteOutput(dir, out); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cv.json", "cv.de.pdf", "cv.en.pdf"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Error(err)
		}
	}
}
