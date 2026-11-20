package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLocalWritesDatasetAndPDFs(t *testing.T) {
	out := t.TempDir()
	var stderr bytes.Buffer
	args := []string{"-asset-dir", "../../internal/content/testdata/assets", "-out", out}
	env := map[string]string{"SOURCE_DATE_EPOCH": "1790000000"}
	if err := run(context.Background(), args, func(k string) string { return env[k] }, &stderr); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "cv.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"release": "local"`) {
		t.Fatalf("dataset: %s", data)
	}
	if _, err := os.Stat(filepath.Join(out, "cv.en.pdf")); err != nil {
		t.Fatal(err)
	}
}

func TestRunWithoutPinFailsAndWritesNothing(t *testing.T) {
	out := filepath.Join(t.TempDir(), "generated")
	var stderr bytes.Buffer
	err := run(context.Background(), []string{"-out", out}, func(string) string { return "" }, &stderr)
	if err == nil || !strings.Contains(err.Error(), "no cv release is pinned") {
		t.Fatalf("error = %v", err)
	}
	if _, statErr := os.Stat(out); statErr == nil {
		t.Fatal("a failed run must write nothing")
	}
}
