package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chartassets "github.com/araihu/goshtoso-charts/assets"
)

func TestDemoRoutesRender(t *testing.T) {
	t.Parallel()
	handler := New()
	for _, test := range []struct {
		path string
		want string
	}{
		{"/", "Charts for Goshtoso"},
		{"/attributions", "Backing libraries"},
		{"/components/heartbeat", "Heartbeat"},
		{"/components/line", "Line chart"},
		{"/components/bar", "Bar chart"},
		{"/components/pie", "Pie chart"},
		{"/components/interactive/bar", "Interactive bar"},
		{"/components/interactive/line", "Interactive line"},
		{"/components/interactive/scatter", "Interactive scatter"},
		{"/components/interactive/pie", "Interactive pie"},
		{"/components/interactive/radar", "Interactive radar"},
		{"/components/interactive/heatmap", "Interactive heatmap"},
		{"/components/interactive/boxplot", "Interactive box plot"},
		{"/components/interactive/gauge", "Interactive gauge"},
		{"/components/interactive/funnel", "Interactive funnel"},
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

func TestAttributionsCentralizeBackingLibraryCredits(t *testing.T) {
	t.Parallel()
	handler := New()

	attributions := httptest.NewRecorder()
	handler.ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	if attributions.Code != http.StatusOK {
		t.Fatalf("GET /attributions status = %d, want %d", attributions.Code, http.StatusOK)
	}
	for _, want := range []string{"Goshtoso App Shells", "templ", "go-analyze/charts", "go-echarts", "Apache ECharts", ">5.4.3<", "MIT license", "Apache-2.0 license", ">assets/NOTICE.md<"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("attributions page missing %q", want)
		}
	}
	if strings.Contains(attributions.Body.String(), "optional extensions") {
		t.Error("attributions page claims removed optional extensions are still pinned")
	}

	for _, path := range []string{"/components/interactive/bar", "/components/interactive/line", "/components/interactive/scatter", "/components/interactive/pie", "/components/interactive/radar", "/components/interactive/heatmap", "/components/interactive/boxplot", "/components/interactive/gauge", "/components/interactive/funnel"} {
		page := httptest.NewRecorder()
		handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, path, nil))
		for _, unwanted := range []string{"backed by go-echarts", "with go-echarts options", ">go-echarts catalog<"} {
			if strings.Contains(page.Body.String(), unwanted) {
				t.Errorf("GET %s repeats centralized attribution %q", path, unwanted)
			}
		}
	}
}

func TestEffectScatterDocumentationRedirectsToUnifiedScatter(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/effect-scatter", nil))
	if recorder.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusPermanentRedirect)
	}
	if location := recorder.Header().Get("Location"); location != "/components/interactive/scatter" {
		t.Fatalf("Location = %q, want unified scatter route", location)
	}
}

func TestEngineNamedComponentRoutesRedirectToPublicInteractiveRoutes(t *testing.T) {
	t.Parallel()
	for _, component := range []string{"bar", "line", "scatter", "pie", "radar", "heatmap", "boxplot", "gauge", "funnel"} {
		recorder := httptest.NewRecorder()
		New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/echarts/"+component, nil))
		if recorder.Code != http.StatusPermanentRedirect {
			t.Errorf("GET legacy %s status = %d, want %d", component, recorder.Code, http.StatusPermanentRedirect)
		}
		if location := recorder.Header().Get("Location"); location != "/components/interactive/"+component {
			t.Errorf("GET legacy %s Location = %q", component, location)
		}
	}
}

func TestOverviewAndAttributionsUseNativeSidebarIcons(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	body := recorder.Body.String()
	for _, want := range []string{`data-sidebar-icon="overview"`, `data-sidebar-icon="attributions"`, `aria-hidden="true"`} {
		if !strings.Contains(body, want) {
			t.Errorf("sidebar navigation missing icon contract %q", want)
		}
	}
	if !strings.Contains(body, `href="/attributions"`) || !strings.Contains(body, `aria-current="page"`) {
		t.Error("attributions navigation does not retain linked, active accessible state")
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

func TestInteractiveRendererRuntimeIsLocal(t *testing.T) {
	t.Parallel()
	handler := New()

	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/components/interactive/bar", nil))
	if !strings.Contains(page.Body.String(), `src="`+chartassets.RuntimeURL+`"`) {
		t.Fatalf("interactive page missing public local dependency tag %q", chartassets.RuntimeURL)
	}
	if strings.Contains(page.Body.String(), "cdn.jsdelivr.net") {
		t.Fatal("demo opted into CDN runtime")
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, chartassets.RuntimeURL, nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `version="5.4.3"`) {
		t.Fatalf("GET modern runtime status/version = %d/%t", recorder.Code, strings.Contains(recorder.Body.String(), `version="5.4.3"`))
	}

	removed := httptest.NewRecorder()
	handler.ServeHTTP(removed, httptest.NewRequest(http.MethodGet, "/charts/echarts/echarts@4.min.js", nil))
	if removed.Code != http.StatusNotFound {
		t.Fatalf("GET removed catalog extension status = %d, want %d", removed.Code, http.StatusNotFound)
	}
}

func TestComponentDocsNavigationHasSearchGroupsAndComponentContract(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/heartbeat", nil))
	body := recorder.Body.String()
	for _, want := range []string{"Search docs...", "Static / Vector", "Interactive / Cartesian", "Examples", "component-doc-shell__sidebar", "components.KindHeartbeat", "Bar chart", "Pie chart", "Accessibility", "lg:grid-cols-2"} {
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

func TestAraiHuThemeIsCurrentAndDefault(t *testing.T) {
	t.Parallel()
	handler := New()

	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	for _, want := range []string{`"theme":"araihu"`, `/componentdocshell/assets/araihu.css`} {
		if !strings.Contains(page.Body.String(), want) {
			t.Errorf("charts demo default theme missing %q", want)
		}
	}

	theme := httptest.NewRecorder()
	handler.ServeHTTP(theme, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/araihu.css", nil))
	for _, want := range []string{`--araihu-logo-surface`, `--araihu-logo-ink`, `--araihu-logo-signal`, `.dark [data-theme="araihu"]`} {
		if !strings.Contains(theme.Body.String(), want) {
			t.Errorf("charts demo theme missing V11 contract %q", want)
		}
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
