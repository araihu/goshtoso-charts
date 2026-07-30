package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInteractiveHeatMapDocumentationCoversPinnedCategoryAndCalendarExamples(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/heatmap", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Weekly activity by hour", "Calendar coordinates", "Calendar activity",
		"Saturday", "11p", "Exact heatmap values", "No data",
		"Cold to warm", "Missing is not zero", "Equivalent data",
		"/docs/chart-controls", "/docs/chart-modes", "Go API",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("interactive heatmap page missing %q", want)
		}
	}
	for _, unwanted := range []string{
		"Component contract", "PRIMITIVE", "KIND", "CONFIGURATION",
		"go-echarts", "Apache ECharts", "examples/heatmap.go", "infrastructure", "operations", "Deployment activity",
	} {
		if strings.Contains(body, unwanted) {
			t.Errorf("interactive heatmap page exposes unwanted %q", unwanted)
		}
	}
}

func TestInteractiveHeatMapCoverageIsCentralized(t *testing.T) {
	t.Parallel()
	for _, route := range []string{"/attributions", "/components/interactive/heatmap"} {
		recorder := httptest.NewRecorder()
		New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d", route, recorder.Code)
		}
		body := recorder.Body.String()
		if route == "/attributions" {
			for _, want := range []string{"examples/heatmap.go", "both dedicated behaviors", "168 category cells", "366-day calendar", "docs/upstream-example-coverage.md"} {
				if !strings.Contains(body, want) {
					t.Errorf("attributions missing %q", want)
				}
			}
			continue
		}
		if strings.Contains(body, "examples/heatmap.go") {
			t.Error("component page repeats centralized backing-source path")
		}
	}
}
