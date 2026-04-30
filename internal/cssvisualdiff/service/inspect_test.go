package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/stretchr/testify/require"
)

func TestInspectPreparedPageRejectsOutputFileWithMultipleRequests(t *testing.T) {
	_, err := InspectPreparedPage(nil, PageTarget{}, "script", []InspectRequest{
		{Name: "one", Selector: "#one"},
		{Name: "two", Selector: "#two"},
	}, InspectAllOptions{OutDir: t.TempDir(), OutputFile: filepath.Join(t.TempDir(), "computed-css.json"), Format: InspectFormatCSSJSON})
	require.Error(t, err)
	require.Contains(t, err.Error(), "outputFile requires exactly one inspect request")
}

func TestInspectPreparedPageWritesArtifactsWithoutReloadingPerProbe(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			atomic.AddInt32(&hits, 1)
		}
		_, _ = fmt.Fprint(w, `<html><body><div id="one" style="display:block;color:red">One</div><div id="two" style="display:block;color:blue">Two</div></body></html>`)
	}))
	defer server.Close()

	browser, err := driver.NewBrowser(context.Background())
	require.NoError(t, err)
	defer browser.Close()

	page, err := browser.NewPage()
	require.NoError(t, err)
	defer page.Close()

	target := PageTarget{
		Name:     "fixture",
		URL:      server.URL,
		Viewport: Viewport{Width: 400, Height: 300},
	}
	require.NoError(t, LoadAndPreparePage(page, target))
	require.Equal(t, int32(1), atomic.LoadInt32(&hits))

	outDir := t.TempDir()
	result, err := InspectPreparedPage(page, target, "original", []InspectRequest{
		{Name: "one", Selector: "#one", Props: []string{"color"}, Source: "test"},
		{Name: "two", Selector: "#two", Props: []string{"color"}, Source: "test"},
	}, InspectAllOptions{OutDir: outDir, Format: InspectFormatCSSJSON})
	require.NoError(t, err)
	require.Len(t, result.Results, 2)
	require.Equal(t, int32(1), atomic.LoadInt32(&hits), "inspect should reuse the prepared page instead of reloading per probe")

	for _, name := range []string{"one", "two"} {
		path := filepath.Join(outDir, name, "computed-css.json")
		_, err := os.Stat(path)
		require.NoError(t, err)
	}
	_, err = os.Stat(filepath.Join(outDir, "index.json"))
	require.NoError(t, err)
}
