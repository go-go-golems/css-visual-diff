package main

import (
	"path/filepath"
	"testing"
)

func TestReviewComparePathRejectsTraversalSegments(t *testing.T) {
	base := t.TempDir()
	badInputs := []struct {
		name    string
		page    string
		section string
	}{
		{name: "page dotdot", page: "..", section: "content"},
		{name: "section dotdot", page: "shows", section: ".."},
		{name: "page slash traversal", page: "../secret", section: "content"},
		{name: "section slash traversal", page: "shows", section: "../secret"},
		{name: "page absolute", page: filepath.Join(string(filepath.Separator), "tmp"), section: "content"},
		{name: "section absolute", page: "shows", section: filepath.Join(string(filepath.Separator), "tmp")},
		{name: "page nested path", page: "shows/archive", section: "content"},
		{name: "section nested path", page: "shows", section: "content/archive"},
	}

	for _, tc := range badInputs {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := reviewComparePath(base, tc.page, tc.section); err == nil {
				t.Fatalf("expected %q/%q to be rejected", tc.page, tc.section)
			}
		})
	}
}

func TestReviewComparePathStaysUnderDataDir(t *testing.T) {
	base := t.TempDir()
	got, err := reviewComparePath(base, "shows", "content")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(base, "shows", "artifacts", "content", "compare.json")
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestReviewArtifactPathRejectsTraversalFile(t *testing.T) {
	base := t.TempDir()
	badFiles := []string{"..", "../secret.json", "nested/secret.json", filepath.Join(string(filepath.Separator), "tmp", "secret.json")}
	for _, file := range badFiles {
		t.Run(file, func(t *testing.T) {
			if _, err := reviewArtifactPath(base, "shows", "content", file); err == nil {
				t.Fatalf("expected file %q to be rejected", file)
			}
		})
	}
}
