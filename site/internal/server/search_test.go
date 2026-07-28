package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocumentationSearchIndexesCategorizedNavigation(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/line", nil))

	body := recorder.Body.String()
	for _, want := range []string{
		`id="docs-search"`,
		`role="combobox"`,
		`data-docs-search-results`,
		`data-search="getting started general getting-started"`,
		`data-search="bar chart static / vector bar"`,
		`data-search="radar chart static / vector radar"`,
		`data-search="bar interactive / cartesian interactive-bar"`,
		`data-search="scatter interactive / cartesian interactive-scatter"`,
		`data-search="pie interactive / statistical interactive-pie"`,
		`data-search="radar interactive / statistical interactive-radar"`,
		`data-search="heatmap interactive / cartesian interactive-heatmap"`,
		`data-search="box plot interactive / statistical interactive-boxplot"`,
		`data-search="gauge interactive / statistical interactive-gauge"`,
		`data-search="funnel interactive / statistical interactive-funnel"`,
		`data-search="graph interactive / relationships interactive-graph"`,
		`data-search="sankey interactive / relationships interactive-sankey"`,
		`data-search="live availability examples live-availability"`,
		`data-search="attributions general attributions"`,
		`href="/attributions"`,
		`href="/components/interactive/line"`,
		`hx-get="/components/interactive/line"`,
		`src="/search/assets/search.js"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("documentation search missing %q", want)
		}
	}
	if strings.Contains(body, `data-search="heartbeat static / vector heartbeat"`) {
		t.Error("documentation search still indexes heartbeat as a component")
	}
	if strings.Contains(body, `data-search="status page examples status-page"`) || strings.Contains(body, `href="/examples/status-page"`) {
		t.Error("documentation search still indexes the removed status page")
	}
}

func TestDocumentationSearchRuntimeIsLocalAndInteractive(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/search/assets/search.js", nil))

	body := recorder.Body.String()
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET search runtime status = %d, want 200", recorder.Code)
	}
	for _, want := range []string{"data-docs-search-item", "terms.every", "ArrowDown", "Enter", "aria-expanded"} {
		if !strings.Contains(body, want) {
			t.Errorf("search runtime missing %q", want)
		}
	}
	if strings.Contains(body, "https://") {
		t.Error("search runtime references a hosted service")
	}
}
