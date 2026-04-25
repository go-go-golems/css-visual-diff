package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/stretchr/testify/require"
)

func TestSnapshotPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" style="color: rgb(255, 0, 0)">Book now</button><p id="copy">Hello</p></body></html>`)
	}))
	defer server.Close()

	page := newSnapshotTestPage(t, server.URL)
	snapshot, err := SnapshotPage(page, []SnapshotProbeSpec{
		{Name: "cta", Selector: "#cta", Required: true, Extractors: []ExtractorSpec{{Kind: ExtractorText, Text: TextOptions{NormalizeWhitespace: true, Trim: true}}, {Kind: ExtractorComputedStyle, Props: []string{"color"}}}},
		{Name: "copy", Selector: "#copy", Required: true, Extractors: []ExtractorSpec{{Kind: ExtractorText, Text: TextOptions{NormalizeWhitespace: true, Trim: true}}}},
	})
	require.NoError(t, err)
	require.Len(t, snapshot.Results, 2)
	require.Equal(t, "cta", snapshot.Results[0].Name)
	require.Equal(t, "Book now", snapshot.Results[0].Snapshot.Text)
	require.Equal(t, "rgb(255, 0, 0)", snapshot.Results[0].Snapshot.Computed["color"])
	require.Equal(t, "Hello", snapshot.Results[1].Snapshot.Text)
}

func TestSnapshotPageOptionalProbeRecordsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta">Book</button></body></html>`)
	}))
	defer server.Close()

	page := newSnapshotTestPage(t, server.URL)
	snapshot, err := SnapshotPage(page, []SnapshotProbeSpec{
		{Name: "invalid", Selector: "#bad[", Required: false, Extractors: []ExtractorSpec{{Kind: ExtractorExists}}},
	})
	require.NoError(t, err)
	require.Len(t, snapshot.Results, 1)
	require.Contains(t, snapshot.Results[0].Error, "selector status")
}

func newSnapshotTestPage(t *testing.T, url string) *driver.Page {
	t.Helper()
	browser, err := driver.NewBrowser(context.Background())
	require.NoError(t, err)
	t.Cleanup(browser.Close)

	page, err := browser.NewPage()
	require.NoError(t, err)
	t.Cleanup(page.Close)

	require.NoError(t, page.SetViewport(400, 300))
	require.NoError(t, page.Goto(url))
	return page
}
