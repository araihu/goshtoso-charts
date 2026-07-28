package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/site/internal/echartsexamples"
)

func TestDemoRoutesRender(t *testing.T) {
	t.Parallel()
	handler := New()
	for _, test := range []struct {
		path string
		want string
	}{
		{"/", "SSR charts for Goshtoso"},
		{"/attributions", "Backing libraries"},
		{"/components/heartbeat", "Heartbeat"},
		{"/components/line", "Line chart"},
		{"/components/bar", "Bar chart"},
		{"/components/pie", "Pie chart"},
		{"/components/echarts/bar", "ECharts bar"},
		{"/components/echarts/line", "ECharts line"},
		{"/components/echarts/scatter", "ECharts scatter"},
		{"/components/echarts/pie", "ECharts pie"},
		{"/components/echarts/radar", "ECharts radar"},
		{"/components/echarts/heatmap", "ECharts heatmap"},
		{"/components/echarts/boxplot", "ECharts box plot"},
		{"/examples/status-page", "Status page example"},
		{"/examples/go-echarts", "go-echarts catalog"},
		{"/examples/go-echarts/bar-basic", "Basic bar"},
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
	for _, want := range []string{"Goshtoso App Shells", "templ", "go-analyze/charts", "go-echarts", "Apache ECharts", "MIT license", "Apache-2.0 license", "site/internal/echartsassets/NOTICE.md"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("attributions page missing %q", want)
		}
	}

	for _, path := range []string{"/components/echarts/bar", "/components/echarts/line", "/components/echarts/scatter", "/components/echarts/pie", "/components/echarts/radar", "/components/echarts/heatmap", "/components/echarts/boxplot"} {
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
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/echarts/effect-scatter", nil))
	if recorder.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusPermanentRedirect)
	}
	if location := recorder.Header().Get("Location"); location != "/components/echarts/scatter" {
		t.Fatalf("Location = %q, want unified scatter route", location)
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

func TestEveryPortedEChartsRouteRenders(t *testing.T) {
	t.Parallel()
	handler := New()
	for _, example := range echartsexamples.All() {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/examples/go-echarts/"+example.Slug, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", example.Slug, recorder.Code)
			continue
		}
		body := recorder.Body.String()
		for _, want := range []string{example.Title, "echarts.init", "/charts/echarts/echarts@4.min.js"} {
			if !strings.Contains(body, want) {
				t.Errorf("GET %s missing %q", example.Slug, want)
			}
		}
	}
}

func TestPageLayoutPortsRenderUpstreamChartSet(t *testing.T) {
	t.Parallel()
	for _, slug := range []string{"page-center-layout", "page-flex-layout", "page-none-layout"} {
		recorder := httptest.NewRecorder()
		New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/examples/go-echarts/"+slug, nil))
		if count := strings.Count(recorder.Body.String(), "echarts.init"); count != 16 {
			t.Errorf("GET %s renders %d charts, want 16", slug, count)
		}
	}
}

func TestUnknownEChartsExampleIsNotFound(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/examples/go-echarts/not-real", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("GET unknown example = %d, want 404", recorder.Code)
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

func TestEChartsRuntimeIsLocal(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/charts/echarts/echarts@4.min.js", "/charts/echarts/echarts-gl.min.js", "/charts/echarts/echarts-liquidfill.min.js", "/charts/echarts/echarts-wordcloud.min.js", "/charts/echarts/maps/china.js", "/charts/echarts/maps/guangdong.js"} {
		recorder := httptest.NewRecorder()
		New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusOK || recorder.Body.Len() == 0 {
			t.Fatalf("GET %s status/body = %d/%d", path, recorder.Code, recorder.Body.Len())
		}
	}
}

func TestComponentDocsNavigationHasSearchGroupsAndComponentContract(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/heartbeat", nil))
	body := recorder.Body.String()
	for _, want := range []string{"Search docs...", "Server-rendered", "Interactive / Cartesian", "Examples", "component-doc-shell__sidebar", "components.KindHeartbeat", "Bar chart", "Pie chart", "Accessibility", "lg:grid-cols-2"} {
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
