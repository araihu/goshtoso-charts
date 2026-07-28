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
		{"/components/scatter", "Scatter chart"},
		{"/components/radar", "Radar chart"},
		{"/components/candlestick", "Candlestick"},
		{"/components/funnel", "Funnel chart"},
		{"/components/heatmap", "Heat map"},
		{"/components/table", "Table"},
		{"/components/violin", "Violin chart"},
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
		{"/components/interactive/tree", "Interactive tree"},
		{"/components/interactive/sunburst", "Interactive sunburst"},
		{"/components/interactive/treemap", "Interactive treemap"},
		{"/components/interactive/parallel", "Interactive parallel coordinates"},
		{"/components/interactive/theme-river", "Interactive theme river"},
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

func TestViolinDocumentationPreservesUpstreamSamplesWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/violin", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Violin chart", "Distribution Shapes", "Normal", "Right Skewed", "Bimodal", "Tight",
		"200", "80 Gaussian density bands", "1200 x 800", "Mean", "Median", "Q1", "Q3",
		"violin.Violin", "components.KindViolinChart", "Distribution", "MarkLines", "Quantiles",
		"Exact sample statistics", "Expand", "SVG", "PNG",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("violin documentation missing upstream or contract content %q", want)
		}
	}
	for _, unwanted := range []string{"go-analyze", "violin_chart-2-samples/main.go", "infrastructure", "operations", "raw map"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("violin component page contains non-neutral content %q", unwanted)
		}
	}
	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{"examples/1-Painter/violin_chart-2-samples/main.go", "1fe31b06b8a82e00df877ff4417a75858547c1c2"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attribution missing %q", want)
		}
	}
}

func TestHeatMapDocumentationPreservesUpstreamBasicExampleWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/heatmap", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Basic heat map", "Heat Map Chart", "X-Axis", "Y-Axis",
		"4.4", "4.9", "7", "7.5", "4.3", "2.6", "5.9", "9", "6.4", "2.3",
		"3.3", "3.2", "1.9", "6", "4.6", "1.9 · cold", "9 · warm",
		"heatmap.HeatMap", "components.KindHeatMapChart", "ValueRange", "GradientStop",
		"Exact values", "Custom semantic scale", "Fullscreen", "Collapse", "Expand", "SVG", "PNG",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("heat-map documentation missing upstream or contract content %q", want)
		}
	}
	for _, unwanted := range []string{"go-analyze", "examples/1-Painter/heat_map-1-basic/main.go", "infrastructure", "operations", "raw map"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("heat-map component page contains non-neutral content %q", unwanted)
		}
	}
}

func TestTableDocumentationPreservesOfficialExampleWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/table", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"People directory", "Name", "Age", "Address", "Tag", "Action",
		"John Brown", "32", "New York No. 1 Lake Park", "nice, developer", "Send Mail",
		"Jim Green", "42", "London No. 1 Lake Park", "wow",
		"Joe Black", "Sidney No. 1 Lake Park", "cool, teacher",
		"Market changes", "Datadog Inc", "97.32", "-7.49%", "Hashicorp Inc", "28.66", "-9.25%", "Gitlab Inc", "51.63", "+4.32%",
		"Accessible data table", "components.KindTable", "Explicit presentation overrides",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("table documentation missing upstream example content %q", want)
		}
	}
	for _, unwanted := range []string{"go-analyze", "examples/1-Painter/table-1/main.go", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("table component page contains non-neutral content %q", unwanted)
		}
	}
	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{"examples/1-Painter/table-1/main.go", "1fe31b06b8a82e00df877ff4417a75858547c1c2"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attributions missing pinned table source %q", want)
		}
	}
}

func TestFunnelDocumentationPreservesPinnedOfficialExampleWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/funnel", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Basic funnel", "Funnel", "Show", "Click", "Visit", "Inquiry", "Order", "Pay", "Cancel",
		"100", "80", "60", "40", "20", "10", "2", "Exact stage values", "Share of first stage",
		"funnel.Funnel", "components.KindFunnelChart", "typed Stages", "Color and Class", "SVG", "PNG",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("funnel documentation missing upstream or contract content %q", want)
		}
	}
	for _, unwanted := range []string{"go-analyze", "examples/1-Painter/funnel_chart-1-basic/main.go", "infrastructure", "operations", "raw map"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("funnel component page contains non-neutral content %q", unwanted)
		}
	}
	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{"examples/1-Painter/funnel_chart-1-basic/main.go", "1fe31b06b8a82e00df877ff4417a75858547c1c2"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attributions missing pinned funnel source %q", want)
		}
	}
}

