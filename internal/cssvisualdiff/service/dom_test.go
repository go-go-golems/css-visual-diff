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

func TestLocatorDOMPrimitives(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>
<button id="cta" class="primary" data-kind="booking" style="color: rgb(255, 0, 0)">  Book
 now  </button>
<div id="hidden" style="display:none">Hidden</div>
<section id="panel"><strong>Inner</strong></section>
</body></html>`)
	}))
	defer server.Close()

	page := newDOMTestPage(t, server.URL)

	status, err := LocatorStatus(page, LocatorSpec{Name: "cta", Selector: "#cta", Source: "test"})
	require.NoError(t, err)
	require.Equal(t, "cta", status.Name)
	require.Equal(t, "#cta", status.Selector)
	require.True(t, status.Exists)
	require.True(t, status.Visible)
	require.NotNil(t, status.Bounds)
	require.Contains(t, status.TextStart, "Book now")

	rawText, err := LocatorText(page, LocatorSpec{Selector: "#cta"}, TextOptions{})
	require.NoError(t, err)
	require.Contains(t, rawText, "Book")
	require.Contains(t, rawText, "now")

	normalizedText, err := LocatorText(page, LocatorSpec{Selector: "#cta"}, TextOptions{NormalizeWhitespace: true, Trim: true})
	require.NoError(t, err)
	require.Equal(t, "Book now", normalizedText)

	innerHTML, err := LocatorHTML(page, LocatorSpec{Selector: "#panel"}, false)
	require.NoError(t, err)
	require.True(t, innerHTML.Exists)
	require.Equal(t, "<strong>Inner</strong>", innerHTML.HTML)

	outerHTML, err := LocatorHTML(page, LocatorSpec{Selector: "#panel"}, true)
	require.NoError(t, err)
	require.True(t, outerHTML.Exists)
	require.Contains(t, outerHTML.HTML, `<section id="panel">`)

	bounds, err := LocatorBounds(page, LocatorSpec{Selector: "#cta"})
	require.NoError(t, err)
	require.NotNil(t, bounds)
	require.Greater(t, bounds.Width, 0.0)
	require.Greater(t, bounds.Height, 0.0)

	attrs, err := LocatorAttributes(page, LocatorSpec{Selector: "#cta"}, []string{"id", "class", "data-kind", "missing"})
	require.NoError(t, err)
	require.Equal(t, "cta", attrs["id"])
	require.Equal(t, "primary", attrs["class"])
	require.Equal(t, "booking", attrs["data-kind"])
	require.Equal(t, "", attrs["missing"])

	styles, err := LocatorComputedStyle(page, LocatorSpec{Selector: "#cta"}, []string{"color", "display"})
	require.NoError(t, err)
	require.Equal(t, "rgb(255, 0, 0)", styles["color"])
	require.NotEmpty(t, styles["display"])
}

func TestWaitForLocator(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>
<div id="hidden" style="display:none">Hidden</div>
<script>
setTimeout(() => {
  const delayed = document.createElement('div');
  delayed.id = 'delayed';
  delayed.textContent = 'Ready';
  document.body.appendChild(delayed);
}, 150);
setTimeout(() => {
  document.getElementById('hidden').style.display = 'block';
}, 250);
</script>
</body></html>`)
	}))
	defer server.Close()

	page := newDOMTestPage(t, server.URL)

	existing, err := WaitForLocator(page, LocatorSpec{Selector: "body"}, WaitForSelectorOptions{TimeoutMS: 1000})
	require.NoError(t, err)
	require.True(t, existing.Exists)
	require.True(t, existing.Visible)
	require.Equal(t, "body", existing.Selector)

	delayed, err := WaitForLocator(page, LocatorSpec{Selector: "#delayed"}, WaitForSelectorOptions{TimeoutMS: 2000, PollIntervalMS: 50})
	require.NoError(t, err)
	require.True(t, delayed.Exists)
	require.True(t, delayed.Visible)
	require.Contains(t, delayed.TextStart, "Ready")

	visible, err := WaitForLocator(page, LocatorSpec{Selector: "#hidden"}, WaitForSelectorOptions{TimeoutMS: 2000, PollIntervalMS: 50, Visible: true})
	require.NoError(t, err)
	require.True(t, visible.Exists)
	require.True(t, visible.Visible)

	missing, err := WaitForLocator(page, LocatorSpec{Selector: "#missing"}, WaitForSelectorOptions{TimeoutMS: 100, PollIntervalMS: 25})
	require.Error(t, err)
	require.Contains(t, err.Error(), "did not become present")
	require.False(t, missing.Exists)

	_, err = WaitForLocator(page, LocatorSpec{Selector: "#bad["}, WaitForSelectorOptions{TimeoutMS: 100})
	require.Error(t, err)
	require.Contains(t, err.Error(), "selector")
}

func TestLocatorDOMPrimitivesMissingHiddenAndInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><div id="hidden" style="display:none">Hidden</div></body></html>`)
	}))
	defer server.Close()

	page := newDOMTestPage(t, server.URL)

	missingStatus, err := LocatorStatus(page, LocatorSpec{Name: "missing", Selector: "#missing"})
	require.NoError(t, err)
	require.False(t, missingStatus.Exists)
	require.False(t, missingStatus.Visible)
	require.Empty(t, missingStatus.Error)

	missingText, err := LocatorText(page, LocatorSpec{Selector: "#missing"}, TextOptions{NormalizeWhitespace: true, Trim: true})
	require.NoError(t, err)
	require.Empty(t, missingText)

	missingHTML, err := LocatorHTML(page, LocatorSpec{Selector: "#missing"}, true)
	require.NoError(t, err)
	require.False(t, missingHTML.Exists)
	require.Empty(t, missingHTML.HTML)

	missingBounds, err := LocatorBounds(page, LocatorSpec{Selector: "#missing"})
	require.NoError(t, err)
	require.Nil(t, missingBounds)

	missingAttrs, err := LocatorAttributes(page, LocatorSpec{Selector: "#missing"}, []string{"id"})
	require.NoError(t, err)
	require.Empty(t, missingAttrs)

	missingStyles, err := LocatorComputedStyle(page, LocatorSpec{Selector: "#missing"}, []string{"color"})
	require.NoError(t, err)
	require.Empty(t, missingStyles)

	hiddenStatus, err := LocatorStatus(page, LocatorSpec{Name: "hidden", Selector: "#hidden"})
	require.NoError(t, err)
	require.True(t, hiddenStatus.Exists)
	require.False(t, hiddenStatus.Visible)

	invalidStatus, err := LocatorStatus(page, LocatorSpec{Name: "invalid", Selector: "#bad["})
	require.NoError(t, err)
	require.False(t, invalidStatus.Exists)
	require.NotEmpty(t, invalidStatus.Error)

	_, err = LocatorText(page, LocatorSpec{Selector: "#bad["}, TextOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "evaluate locator text")

	_, err = LocatorHTML(page, LocatorSpec{Selector: "#bad["}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "evaluate locator html")

	_, err = LocatorBounds(page, LocatorSpec{Selector: "#bad["})
	require.Error(t, err)
	require.Contains(t, err.Error(), "evaluate locator bounds")

	_, err = LocatorAttributes(page, LocatorSpec{Selector: "#bad["}, []string{"id"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "evaluate locator attributes")

	_, err = LocatorComputedStyle(page, LocatorSpec{Selector: "#bad["}, []string{"color"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "evaluate locator computed style")
}

func newDOMTestPage(t *testing.T, url string) *driver.Page {
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
