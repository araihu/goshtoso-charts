package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInteractiveFunnelDocumentationCoversPinnedSourceFamily(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/funnel", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET interactive Funnel status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Interactive funnel", "Purpose", "Use when", "Avoid when", "Equivalent data",
		"Basic five-stage funnel", "basic funnel example", "Analytics", "Visit", "Add", "Order", "Payment", "Deal",
		`"value":31`, `"value":37`, `"value":47`, `"value":9`,
		"Visible labels on the left", "show label", `"value":18`, `"value":25`, `"value":40`, `"value":6`, `"value":0`, `"show":true`, `"position":"left"`,
		"Exact funnel values", `data-funnel-variant="base"`, `data-funnel-variant="labels-left"`,
		"upstream source sequence in the exact table", "upstream default descending-by-value order",
		"Stage meaning", "FunnelOrderData", "chart controls", "static/vector and interactive capabilities",
		`data-go-api-version="v0.0.1"`, "Open v0.0.1 API", "github.com/araihu/goshtoso-charts/components/interactive",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("interactive Funnel documentation missing %q", want)
		}
	}
	for _, unwanted := range []string{"PRIMITIVE", "KIND", "CONFIGURATION", "Component contract", "go-echarts", "ECharts", "examples/funnel.go", "release pipeline", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("interactive Funnel component page contains private or low-value framing %q", unwanted)
		}
	}
}

func TestInteractiveFunnelAttributionCentralizesPinnedEvidence(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	body := recorder.Body.String()
	for _, want := range []string{"examples/funnel.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd", "both dedicated behaviors", "[0,50) value domain", "docs/upstream-example-coverage.md"} {
		if !strings.Contains(body, want) {
			t.Errorf("central attribution missing Funnel evidence %q", want)
		}
	}
}