func TestRadarDocumentationPreservesUpstreamBasicExampleWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/radar", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Basic radar chart", "Allocated Budget", "Actual Spending", "Sales", "Administration",
		"Information Technology", "Customer Support", "Development", "Marketing",
		"4200", "3000", "20000", "35000", "50000", "18000",
		"5000", "14000", "28000", "26000", "42000", "21000",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("radar documentation missing upstream example content %q", want)
		}
	}
	for _, unwanted := range []string{"go-analyze", "examples/1-Painter/radar_chart-1-basic/main.go", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("radar component page contains non-neutral content %q", unwanted)
		}
	}
}

func TestCandlestickDocumentationPreservesUpstreamBasicExampleWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/candlestick", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Candlestick Chart", "Stock Price", "Day 1", "Day 2", "Day 3", "Day 4", "Day 5", "Day 6", "Day 7",
		"100", "110", "95", "105", "115", "112", "118", "108", "120", "104", "113", "109", "116", "106", "114", "121", "111", "119",
		"Exact OHLC values", "Increase", "Decrease", "Low &lt;= Open/Close &lt;= High", "components.KindCandlestickChart",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("candlestick documentation missing upstream example content %q", want)
		}
	}
	for _, unwanted := range []string{"go-analyze", "examples/1-Painter/candlestick_chart-1-basic/main.go", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("candlestick component page contains non-neutral content %q", unwanted)
		}
	}
	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	if !strings.Contains(attributions.Body.String(), "examples/1-Painter/candlestick_chart-1-basic/main.go") {
		t.Error("central attributions missing official candlestick source path")
	}
}

func TestThemeRiverDocumentationPreservesPinnedUpstreamExampleWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/theme-river", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"ThemeRiver-SingleAxis-Time", "DQ", "TY", "SS", "QG", "SY", "DD",
		"2015/11/08", "2015/11/28", "Exact stream values",
		"components.KindInteractiveThemeRiver",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("theme river documentation missing upstream example content %q", want)
		}
	}
	for _, unwanted := range []string{"go-echarts", "examples/themeriver.go", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("theme river component page contains non-neutral content %q", unwanted)
		}
	}
	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{"examples/themeriver.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attributions missing %q", want)
		}
	}
}

