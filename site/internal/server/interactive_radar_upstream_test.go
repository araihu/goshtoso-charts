package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInteractiveRadarDocumentationCoversPinnedSourceFamily(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/radar", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET interactive Radar status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Interactive radar", "Purpose", "Use when", "Avoid when", "Equivalent data",
		"Daily Beijing air quality", "AQI", "PM2.5", "PM10", "CO", "NO2", "SO2", "Day 21",
		"Circular style", `"shape":"circle"`, `"splitNumber":5`,
		"Multiple-series legend", "Beijing", "Guangzhou", "Shanghai", `"selectedMode":"multiple"`,
		"Single-series legend", `"selectedMode":"single"`,
		"Exact radar values", `data-radar-variant="base"`, `data-radar-variant="style"`, `data-radar-variant="legend-multiple"`, `data-radar-variant="legend-single"`,
		"When radar works", "When radar fails", "Shared controls and modes", "chart controls", "static/vector and interactive capabilities",
		`data-go-api-version="v0.0.1"`, "Open v0.0.1 API", "github.com/araihu/goshtoso-charts/components/interactive/radar",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("interactive Radar documentation missing %q", want)
		}
	}
	for _, unwanted := range []string{"PRIMITIVE", "KIND", "CONFIGURATION", "Component contract", "go-echarts", "ECharts", "examples/radar.go"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("interactive Radar component page contains private or low-value framing %q", unwanted)
		}
	}
}

func TestInteractiveRadarAttributionCentralizesPinnedEvidence(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	body := recorder.Body.String()
	for _, want := range []string{"examples/radar.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd", "six pollutant dimensions", "docs/upstream-example-coverage.md"} {
		if !strings.Contains(body, want) {
			t.Errorf("central attribution missing Radar evidence %q", want)
		}
	}
}
