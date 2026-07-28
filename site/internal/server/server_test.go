package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	chartassets "github.com/araihu/goshtoso-charts/assets"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
)

func TestDemoRoutesRender(t *testing.T) {
	t.Parallel()
	handler := New()
	for _, test := range []struct {
		path string
		want string
	}{
		{"/", "Getting Started"},
		{"/attributions", "Foundation dependencies"},
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
		{"/components/interactive/graph", "Interactive graph"},
		{"/components/interactive/sankey", "Interactive Sankey"},
		{"/examples/live-availability", "Live availability"},
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

func TestGettingStartedReplacesChartCardOverview(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Install the module", "Mount chart assets", "Include chart dependencies",
		"Render a static chart", "Render an interactive chart", "Explore the catalog",
		"chartassets", "Handler", "dependencies", "Dependencies", "dependencies.WithCDN()",
		`href="/examples/live-availability"`, `class="max-w-3xl space-y-12"`,
		`class="codeblock overflow-x-auto"`, `x-data="{ copied: false, copyCode()`,
		`aria-label="Copy bash code"`, `aria-label="Copy Go code"`, `aria-label="Copy templ code"`,
		`x-text="copied ? 'Copied!' : 'Copy'"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("getting-started page missing %q", want)
		}
	}
	if count := strings.Count(body, `data-getting-started-step`); count != 6 {
		t.Errorf("getting-started step sections = %d, want 6", count)
	}
	for _, unwanted := range []string{
		"Charts for Goshtoso", `aria-label="Example monitor status"`,
		`class="grid gap-4 md:grid-cols-2"`, "Observation states",
		`<pre class="min-w-max p-4 text-sm`,
	} {
		if strings.Contains(body, unwanted) {
			t.Errorf("getting-started page retains old overview content %q", unwanted)
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
	body := attributions.Body.String()
	for _, want := range []string{
		"Foundation dependencies", "Chart and rendering libraries", "Bundled runtime and assets",
		"Goshtoso", "v0.0.13", "Goshtoso App Shells", "commit 4c4aa5ae787e", "templ", "v0.3.1020",
		"go-analyze/charts", "v0.6.0", "go-echarts", "v2.7.2", "Apache ECharts", "v5.4.3",
		`href="https://github.com/araihu/goshtoso"`, `href="https://github.com/go-echarts/go-echarts"`,
		`href="https://echarts.apache.org/"`, `href="https://github.com/apache/echarts/blob/5.4.3/LICENSE"`,
		`bg-primary/10`, "MIT", "Apache-2.0",
		"SHA-256 987554a0014ad7be585eccc91c4329d050b40c2c0ebd2e8ec84adca82c0eb843", "assets/NOTICE.md",
		`class="overflow-x-auto rounded-radius border`, `class="min-w-full w-full`,
		`class="min-w-52 px-4 py-3 font-bold"`, `class="min-w-32 px-4 py-3 font-bold"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("attributions page missing %q", want)
		}
	}
	if count := strings.Count(body, "<table"); count != 3 {
		t.Errorf("attributions table count = %d, want 3", count)
	}
	for _, header := range []string{"Project", "License", "Used for"} {
		if count := strings.Count(body, `scope="col">`+header+"</th>"); count != 3 {
			t.Errorf("attributions %q column header count = %d, want 3", header, count)
		}
	}
	if strings.Contains(body, "optional extensions") {
		t.Error("attributions page claims removed optional extensions are still pinned")
	}

	for _, path := range []string{"/components/interactive/bar", "/components/interactive/line", "/components/interactive/scatter", "/components/interactive/pie", "/components/interactive/radar", "/components/interactive/heatmap", "/components/interactive/boxplot", "/components/interactive/gauge", "/components/interactive/funnel", "/components/interactive/graph", "/components/interactive/sankey"} {
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
	for _, component := range []string{"bar", "line", "scatter", "pie", "radar", "heatmap", "boxplot", "gauge", "funnel", "graph", "sankey"} {
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

func TestHeartbeatDocumentationRedirectsToLiveAvailabilityExample(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/heartbeat", nil))
	if recorder.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusPermanentRedirect)
	}
	if location := recorder.Header().Get("Location"); location != "/examples/live-availability" {
		t.Fatalf("Location = %q, want live availability example", location)
	}
}

func TestStatusPageRedirectsToLiveAvailabilityExample(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/examples/status-page", nil))
	if recorder.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusPermanentRedirect)
	}
	if location := recorder.Header().Get("Location"); location != "/examples/live-availability" {
		t.Fatalf("Location = %q, want live availability example", location)
	}
}

type cancelOnFlushRecorder struct {
	*httptest.ResponseRecorder
	cancel context.CancelFunc
}

func (recorder *cancelOnFlushRecorder) Flush() {
	recorder.ResponseRecorder.Flush()
	recorder.cancel()
}

func TestLiveAvailabilityEventsEmitRendererNeutralSnapshots(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	recorder := &cancelOnFlushRecorder{ResponseRecorder: httptest.NewRecorder(), cancel: cancel}
	request := httptest.NewRequest(http.MethodGet, "/examples/live-availability/events", nil).WithContext(ctx)

	New().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "text/event-stream" {
		t.Fatalf("Content-Type = %q, want text/event-stream", contentType)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "event: chart\n") {
		t.Fatalf("event stream missing named chart event: %q", body)
	}
	data := strings.TrimPrefix(strings.SplitN(body, "\n\n", 2)[0], "event: chart\ndata: ")
	var snapshot interactive.CartesianSnapshot
	if err := json.Unmarshal([]byte(data), &snapshot); err != nil {
		t.Fatalf("decode CartesianSnapshot: %v; body=%q", err, body)
	}
	if len(snapshot.Categories) != 36 || len(snapshot.Series) != 3 {
		t.Fatalf("snapshot = %#v, want 36 categories and three availability-state series", snapshot)
	}
	assertOneHotAvailabilitySnapshot(t, snapshot)
}

func TestAvailabilityPagesExplicitlyDisableMonitoringAnimation(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/examples/live-availability"} {
		recorder := httptest.NewRecorder()
		New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		body := recorder.Body.String()
		if !strings.Contains(body, `data-goshtoso-charts-explicit-animation="false"`) || !strings.Contains(body, `"animation":false`) {
			t.Errorf("GET %s does not preserve explicit no-motion contract", path)
		}
	}
}

func TestAvailabilityPagesUseSemanticStatusPalette(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/examples/live-availability"} {
		recorder := httptest.NewRecorder()
		New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		body := recorder.Body.String()
		if !strings.Contains(body, "goshtoso-charts-palette-status") {
			t.Errorf("GET %s missing semantic status palette class", path)
		}
	}
}

func TestAvailabilityPagesUseFixedLabelCadenceAndRealCategories(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/examples/live-availability"} {
		recorder := httptest.NewRecorder()
		New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		body := recorder.Body.String()
		if !strings.Contains(body, `"axisLabel":{"interval":5,`) {
			t.Errorf("GET %s missing fixed six-bucket label cadence", path)
		}
		if !strings.Contains(body, `"showMinLabel":true,"showMaxLabel":true`) {
			t.Errorf("GET %s missing explicit endpoint-label visibility", path)
		}
		if strings.Contains(body, `"","`) {
			t.Errorf("GET %s encodes label sparsity as empty categories", path)
		}
	}
}

func TestAvailabilitySnapshotsShiftOneDenseOneHotBucket(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, time.July, 28, 12, 0, 0, 0, time.UTC)
	first := availabilitySnapshot(base, 0)
	next := availabilitySnapshot(base.Add(2*time.Second), 1)
	assertOneHotAvailabilitySnapshot(t, first)
	assertOneHotAvailabilitySnapshot(t, next)

	for index := 0; index < len(first.Categories)-1; index++ {
		if first.Categories[index+1] != next.Categories[index] {
			t.Fatalf("category window did not shift at %d: %q != %q", index, first.Categories[index+1], next.Categories[index])
		}
		for seriesIndex := range first.Series {
			if first.Series[seriesIndex].Values[index+1] != next.Series[seriesIndex].Values[index] {
				t.Fatalf("series %d window did not shift at %d", seriesIndex, index)
			}
		}
	}
	for _, seriesIndex := range []int{1, 2} {
		longest := 0
		run := 0
		for _, value := range first.Series[seriesIndex].Values {
			if value == 1 {
				run++
				if run > longest {
					longest = run
				}
			} else {
				run = 0
			}
		}
		if longest < 3 {
			t.Fatalf("series %q longest active run = %d, want at least 3", first.Series[seriesIndex].Name, longest)
		}
	}
}

func assertOneHotAvailabilitySnapshot(t *testing.T, snapshot interactive.CartesianSnapshot) {
	t.Helper()
	count := len(snapshot.Categories)
	if count != 36 || len(snapshot.Series) != 3 {
		t.Fatalf("snapshot shape = %d categories/%d series, want 36/3", count, len(snapshot.Series))
	}
	for seriesIndex, series := range snapshot.Series {
		if len(series.Values) != count {
			t.Fatalf("series %d length = %d, want %d", seriesIndex, len(series.Values), count)
		}
	}
	seenCategories := make(map[string]struct{}, count)
	for index, category := range snapshot.Categories {
		if category == "" {
			t.Fatalf("category %d is empty", index)
		}
		if _, exists := seenCategories[category]; exists {
			t.Fatalf("category %d duplicates %q", index, category)
		}
		seenCategories[category] = struct{}{}
	}
	for index := 0; index < count; index++ {
		active := 0
		for _, series := range snapshot.Series {
			switch series.Values[index] {
			case 0:
			case 1:
				active++
			default:
				t.Fatalf("bucket %d contains non-one-hot value %v", index, series.Values[index])
			}
		}
		if active != 1 {
			t.Fatalf("bucket %d active series = %d, want 1", index, active)
		}
	}
}

func TestGettingStartedAndAttributionsUseNativeSidebarIcons(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	body := recorder.Body.String()
	for _, want := range []string{`data-sidebar-icon="getting-started"`, `data-sidebar-icon="attributions"`, `aria-hidden="true"`} {
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
		{"/", `<title>Getting Started · Goshtoso Charts</title>`},
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
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/line", nil))
	body := recorder.Body.String()
	for _, want := range []string{"Search docs...", "Static / Vector", "Interactive / Cartesian", "Interactive / Relationships", "Examples", "component-doc-shell__sidebar", "components.KindLineChart", "Live availability", "Bar chart", "Pie chart", "Accessibility", "lg:grid-cols-2"} {
		if !strings.Contains(body, want) {
			t.Errorf("component docs page missing %q", want)
		}
	}
	if strings.Contains(body, "components.KindHeartbeat") {
		t.Error("component navigation still exposes domain-specific heartbeat kind")
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