func TestInteractiveCandlestickDocumentationPreservesPinnedUpstreamExampleWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/candlestick", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Interactive candlestick", "Candlestick example", "2018/1/24", "2018/6/13",
		"2320.26", "2287.3", "2362.94", "2148.35", "2126.22", "2190.1",
		"Exact OHLC values", "Rise", "Fall", "components.KindInteractiveCandlestick",
		`"type":"inside"`, `"start":50`, `"end":100`, `"valueDim":"highest"`, `"valueDim":"lowest"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("interactive candlestick documentation missing upstream example content %q", want)
		}
	}
	for _, unwanted := range []string{"go-echarts", "examples/kline.go", "Kline", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("interactive candlestick component page contains non-neutral content %q", unwanted)
		}
	}
	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{"examples/kline.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attributions missing interactive candlestick source evidence %q", want)
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
		"examples/1-Painter/scatter_chart-3-dense_data/main.go",
		"examples/1-Painter/radar_chart-1-basic/main.go",
		"examples/1-Painter/candlestick_chart-1-basic/main.go", "examples/1-Painter/funnel_chart-1-basic/main.go", "examples/1-Painter/heat_map-1-basic/main.go", "examples/parallel.go",
		"examples/1-Painter/table-1/main.go", "1fe31b06b8a82e00df877ff4417a75858547c1c2",
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

	for _, path := range []string{"/components/scatter", "/components/radar", "/components/candlestick", "/components/funnel", "/components/heatmap", "/components/table", "/components/interactive/bar", "/components/interactive/line", "/components/interactive/scatter", "/components/interactive/pie", "/components/interactive/radar", "/components/interactive/heatmap", "/components/interactive/boxplot", "/components/interactive/gauge", "/components/interactive/funnel", "/components/interactive/graph", "/components/interactive/sankey", "/components/interactive/sunburst", "/components/interactive/parallel"} {
		page := httptest.NewRecorder()
		handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, path, nil))
		for _, unwanted := range []string{"backed by go-echarts", "with go-echarts options", ">go-echarts catalog<", "go-analyze"} {
			if strings.Contains(page.Body.String(), unwanted) {
				t.Errorf("GET %s repeats centralized attribution %q", path, unwanted)
			}
		}
	}
}

func TestParallelDocumentationPreservesOfficialMultiSeriesExampleWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	handler := New()
	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/components/interactive/parallel", nil))
	body := page.Body.String()
	for _, want := range []string{
		"Multi Series", "Beijing", "Guangzhou", "Shanghai", "Date", "AQI", "PM2.5", "PM10", "CO", "NO2", "SO2", "Level",
		"Good", "Moderate", "Lightly", "Moderately", "Heavily", "Severely",
		`style="width:900px;height:500px;"`, `"max":31`, `"inverse":true`, `"nameLocation":"start"`,
		`"value":[9,267,216,280,4.8,108,64,"Heavily"]`,
		`"value":[20,73,102,182,2.787,44,19,"Moderate"]`,
		`"value":[16,134,83,167,1.16,57,43,"Lightly"]`,
		"Exact observations and values", "Semantic class", "components.KindInteractiveParallel",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("parallel documentation missing upstream example content %q", want)
		}
	}
	for series, want := range map[string]int{"Beijing": 21, "Guangzhou": 21, "Shanghai": 21} {
		if count := strings.Count(body, `scope="row">`+series+"</th>"); count != want {
			t.Errorf("parallel exact table %s rows = %d, want %d", series, count, want)
		}
	}
	for _, unwanted := range []string{"go-echarts", "examples/parallel.go", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("parallel component page contains non-neutral content %q", unwanted)
		}
	}

	attributions := httptest.NewRecorder()
	handler.ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{"examples/parallel.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attributions missing parallel source evidence %q", want)
		}
	}
}

func TestGaugeDocumentationConsolidatesOfficialLiquidVariantsWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	handler := New()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/gauge", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Interactive gauge", "interactive.Gauge", "components.KindInteractiveGauge", "GaugeVariantLiquid",
		"basic liquid example", "show label", "show outline", "disable wave animation",
		"shape(Diamond)", "shape(Pin)", "shape(Arrow)", "shape(Triangle)",
		"Wave 1", "0.3", "Wave 2", "0.4", "Wave 3", "0.5", "Range: 0 to 1",
		`data-gauge-variant="progress"`, `data-gauge-liquid-variants`, `data-gauge-liquid-variant="diamond"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("gauge documentation missing liquid content %q", want)
		}
	}
	if count := strings.Count(body, `data-goshtoso-gauge-liquid-values`); count != 8 {
		t.Errorf("liquid exact-summary count = %d, want 8", count)
	}
	for _, unwanted := range []string{"go-echarts", "echarts-liquidfill", "Apache ECharts", "examples/liquid.go", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("gauge component page contains private or invented framing %q", unwanted)
		}
	}
	if strings.Contains(body, "/components/interactive/liquid") || strings.Contains(body, "KindInteractiveLiquid") {
		t.Fatal("liquid treatment created a separate public component identity")
	}

	attributions := httptest.NewRecorder()
	handler.ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{"examples/liquid.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd", "v3.1.0", "BSD-3-Clause"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attributions missing liquid source evidence %q", want)
		}
	}
}

func TestWordCloudDocumentationPreservesOfficialVariantsWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/word-cloud", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Interactive word cloud", "interactive.WordCloud", "components.KindInteractiveWordCloud",
		"basic WordCloud example", "cardioid shape", "star shape", "Sam S Club", "Macys",
		"NCAA baseball tournament", "Point Break", "Exact word values", "semantic classes or explicit colors",
		`"sizeRange":[14,80]`, `"shape":"cardioid"`, `"shape":"star"`,
		`aria-label="basic WordCloud example"`, `data-word-cloud-variants`, "Interactive / Statistical",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("word-cloud documentation missing upstream content %q", want)
		}
	}
	if count := strings.Count(body, `scope="row">Sam S Club</th>`); count != 3 {
		t.Errorf("word-cloud exact-list variant count = %d, want 3", count)
	}
	for _, unwanted := range []string{"go-echarts", "echarts-wordcloud", "Apache ECharts", "examples/wordcloud.go", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("word-cloud component page contains private or invented framing %q", unwanted)
		}
	}

	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{"examples/wordcloud.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd", "echarts-wordcloud", "v2.1.0", "ISC"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attributions missing word-cloud source evidence %q", want)
		}
	}
}

