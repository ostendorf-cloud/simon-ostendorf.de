// Package content builds the data the CV page reads. At build time it fetches
// the cv repository's release assets (a JSON artefact and a PDF per language),
// verifies them against the pin in cv-release.env and writes the page's dataset
// and the PDFs into web/src/generated. Nothing in this package runs at request
// time.
package content

import (
	"encoding/json"
	"fmt"
)

// SupportedSchemaVersion is the only cv artefact schema version this package
// reads. The cv repository raises its version on any change a consumer reading
// the old shape would get wrong, so an unknown version fails the build instead
// of being guessed at.
const SupportedSchemaVersion = 1

// Languages is every language the cv repository publishes, one artefact and
// one PDF each.
var Languages = []string{"de", "en"}

// ArtefactFileName is the release asset name of the JSON artefact of lang, for
// example cv.en.json. It is part of the contract with the cv repository.
func ArtefactFileName(lang string) string {
	return "cv." + lang + ".json"
}

// PDFFileName is the release asset name of the PDF of lang, for example
// cv.en.pdf. The cv repository does not attach a PDF yet, so this name is this
// site's assumption: it mirrors the JSON name and must be agreed with cv.
func PDFFileName(lang string) string {
	return "cv." + lang + ".pdf"
}

// Document is one language's CV as the page shows it. It carries only what the
// page renders: the artefact's references, projects, contact data (the postal
// address among it) and the sections the page has no place for are dropped
// here, so they cannot reach a visitor by accident.
type Document struct {
	Language     string        `json:"language"`
	Experience   []Experience  `json:"experience"`
	Education    []Education   `json:"education"`
	Certificates []Certificate `json:"certificates"`
	Skills       []Skill       `json:"skills"`
	Engagements  []Engagement  `json:"engagements"`
	Languages    []Proficiency `json:"languages"`
}

// Experience is one employment. End is null for an ongoing one.
type Experience struct {
	ID             string  `json:"id"`
	Employer       string  `json:"employer"`
	Title          string  `json:"title"`
	Start          string  `json:"start"`
	End            *string `json:"end"`
	Summary        *string `json:"summary"`
	Location       *string `json:"location"`
	EmploymentType *string `json:"employment_type"`
}

// Education is one course of study.
type Education struct {
	ID          string  `json:"id"`
	Institution string  `json:"institution"`
	Degree      string  `json:"degree"`
	Start       string  `json:"start"`
	End         *string `json:"end"`
	Summary     *string `json:"summary"`
	Grade       *string `json:"grade"`
}

// Certificate is one certificate. ValidUntil (YYYY-MM) is null for one that
// does not expire.
type Certificate struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Issuer     string  `json:"issuer"`
	Issued     string  `json:"issued"`
	ValidUntil *string `json:"valid_until"`
}

// Skill is one skill with its level.
type Skill struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level string `json:"level"`
}

// Engagement is one anonymised project of a position. Experience is the id of
// the position it belongs to.
type Engagement struct {
	ID         string  `json:"id"`
	Experience string  `json:"experience"`
	Title      string  `json:"title"`
	Industry   *string `json:"industry"`
	Start      *string `json:"start"`
	End        *string `json:"end"`
	Summary    string  `json:"summary"`
}

// Proficiency is one spoken language and the level it is spoken at.
type Proficiency struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level string `json:"level"`
}

// artefactDocument is the artefact's shape on the wire: Document plus the
// schema version, which the dataset does not carry.
type artefactDocument struct {
	SchemaVersion int `json:"schema_version"`
	Document
}

// ParseArtefact reads the artefact of lang. It checks the schema version before
// anything else, so a document of a future shape gets a clear message instead
// of a decoding error about one of its fields. It then refuses any entry that
// carries a visibility marker other than public. Unknown keys are ignored.
// A list that is missing or null becomes an empty list.
func ParseArtefact(data []byte, lang string) (Document, error) {
	var head struct {
		SchemaVersion *int `json:"schema_version"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return Document{}, fmt.Errorf("cv artefact %s: not valid JSON: %w", lang, err)
	}
	if head.SchemaVersion == nil {
		return Document{}, fmt.Errorf("cv artefact %s: schema_version is missing", lang)
	}
	if *head.SchemaVersion != SupportedSchemaVersion {
		return Document{}, fmt.Errorf(
			"cv artefact %s: unknown schema_version %d, this build reads version %d only: "+
				"update internal/content for the new version, or pin an older cv release",
			lang, *head.SchemaVersion, SupportedSchemaVersion)
	}

	if err := rejectNonPublic(data); err != nil {
		return Document{}, fmt.Errorf("cv artefact %s: %w", lang, err)
	}

	var doc artefactDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return Document{}, fmt.Errorf("cv artefact %s: %w", lang, err)
	}
	if doc.Language != lang {
		return Document{}, fmt.Errorf("cv artefact %s: language is %q, want %q", lang, doc.Language, lang)
	}
	d := doc.Document
	d.normalize()
	return d, nil
}

// rejectNonPublic fails on any object, at any depth, that marks itself as not
// public: a "visibility" key with a value other than "public", or a "public"
// key that is false or anything but true. The cv repository's current model has
// no such marker and filters nothing, because its data files are public. An
// artefact that carries one anyway is of a shape this site does not understand,
// and the safe reading is that the entry is not meant for the open web, so the
// build fails instead of showing it.
func rejectNonPublic(data []byte) error {
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("not valid JSON: %w", err)
	}
	return walkVisibility(root, "$")
}

func walkVisibility(v any, path string) error {
	switch x := v.(type) {
	case map[string]any:
		if vis, ok := x["visibility"]; ok && vis != "public" {
			return fmt.Errorf("%s: entry is marked visibility %v, only public entries may reach this site", path, vis)
		}
		if pub, ok := x["public"]; ok && pub != true {
			return fmt.Errorf("%s: entry is marked public %v, only public entries may reach this site", path, pub)
		}
		for k, child := range x {
			if err := walkVisibility(child, path+"."+k); err != nil {
				return err
			}
		}
	case []any:
		for i, child := range x {
			if err := walkVisibility(child, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

// normalize turns a missing or null list into an empty one, so the page never
// has to tell the two apart.
func (d *Document) normalize() {
	if d.Experience == nil {
		d.Experience = []Experience{}
	}
	if d.Education == nil {
		d.Education = []Education{}
	}
	if d.Certificates == nil {
		d.Certificates = []Certificate{}
	}
	if d.Skills == nil {
		d.Skills = []Skill{}
	}
	if d.Engagements == nil {
		d.Engagements = []Engagement{}
	}
	if d.Languages == nil {
		d.Languages = []Proficiency{}
	}
}
