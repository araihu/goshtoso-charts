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
		{"/components/bar", "Bar chart"},
		{"/components/pie", "Pie chart"},
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

func TestV11GoshtosoBrandAssetsAndMetadataRender(t *testing.T) {
	t.Parallel()
	handler := New()
	for _, test := range []struct {
		path string
		want string
	}{
		{"/brand/goshtoso-logo-transparent.svg", "class=\"araihu-brand-v11\""},
		{"/brand/goshtoso-icon-transparent.svg", "class=\"araihu-brand-v11\""},
		{"/", `<link rel="icon" href="/brand/goshtoso-icon-transparent.svg">`},
		{"/", `<title>Overview · Goshtoso Charts</title>`},
		{"/", `aria-label="Goshtoso Charts"`},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), test.want) {
			t.Errorf("GET %s status/body = %d/%q, want %q", test.path, recorder.Code, recorder.Body.String(), test.want)
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

func TestComponentDocsNavigationHasSearchGroupsAndComponentContract(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/heartbeat", nil))
	body := recorder.Body.String()
	for _, want := range []string{"Search docs...", "Components", "Examples", "component-doc-shell__sidebar", "components.KindHeartbeat", "Bar chart", "Pie chart", "Accessibility", "lg:grid-cols-2"} {
		if !strings.Contains(body, want) {
			t.Errorf("component docs page missing %q", want)
		}
	}
	if strings.Contains(body, "sm:grid-cols-2") {
		t.Error("component contract switches to two columns before fixed-sidebar content has room")
	}
}

func TestComponentDocsShellAssetsRender(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/shell.css", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), ".component-doc-shell__frame") {
		t.Fatalf("GET shell stylesheet status/body = %d/%q", recorder.Code, recorder.Body.String())
	}
}

func TestHTMXNavigationRendersContentAndSidebarFragment(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "/components/line", nil)
	request.Header.Set("HX-Request", "true")
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, request)
	body := recorder.Body.String()
	if strings.Contains(body, "<html") {
		t.Fatal("HTMX response contains a complete document")
	}
	for _, want := range []string{`<title>Line chart · Goshtoso Charts</title>`, `id="main-content"`, `id="componentdocshell-sidebar-content"`, `hx-swap-oob`, `components.KindLineChart`} {
		if !strings.Contains(body, want) {
			t.Errorf("HTMX response missing %q", want)
		}
	}
}
