package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DatasetFile is the name of the dataset the page reads, inside the output
// directory. The PDFs sit next to it under PDFFileName.
const DatasetFile = "cv.json"

// Dataset is what the page reads: the release it came from, the build time (the
// reference for the expiry mark of a certificate) and one document per language.
type Dataset struct {
	Release     string              `json:"release"`
	GeneratedAt string              `json:"generated_at"`
	Documents   map[string]Document `json:"documents"`
}

// Options is the input of Generate.
type Options struct {
	// Tag is the pinned release, "local" for an offline build from a directory.
	Tag string
	// Digests maps an asset's Env name to its pinned sha256. Ignored when Local.
	Digests map[string]string
	// Local skips the pin: the assets came from a directory, not from a release.
	Local bool
	// Assets holds every asset's content by asset name.
	Assets map[string][]byte
	// Now is the clock for generated_at.
	Now func() time.Time
}

// Output is what Generate produced: the dataset bytes and the PDFs by language.
type Output struct {
	Dataset []byte
	PDFs    map[string][]byte
}

// Generate verifies every asset against the pin, parses the JSON artefacts and
// builds the dataset. Any failure fails the whole build: there is no partial
// dataset.
func Generate(opts Options) (Output, error) {
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	ds := Dataset{
		Release:     opts.Tag,
		GeneratedAt: now().UTC().Format(time.RFC3339),
		Documents:   map[string]Document{},
	}
	out := Output{PDFs: map[string][]byte{}}
	var errs []error
	for _, a := range Assets() {
		raw, ok := opts.Assets[a.Name]
		if !ok {
			errs = append(errs, fmt.Errorf("asset %s is missing", a.Name))
			continue
		}
		if !opts.Local {
			if err := CheckDigest(a, opts.Tag, raw, opts.Digests[a.Env]); err != nil {
				errs = append(errs, err)
				continue
			}
		}
		if a.PDF {
			if !bytes.HasPrefix(raw, []byte("%PDF-")) {
				errs = append(errs, fmt.Errorf("asset %s is not a PDF", a.Name))
				continue
			}
			out.PDFs[a.Lang] = raw
			continue
		}
		doc, err := ParseArtefact(raw, a.Lang)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ds.Documents[a.Lang] = doc
	}
	if err := errors.Join(errs...); err != nil {
		return Output{}, err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ds); err != nil {
		return Output{}, fmt.Errorf("encode dataset: %w", err)
	}
	out.Dataset = buf.Bytes()
	return out, nil
}

// WriteOutput writes the dataset and the PDFs into dir, each in one step so a
// failed run never leaves half a file.
func WriteOutput(dir string, out Output) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, DatasetFile), out.Dataset); err != nil {
		return err
	}
	for lang, data := range out.PDFs {
		if err := writeFile(filepath.Join(dir, PDFFileName(lang)), data); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".cv-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
