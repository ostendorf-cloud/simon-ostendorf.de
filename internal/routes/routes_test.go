package routes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParsesTheCheckedInRouteTable(t *testing.T) {
	t.Parallel()

	table, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := []Route{
		{Name: "home", Slug: "", File: "index.html"},
		{Name: "cv", Slug: "cv", File: "cv/index.html"},
		{Name: "contact", Slug: "contact", File: "contact/index.html"},
		{Name: "imprint", Slug: "imprint", File: "imprint/index.html"},
		{Name: "privacy", Slug: "privacy", File: "privacy/index.html"},
	}
	if len(table) != len(want) {
		t.Fatalf("Load() returned %d routes, want %d: %+v", len(table), len(want), table)
	}
	for i, r := range want {
		if table[i] != r {
			t.Errorf("Load()[%d] = %+v, want %+v", i, table[i], r)
		}
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   string
		want    []Route
		wantErr bool
	}{
		"valid table": {
			input: `[{"name":"home","slug":"","file":"index.html"}]`,
			want:  []Route{{Name: "home", Slug: "", File: "index.html"}},
		},
		"empty table": {
			input: `[]`,
			want:  []Route{},
		},
		"malformed json": {
			input:   `{not json`,
			wantErr: true,
		},
		"entry missing name": {
			input:   `[{"name":"","slug":"imprint","file":"imprint/index.html"}]`,
			wantErr: true,
		},
		"entry missing file": {
			input:   `[{"name":"imprint","slug":"imprint","file":""}]`,
			wantErr: true,
		},
		"duplicate slug": {
			input: `[
				{"name":"home","slug":"dup","file":"a.html"},
				{"name":"other","slug":"dup","file":"b.html"}
			]`,
			wantErr: true,
		},
		"duplicate name": {
			input: `[
				{"name":"dup","slug":"a","file":"a.html"},
				{"name":"dup","slug":"b","file":"b.html"}
			]`,
			wantErr: true,
		},
		"duplicate file": {
			input: `[
				{"name":"home","slug":"a","file":"dup.html"},
				{"name":"other","slug":"b","file":"dup.html"}
			]`,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse([]byte(tc.input))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) error = nil, want an error", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tc.input, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("Parse(%q) = %+v, want %+v", tc.input, got, tc.want)
			}
			for i, r := range tc.want {
				if got[i] != r {
					t.Errorf("Parse(%q)[%d] = %+v, want %+v", tc.input, i, got[i], r)
				}
			}
		})
	}
}

func TestParseRejectsNilData(t *testing.T) {
	t.Parallel()

	if _, err := Parse(nil); err == nil {
		t.Fatal("Parse(nil) error = nil, want an error")
	}
}

// TestParseAgainstSharedFixtures runs Parse against every file under
// testdata/invalid and testdata/valid, the same fixtures
// web/src/lib/routeTable.test.ts reads to check the frontend's own
// validateRouteTable. A fixture added under either directory is
// exercised by both test suites without any further code change, which
// is what proves the two validators keep agreeing on a given case rather
// than merely being alike by eye; see Parse's own doc comment for what it
// cannot catch.
func TestParseAgainstSharedFixtures(t *testing.T) {
	t.Parallel()

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		checkFixtures(t, "testdata/invalid", true)
	})
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		checkFixtures(t, "testdata/valid", false)
	})
}

// checkFixtures runs Parse over every JSON file in dir, each expected to
// fail when wantErr is true and to succeed otherwise.
func checkFixtures(t *testing.T, dir string, wantErr bool) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("os.ReadDir(%q) error = %v", dir, err)
	}
	if len(entries) == 0 {
		t.Fatalf("%q has no fixtures", dir)
	}

	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("os.ReadFile(%q) error = %v", entry.Name(), err)
			}

			_, err = Parse(data)
			if wantErr && err == nil {
				t.Fatalf("Parse(%s) error = nil, want an error", entry.Name())
			}
			if !wantErr && err != nil {
				t.Fatalf("Parse(%s) error = %v, want nil", entry.Name(), err)
			}
		})
	}
}
