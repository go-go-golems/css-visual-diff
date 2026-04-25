package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/stretchr/testify/require"
)

func TestCollectSelectionProfilesMissingAllAndJSON(t *testing.T) {
	page := newCollectionTestPage(t, `<html><body>
<button id="cta" class="primary" data-kind="booking" role="button" style="color: rgb(255, 0, 0); padding: 4px">  Book
 now  </button>
<div id="card" class="panel" data-kind="summary" style="color: rgb(1, 2, 3); margin-top: 7px">Card</div>
</body></html>`)

	minimal, err := CollectSelection(page, LocatorSpec{Name: "cta", Selector: "#cta", Source: "test"}, CollectOptions{Inspect: CollectInspectMinimal})
	require.NoError(t, err)
	require.Equal(t, CollectedSelectionSchemaVersion, minimal.SchemaVersion)
	require.Equal(t, "cta", minimal.Name)
	require.Equal(t, "#cta", minimal.Selector)
	require.Equal(t, "test", minimal.Source)
	require.NotEmpty(t, minimal.URL)
	require.True(t, minimal.Exists)
	require.True(t, minimal.Visible)
	require.NotNil(t, minimal.Bounds)
	require.Empty(t, minimal.Text)
	require.Empty(t, minimal.ComputedStyles)
	require.Empty(t, minimal.Attributes)

	rich, err := CollectSelection(page, LocatorSpec{Name: "cta", Selector: "#cta"}, CollectOptions{Inspect: CollectInspectRich})
	require.NoError(t, err)
	require.True(t, rich.Exists)
	require.True(t, rich.Visible)
	require.NotNil(t, rich.Bounds)
	require.Equal(t, "Book now", rich.Text)
	require.Equal(t, "cta", rich.Attributes["id"])
	require.Equal(t, "primary", rich.Attributes["class"])
	require.Equal(t, "button", rich.Attributes["role"])
	require.Equal(t, "rgb(255, 0, 0)", rich.ComputedStyles["color"])
	require.NotEmpty(t, rich.ComputedStyles["display"])

	missing, err := CollectSelection(page, LocatorSpec{Name: "missing", Selector: "#missing"}, CollectOptions{Inspect: CollectInspectRich})
	require.NoError(t, err)
	require.Equal(t, CollectedSelectionSchemaVersion, missing.SchemaVersion)
	require.False(t, missing.Exists)
	require.False(t, missing.Visible)
	require.Empty(t, missing.Text)
	require.Empty(t, missing.ComputedStyles)
	require.Empty(t, missing.Attributes)

	all, err := CollectSelection(page, LocatorSpec{Name: "card", Selector: "#card"}, CollectOptions{
		Inspect:       CollectInspectRich,
		AllStyles:     true,
		AllAttributes: true,
		IncludeHTML:   true,
		OuterHTML:     true,
	})
	require.NoError(t, err)
	require.Equal(t, "card", all.Attributes["id"])
	require.Equal(t, "panel", all.Attributes["class"])
	require.Equal(t, "summary", all.Attributes["data-kind"])
	require.Contains(t, all.HTML, `id="card"`)
	require.Contains(t, all.ComputedStyles, "color")
	require.Contains(t, all.ComputedStyles, "display")
	require.NotEmpty(t, all.ComputedStyles["color"])

	payload, err := json.Marshal(all)
	require.NoError(t, err)
	require.Contains(t, string(payload), `"schemaVersion":"cssvd.collectedSelection.v1"`)
	require.Contains(t, string(payload), `"computedStyles"`)

	var roundTrip CollectedSelectionData
	require.NoError(t, json.Unmarshal(payload, &roundTrip))
	require.Equal(t, all.SchemaVersion, roundTrip.SchemaVersion)
	require.Equal(t, all.Selector, roundTrip.Selector)
}

func TestCollectSelectionInvalidSelector(t *testing.T) {
	page := newCollectionTestPage(t, `<html><body><div id="exists">Exists</div></body></html>`)

	_, err := CollectSelection(page, LocatorSpec{Name: "bad", Selector: "#bad["}, CollectOptions{Inspect: CollectInspectRich})
	require.Error(t, err)
	var collectionErr *CollectionError
	require.True(t, errors.As(err, &collectionErr))
	require.Equal(t, CollectionErrorInvalidSelector, collectionErr.Kind)
	require.Equal(t, "status", collectionErr.Op)
}

func newCollectionTestPage(t *testing.T, html string) *driver.Page {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, html)
	}))
	t.Cleanup(server.Close)
	return newDOMTestPage(t, server.URL)
}
