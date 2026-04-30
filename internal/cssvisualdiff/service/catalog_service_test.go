package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalogWritesManifestAndIndex(t *testing.T) {
	outDir := t.TempDir()
	catalog, err := NewCatalog(CatalogOptions{Title: "Smoke Catalog", OutDir: outDir, ArtifactRoot: "../artifacts//raw", IndexName: "../index.md"})
	require.NoError(t, err)

	artifactDir := catalog.ArtifactDir("../Prototype Public Shows!")
	require.Equal(t, filepath.Join(outDir, "artifacts", "raw", "prototype-public-shows"), artifactDir)

	target := CatalogTargetRecord{
		Slug:     "Prototype Public Shows!",
		Name:     "Prototype Public Shows",
		URL:      "http://example.test/shows",
		Selector: "#root",
		Viewport: Viewport{Width: 400, Height: 300},
	}
	catalog.AddTarget(target)
	catalog.RecordPreflight(target, []SelectorStatus{{Name: "root", Selector: "#root", Exists: true, Visible: true}})
	catalog.AddResult(target, InspectResult{OutputDir: artifactDir, Results: []InspectArtifactResult{{
		Metadata: InspectMetadata{Name: "root", Selector: "#root"},
		Style:    &StyleSnapshot{Exists: true, Computed: map[string]string{"color": "rgb(0, 0, 0)"}},
	}}})
	catalog.AddFailure(CatalogTargetRecord{Slug: "Broken Target"}, CatalogFailureRecord{Code: "SELECTOR_ERROR", Operation: "inspect", Message: "missing selector"})

	manifestPath, err := catalog.WriteManifest()
	require.NoError(t, err)
	indexPath, err := catalog.WriteIndex()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(outDir, "manifest.json"), manifestPath)
	require.Equal(t, filepath.Join(outDir, "index.md"), indexPath)

	manifestBytes, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	var manifest CatalogManifest
	require.NoError(t, json.Unmarshal(manifestBytes, &manifest))
	require.Equal(t, CatalogSchemaVersion, manifest.SchemaVersion)
	require.Equal(t, "Smoke Catalog", manifest.Title)
	require.Equal(t, "artifacts/raw", filepath.ToSlash(manifest.ArtifactRoot))
	require.Len(t, manifest.Targets, 2)
	require.Equal(t, "prototype-public-shows", manifest.Targets[0].Slug)
	require.Equal(t, CatalogSummary{TargetCount: 2, PreflightCount: 1, ResultCount: 1, FailureCount: 1, ArtifactCount: 1}, manifest.Summary)

	indexBytes, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	require.Contains(t, string(indexBytes), "# Smoke Catalog")
	require.Contains(t, string(indexBytes), "prototype-public-shows")
	require.Contains(t, string(indexBytes), "SELECTOR_ERROR")
}

func TestSanitizeCatalogSlug(t *testing.T) {
	require.Equal(t, "prototype-public-shows", SanitizeCatalogSlug("../Prototype Public Shows!"))
	require.Equal(t, "a-b_c", SanitizeCatalogSlug(" A---B_C "))
	require.Equal(t, "target", NormalizeCatalogTarget(CatalogTargetRecord{}).Slug)
}
