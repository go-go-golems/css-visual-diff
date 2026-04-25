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

func TestExtractElementMultipleExtractors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="primary" style="color: rgb(255, 0, 0)">Book now</button></body></html>`)
	}))
	defer server.Close()

	page := newExtractTestPage(t, server.URL)
	snapshot, err := ExtractElement(page, LocatorSpec{Selector: "#cta"}, []ExtractorSpec{
		{Kind: ExtractorExists},
		{Kind: ExtractorVisible},
		{Kind: ExtractorText, Text: TextOptions{NormalizeWhitespace: true, Trim: true}},
		{Kind: ExtractorBounds},
		{Kind: ExtractorComputedStyle, Props: []string{"color"}},
		{Kind: ExtractorAttributes, Attributes: []string{"id", "class"}},
	})
	require.NoError(t, err)
	require.Equal(t, "#cta", snapshot.Selector)
	require.NotNil(t, snapshot.Exists)
	require.True(t, *snapshot.Exists)
	require.NotNil(t, snapshot.Visible)
	require.True(t, *snapshot.Visible)
	require.Equal(t, "Book now", snapshot.Text)
	require.NotNil(t, snapshot.Bounds)
	require.Equal(t, "rgb(255, 0, 0)", snapshot.Computed["color"])
	require.Equal(t, "cta", snapshot.Attributes["id"])
	require.Equal(t, "primary", snapshot.Attributes["class"])
}

func TestExtractElementMissingAndInvalidSelectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta">Book</button></body></html>`)
	}))
	defer server.Close()

	page := newExtractTestPage(t, server.URL)
	missing, err := ExtractElement(page, LocatorSpec{Selector: "#missing"}, []ExtractorSpec{
		{Kind: ExtractorExists},
		{Kind: ExtractorText, Text: TextOptions{NormalizeWhitespace: true, Trim: true}},
	})
	require.NoError(t, err)
	require.NotNil(t, missing.Exists)
	require.False(t, *missing.Exists)
	require.Empty(t, missing.Text)

	_, err = ExtractElement(page, LocatorSpec{Selector: "#bad["}, []ExtractorSpec{{Kind: ExtractorExists}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "selector status")
}

func newExtractTestPage(t *testing.T, url string) *driver.Page {
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
