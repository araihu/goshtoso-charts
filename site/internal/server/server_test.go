package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDemoRoutesRender(t *testing.T) {
	t.Parallel()
	handler := New()
	for _, test := range []struct {
		path string
		want string
	}{
		{"/", "SSR charts for Goshtoso"},
		{"/components/heartbeat", "Heartbeat"},
		{"/components/line", "Line chart"},
		{"/examples/status-page", "Status page example"},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", test.path, recorder.Code, http.StatusOK)
		}
		if !strings.Contains(recorder.Body.String(), test.want) {
			t.Errorf("GET %s missing %q", test.path, test.want)
		}
	}
}

func TestAssetsAreMountedWithoutStripPrefix(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/styles.css", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /assets/styles.css status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestCatalogNavigationHasSearchGroupsAndComponentContract(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/heartbeat", nil))
	body := recorder.Body.String()
	for _, want := range []string{"Search catalog", "Components", "Examples", "x-model=\"query\"", "components.KindHeartbeat", "Accessibility"} {
		if !strings.Contains(body, want) {
			t.Errorf("catalog page missing %q", want)
		}
	}
}

func TestDemoStylesheetRenders(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/site.css", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), ".charts-shell") {
		t.Fatalf("GET /site.css status/body = %d/%q", recorder.Code, recorder.Body.String())
	}
}
