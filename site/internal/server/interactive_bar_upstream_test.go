package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInteractiveBarPageExhaustsPinnedExamplesWithoutRendererBranding(t *testing.T) {
	t.Parallel()
	page := httptest.NewRecorder()
	New().ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/components/interactive/bar", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("GET interactive Bar status = %d", page.Code)
	}
	body := page.Body.String()
	for _, want := range []string{
		"Basic bar example", "This is the subtitle.", "Exact category values",
		"Labels", "Visible value labels",
		"Axes and units", "Named axes with literal units",
		"Explicit colors", "Caller-selected colors override theme series tokens",
		"Series arrangement", "Bar widths and gap", "Horizontal bar orientation", "Stacked bar series",
		"Category zoom", "Inside category zoom", "Slider category zoom",
		"Calculated and explicit references", "Bar point references", "Bar guide references",
		"Large canvas", "600-pixel height",
		"Shared chart behavior", "/docs/chart-controls", "/docs/chart-modes",
		"Mixed-series composition", "renderer-neutral composite chart API", "Go API", "v0.0.1",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("interactive Bar page missing %q", want)
		}
	}
	if count := strings.Count(body, `data-bar-exact-values`); count != 12 {
		t.Errorf("exact-value disclosure count = %d, want 12", count)
	}
	for _, unwanted := range []string{"go-echarts", "apache echarts", "examples/bar.go", "raw map", "infrastructure", "operations"} {
		if strings.Contains(strings.ToLower(body), unwanted) {
			t.Errorf("interactive Bar page contains renderer or domain leakage %q", unwanted)
		}
	}
}