func TestMapDocumentationPreservesOfficialVariantsWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/map", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Interactive map", "interactive.Map", "components.KindInteractiveMap", "Interactive / Geographic",
		"basic map example", "show label", "VisualMap", "Guangdong province", "Map-theme",
		"北京", "上海", "广东", "辽宁", "山东", "山西", "陕西", "新疆", "内蒙古",
		"深圳市", "广州市", "湛江市", "汕头市", "东莞市", "佛山市", "云浮市", "肇庆市", "梅州市",
		"Exact region values", "semantic classes or explicit colors", `data-map-variants`, `"map":"广东"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("map documentation missing upstream content %q", want)
		}
	}
	if count := strings.Count(body, `data-map-variant=`); count != 5 {
		t.Errorf("map variant count = %d, want 5", count)
	}
	for _, unwanted := range []string{"go-echarts", "Apache ECharts", "examples/map.go", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("map component page contains private or invented framing %q", unwanted)
		}
	}

	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{"examples/map.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd", "41f247b1cbb6", "national and Guangdong geometry"} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attributions missing map source evidence %q", want)
		}
	}
}

func TestGeoDocumentationPreservesOfficialVariantsWithoutEngineBranding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/geo", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Interactive geo", "interactive.Geo", "components.KindInteractiveGeo", "Interactive / Geographic",
		"basic geo example", "Guangdong province", "Geo plots coordinate series", "Map remains the region-value choropleth",
		"北京", "上海", "重庆", "武汉", "台湾", "香港", "汕头", "深圳", "广州",
		"116.40", "39.90", "121.47", "31.23", "106.55", "29.56", "114.31", "30.52",
		"121.30", "25.03", "114.17", "22.28", "116.69", "23.39", "114.07", "22.62", "113.23", "23.16",
		"Exact coordinate values", "semantic Color or Class paint", `data-geo-variants`,
		`data-geo-variant="effect-scatter"`, `data-geo-variant="scatter"`,
		`"type":"effectScatter"`, `"type":"scatter"`, `"map":"广东"`,
		`"period":4`, `"scale":6`, `"brushType":"stroke"`, `"calculable":true`, `"max":100`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("geo documentation missing upstream content %q", want)
		}
	}
	if count := strings.Count(body, `data-geo-variant=`); count != 2 {
		t.Errorf("geo variant count = %d, want 2", count)
	}
	for _, unwanted := range []string{"go-echarts", "Apache ECharts", "examples/geo.go", "infrastructure", "operations"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("geo component page contains private or invented framing %q", unwanted)
		}
	}

	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	for _, want := range []string{
		"examples/geo.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd",
		"fixed literals", "same [0,100) domain", "national and Guangdong geometry",
	} {
		if !strings.Contains(attributions.Body.String(), want) {
			t.Errorf("central attributions missing geo source evidence %q", want)
		}
	}
}

func TestSunburstDocumentationPreservesOfficialBasicExample(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/sunburst", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Basic sunburst example", "parent-0", "child-0", "parent-6", "child-6",
		"Exact hierarchy and values", "root class and attributes",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("sunburst documentation missing upstream example content %q", want)
		}
	}
	for _, unwanted := range []string{"service ownership", "infrastructure", "ECharts", "go-echarts", "examples/sunburst.go"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("sunburst documentation contains forbidden framing or API branding %q", unwanted)
		}
	}
	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	if !strings.Contains(attributions.Body.String(), "examples/sunburst.go") {
		t.Error("central attributions missing official sunburst source path")
	}
}

func TestTreemapDocumentationPreservesOfficialBasicExample(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/treemap", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Basic treemap example", "File system usage", "d1", "d2", "d3", "f39",
		"1000", "450", "Exact hierarchy and values", "typed recursive Nodes",
		"breadcrumb", "same chart", `style="width:100%;height:500px;"`,
		`aria-label="Basic treemap example"`, `"leafDepth":1`, "LeafDepth",
		"Interactive / Relationships",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("treemap documentation missing upstream example content %q", want)
		}
	}
	for _, unwanted := range []string{"service ownership", "infrastructure", "ECharts", "go-echarts", "examples/treemap.go"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("treemap documentation contains forbidden framing or API branding %q", unwanted)
		}
	}
	attributions := httptest.NewRecorder()
	New().ServeHTTP(attributions, httptest.NewRequest(http.MethodGet, "/attributions", nil))
	if !strings.Contains(attributions.Body.String(), "examples/treemap.go") {
		t.Error("central attributions missing official treemap source path")
	}
}

func TestScatterDocumentationPreservesUpstreamDenseExample(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/scatter", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Dense scatter data", "Dense Scatter Chart Demo", "One", "Two", "Three",
		"1,000 categories", "SMA(100)", "maximum references", "foo 0", "foo 999",
		"examples/1-Painter/scatter_chart-3-dense_data/main.go",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("scatter documentation missing upstream example content %q", want)
		}
	}
	for _, unwanted := range []string{"Scatter series by day", "Email", "Union Ads", "Video Ads"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("scatter documentation retains invented framing %q", unwanted)
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

func TestEveryCurrentChartPageUsesSharedControlsAndCapabilityGating(t *testing.T) {
	t.Parallel()
	handler := New()
	tests := []struct {
		path         string
		wantDropdown bool
	}{
		{path: "/components/line", wantDropdown: true},
		{path: "/components/bar", wantDropdown: true},
		{path: "/components/pie", wantDropdown: true},
		{path: "/components/scatter", wantDropdown: true},
		{path: "/components/heatmap", wantDropdown: true},
		{path: "/components/interactive/bar"},
		{path: "/components/interactive/line"},
		{path: "/components/interactive/scatter"},
		{path: "/components/interactive/heatmap"},
		{path: "/components/interactive/pie"},
		{path: "/components/interactive/radar"},
		{path: "/components/interactive/boxplot"},
		{path: "/components/interactive/candlestick"},
		{path: "/components/interactive/gauge"},
		{path: "/components/interactive/funnel"},
		{path: "/components/interactive/graph"},
		{path: "/components/interactive/sankey"},
		{path: "/components/interactive/tree"},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d", test.path, recorder.Code)
			}
			body := recorder.Body.String()
			for _, want := range []string{
				`data-goshtoso-chart-control="collapse"`,
				`data-goshtoso-chart-control="fullscreen"`,
				`data-goshtoso-chart-expand`, `role="dialog"`,
				`src="` + chartassets.ControlRuntimeURL + `"`,
			} {
				if !strings.Contains(body, want) {
					t.Errorf("GET %s missing %q", test.path, want)
				}
			}
			if got := strings.Contains(body, `data-goshtoso-chart-export-menu`); got != test.wantDropdown {
				t.Errorf("GET %s export dropdown presence = %t, want %t", test.path, got, test.wantDropdown)
			}
			if test.wantDropdown {
				for _, want := range []string{`>SVG</button>`, `>PNG</button>`} {
					if !strings.Contains(body, want) {
						t.Errorf("GET %s dropdown missing %q", test.path, want)
					}
				}
				if strings.Contains(body, `data-goshtoso-chart-export="`) {
					t.Errorf("GET %s rendered direct export button for multi-format capability", test.path)
				}
			} else if !strings.Contains(body, `data-goshtoso-chart-export="png"`) {
				t.Errorf("GET %s missing direct PNG export", test.path)
			}
		})
	}
}

