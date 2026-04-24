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

func TestPreflightProbesReportsExistingMissingAndInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta">Book now</button><div id="hidden" style="display:none">Hidden</div></body></html>`)
	}))
	defer server.Close()

	browser, err := driver.NewBrowser(context.Background())
	require.NoError(t, err)
	defer browser.Close()

	page, err := browser.NewPage()
	require.NoError(t, err)
	defer page.Close()

	require.NoError(t, page.SetViewport(400, 300))
	require.NoError(t, page.Goto(server.URL))

	statuses, err := PreflightProbes(page, []ProbeSpec{
		{Name: "button", Selector: "#cta", Source: "test"},
		{Name: "missing", Selector: "#does-not-exist", Source: "test"},
		{Name: "invalid", Selector: "#bad[", Source: "test"},
		{Name: "hidden", Selector: "#hidden", Source: "test"},
	})
	require.NoError(t, err)
	require.Len(t, statuses, 4)

	require.True(t, statuses[0].Exists)
	require.True(t, statuses[0].Visible)
	require.NotNil(t, statuses[0].Bounds)
	require.Contains(t, statuses[0].TextStart, "Book now")

	require.False(t, statuses[1].Exists)
	require.False(t, statuses[1].Visible)
	require.Empty(t, statuses[1].Error)

	require.False(t, statuses[2].Exists)
	require.NotEmpty(t, statuses[2].Error)

	require.True(t, statuses[3].Exists)
	require.False(t, statuses[3].Visible)
}
