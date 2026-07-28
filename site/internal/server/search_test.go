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
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/heartbeat", nil))

	body := recorder.Body.String()
	for _, want := range []string{
		`id="docs-search"`,
		`role="combobox"`,
		`data-docs-search-results`,
		`data-search="bar chart server-rendered bar"`,
		`data-search="bar interactive / cartesian echarts-bar"`,
		`data-search="scatter interactive / cartesian echarts-scatter"`,
		`data-search="effect scatter interactive / cartesian echarts-effect-scatter"`,
		`data-search="attributions general attributions"`,
		`href="/attributions"`,
		`href="/components/echarts/line"`,
		`hx-get="/components/echarts/line"`,
		`src="/search/assets/search.js"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("documentation search missing %q", want)
		}
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
