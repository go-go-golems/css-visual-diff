package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiffValuesEqualChangedAndIgnored(t *testing.T) {
	before := map[string]any{"results": []any{map[string]any{"name": "cta", "snapshot": map[string]any{"text": "Book", "computed": map[string]any{"color": "red"}}}}}
	after := map[string]any{"results": []any{map[string]any{"name": "cta", "snapshot": map[string]any{"text": "Book now", "computed": map[string]any{"color": "red"}}}}}

	diff, err := DiffValues(before, after, DiffOptions{})
	require.NoError(t, err)
	require.False(t, diff.Equal)
	require.Equal(t, 1, diff.ChangeCount)
	require.Equal(t, "results[0].snapshot.text", diff.Changes[0].Path)

	ignored, err := DiffValues(before, after, DiffOptions{IgnorePaths: []string{"results[0].snapshot.text"}})
	require.NoError(t, err)
	require.True(t, ignored.Equal)
	require.Empty(t, ignored.Changes)

	equal, err := DiffValues(before, before, DiffOptions{})
	require.NoError(t, err)
	require.True(t, equal.Equal)
}

func TestRenderDiffMarkdown(t *testing.T) {
	diff := SnapshotDiff{Equal: false, ChangeCount: 1, Changes: []DiffChange{{Path: "results[0].text", Before: "A", After: "B"}}}
	markdown := RenderDiffMarkdown(diff)
	require.True(t, strings.HasPrefix(markdown, "# Snapshot Diff"))
	require.Contains(t, markdown, "1 change")
	require.Contains(t, markdown, "results[0].text")
}
