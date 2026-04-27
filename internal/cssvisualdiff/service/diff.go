package service

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type DiffOptions struct {
	IgnorePaths []string `json:"ignorePaths,omitempty"`
}

type DiffChange struct {
	Path   string `json:"path"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

type SnapshotDiff struct {
	Equal       bool         `json:"equal"`
	ChangeCount int          `json:"change_count"`
	Changes     []DiffChange `json:"changes"`
}

func DiffValues(before, after any, opts DiffOptions) (SnapshotDiff, error) {
	beforePlain, err := normalizeDiffValue(before)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("normalize before value: %w", err)
	}
	afterPlain, err := normalizeDiffValue(after)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("normalize after value: %w", err)
	}
	ignored := map[string]bool{}
	for _, path := range opts.IgnorePaths {
		ignored[path] = true
	}
	changes := []DiffChange{}
	walkDiff("", beforePlain, afterPlain, ignored, &changes)
	return SnapshotDiff{Equal: len(changes) == 0, ChangeCount: len(changes), Changes: changes}, nil
}

func RenderDiffMarkdown(diff SnapshotDiff) string {
	var b strings.Builder
	if diff.Equal {
		b.WriteString("# Snapshot Diff\n\nNo changes.\n")
		return b.String()
	}
	fmt.Fprintf(&b, "# Snapshot Diff\n\n%d change(s).\n\n", diff.ChangeCount)
	for _, change := range diff.Changes {
		fmt.Fprintf(&b, "- `%s`: `%v` -> `%v`\n", change.Path, change.Before, change.After)
	}
	return b.String()
}

func normalizeDiffValue(value any) (any, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func walkDiff(path string, before, after any, ignored map[string]bool, changes *[]DiffChange) {
	if ignored[path] {
		return
	}
	beforeMap, beforeIsMap := before.(map[string]any)
	afterMap, afterIsMap := after.(map[string]any)
	if beforeIsMap || afterIsMap {
		if !beforeIsMap || !afterIsMap {
			*changes = append(*changes, DiffChange{Path: path, Before: before, After: after})
			return
		}
		keys := map[string]bool{}
		for k := range beforeMap {
			keys[k] = true
		}
		for k := range afterMap {
			keys[k] = true
		}
		ordered := make([]string, 0, len(keys))
		for k := range keys {
			ordered = append(ordered, k)
		}
		sort.Strings(ordered)
		for _, k := range ordered {
			walkDiff(joinDiffPath(path, k), beforeMap[k], afterMap[k], ignored, changes)
		}
		return
	}
	beforeSlice, beforeIsSlice := before.([]any)
	afterSlice, afterIsSlice := after.([]any)
	if beforeIsSlice || afterIsSlice {
		if !beforeIsSlice || !afterIsSlice || len(beforeSlice) != len(afterSlice) {
			*changes = append(*changes, DiffChange{Path: path, Before: before, After: after})
			return
		}
		for i := range beforeSlice {
			walkDiff(fmt.Sprintf("%s[%d]", path, i), beforeSlice[i], afterSlice[i], ignored, changes)
		}
		return
	}
	if !reflect.DeepEqual(before, after) {
		*changes = append(*changes, DiffChange{Path: path, Before: before, After: after})
	}
}

func joinDiffPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
