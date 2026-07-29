package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInteractiveScatterDocumentationCoversPinnedSourceFamily(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/scatter", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET interactive Scatter status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Interactive scatter", "Purpose", "Use when", "Avoid when", "Equivalent data",
		"Basic sports scatter", "Swimming", "Surfing", "Shooting", "Skating", "Wrestling", "Diving",
		"Category A", "Category B", "roundRect", `"symbolSize":20`, `"symbolRotate":10`,
		"Visible point labels", `"position":"right"`, "Named axes and split lines", `"name":"Sports"`, `"name":"Score"`,
		"Effect variant", `"type":"effectScatter"`, "Kobe", "Jordan", "Iverson", "LeBron", "Wade", "McGrady",
		"Per-series ripple styles", `"period":4`, `"scale":10`, `"brushType":"stroke"`, `"period":3`, `"scale":6`, `"brushType":"fill"`,
		"Exact scatter values", "12 points across 2 series", "Motion and emphasis", "reduced-motion",
		`data-scatter-variant="base"`, `data-scatter-variant="labels"`, `data-scatter-variant="split-lines"`, `data-scatter-variant="effect-base"`, `data-scatter-variant="effect-styles"`,
		`data-go-api-version="v0.0.1"`, "chart controls", "static/vector and interactive capabilities",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("interactive Scatter documentation missing %q", want)
		}
	}
	for _, unwanted := range []string{"go-echarts", "ECharts", "examples/scatter.go", "examples/effectscatter.go", "dataset transform"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("interactive Scatter component page contains private or out-of-scope framing %q", unwanted)
		}
	}
}

func TestInteractiveScatterAttributionCentralizesPinnedEvidence(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"examples/scatter.go", "examples/effectscatter.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd",
		"deterministic values", "docs/upstream-example-coverage.md",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("central attribution missing Scatter evidence %q", want)
		}
	}
}