func TestChartControlRuntimeIsLocal(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, chartassets.ControlRuntimeURL, nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "requestFullscreen") {
		t.Fatalf("GET control runtime status/body = %d/%q", recorder.Code, recorder.Body.String())
	}
}

func TestComponentDocsNavigationHasSearchGroupsAndComponentContract(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/line", nil))
	body := recorder.Body.String()
	for _, want := range []string{"Search docs...", "Static / Vector", "Interactive / Cartesian", "Interactive / Geographic", "Interactive / Relationships", "Examples", "component-doc-shell__sidebar", "components.KindLineChart", "Live availability", "Bar chart", "Pie chart", "Scatter chart", "Map", "Geo", "Tree", "Accessibility", "lg:grid-cols-2"} {
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

func TestInteractiveTreeDocsStayRendererNeutralAndExposeExactDataGuidance(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/tree", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Interactive tree", "interactive.Tree", "components.KindInteractiveTree",
		"typed recursive Roots", "Layered is the zero-value layout",
		"exact-data source", "Basic tree example",
		"One root with three branches", "Node3", "Child3",
		`aria-label="Basic tree example"`, `"collapsed":true`,
		`style="width:100%;height:440px;"`,
		`href="/components/interactive/tree"`,
		"Interactive / Relationships",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("interactive tree docs missing %q", want)
		}
	}
	for _, forbidden := range []string{"go-echarts", "Apache ECharts", "TreeChart", "opts.Tree"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("interactive tree docs expose private renderer term %q", forbidden)
		}
	}
	if strings.Contains(body, `style="width:900px;height:440px;"`) {
		t.Error("interactive tree docs fell back to fixed renderer width")
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
