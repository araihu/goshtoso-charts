package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLinePageExhaustsPinnedStaticExamplesWithoutRendererBranding(t *testing.T) {
	t.Parallel()
	page := httptest.NewRecorder()
	New().ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/components/line", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("GET Line status = %d", page.Code)
	}
	body := page.Body.String()
	for _, want := range []string{
		"Missing observations and basic presentation", "Basic line chart with one missing Email observation", "Unavailable",
		"Per-series symbols", "Line series with distinct point symbols",
		"Smoothed strokes", "Bold smoothed line series",
		"Statistical references", "Average", "Maximum", "Reference evidence",
		"Filled area treatment", "Dual Y-axis treatment", "Presentation overrides",
		"Stacked contributions", "A+B+C Sum", "Max:",
		"Boundary-gap comparison", "Boundary Gap", "Boundary Gap Disabled",
		"Dense axes, gaps, and positioned labels", "Canon RF Zoom Lenses", "100-500mm f/4.5-7.1", "60mm", "510mm",
		"Theme-aware value-label scale", "Cold = Low Values, Warm = High Values", "Sales Performance",
		"Expand", "SVG", "PNG", "Go API", "v0.0.1",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("Line page missing %q", want)
		}
	}
	for _, unwanted := range []string{"go-analyze", "line_chart-", "Painter", "raw option", "raw map", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("Line page contains renderer or domain leakage %q", unwanted)
		}
	}
}
