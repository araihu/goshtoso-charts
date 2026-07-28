package interactive

import (
	"bytes"
	"context"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	charts "github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func TestInteractiveSupportsSharedControlsAndPNGExport(t *testing.T) {
	t.Parallel()
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: "320px"}))
	bar.SetXAxis([]string{"Mon"})
	bar.AddSeries("signups", []opts.BarData{{Value: 12}})
	instance := newInstance(chartcomponents.KindInteractiveBar, renderConfig{
		Label:    "Weekly signups",
		Chart:    bar,
		Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export:   &chartcontrol.ExportOptions{Filename: "weekly-signups"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`data-goshtoso-chart-control="fullscreen"`,
		`data-goshtoso-chart-control="collapse"`,
		`data-goshtoso-chart-expand`,
		`data-goshtoso-chart-export="png"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("markup missing %q", want)
		}
	}
	if strings.Contains(markup, `data-goshtoso-chart-export="svg"`) {
		t.Fatal("canvas interactive chart exposed unsupported SVG export")
	}
	if strings.Contains(markup, `data-goshtoso-chart-export-menu`) {
		t.Fatal("single interactive PNG capability rendered a dropdown")
	}
}

func TestEChartRendersTrustedSnippet(t *testing.T) {
	t.Parallel()
	bar := charts.NewBar()
	bar.SetXAxis([]string{"Mon", "Tue"}).AddSeries("Signups", []opts.BarData{{Value: 12}, {Value: 18}})
	instance := newInstance(chartcomponents.KindInteractiveBar, renderConfig{Label: "Weekly signups", Caption: "Interactive example.", Chart: bar})
	if instance.Kind() != chartcomponents.KindInteractiveBar {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"goshtoso-charts-interactive", "Weekly signups", "echarts.init", "Interactive example.",
		`data-goshtoso-charts-explicit-colors="false"`,
		`data-goshtoso-charts-explicit-animation="default"`,
		"data-goshtoso-charts-theme-runtime", `--color-chart-text-strong`,
		`backgroundColor: surface`, `title: repeat`, `legend: repeat`, `xAxis: repeat`,
		`yAxis: repeat`, `radar: repeat`, `visualMap: themedVisualMaps`, `tooltip: repeat`,
		`series: themedSeries`, `MutationObserver`, `attributeFilter: ["class", "data-theme"]`,
		`subtree: false`, `animationDurationUpdate: 0`,
		`--color-chart-series-`, `themed.color = seriesColors`,
		`--color-chart-scale-low`, `--color-chart-scale-mid`, `--color-chart-scale-high`,
		`getImageData`, `rendererColor`,
		`ResizeObserver`, `pendingResizeHosts`, `scheduleResize(entry.target)`, `data-goshtoso-charts-gauge-scale`,
		`data-goshtoso-charts-candlestick-styles`, `series.type === "boxplot"`, `series.type === "candlestick"`, `series.type === "gauge"`,
		`themedVisualMaps`, `current.color`, `themeSeriesItems`,
		`matchMedia("(prefers-color-scheme: dark)")`,
		`matchMedia("(prefers-reduced-motion: reduce)")`,
		`ResizeObserver`, `chart.resize()`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestThemeRuntimeUsesIdentityPreservingImmediateSilentMerge(t *testing.T) {
	t.Parallel()
	for _, unwanted := range []string{
		`chart.setOption(themed, false, true)`,
		`subtree: true`,
		`type: series.type`,
	} {
		if strings.Contains(themeRuntimeMarkup, unwanted) {
			t.Errorf("theme runtime must not contain %q", unwanted)
		}
	}
	for _, want := range []string{
		`id: series.id`, `name: series.name`, `animationDurationUpdate: 0`,
		`notMerge: false`, `lazyUpdate: false`, `silent: true`,
		`explicitAnimation === "default"`, `themed.animation = false`,
		`subtree: false`,
		`resizeObserver.observe(host)`,
		`requestAnimationFrame(function ()`,
		`if (!explicitColors) visualMap.inRange = { color: [scaleLow, scaleMid, scaleHigh] }`,
		`gaugeScale.stops.map`, `observedHosts`,
		`if (value === inherited) return rendererColor(fallback, fallback)`,
		`series.type === "tree" || series.type === "sunburst" || series.type === "treemap"`,
		`Never`, `partial Treemap levels`,
		`eachSeriesByType("sunburst"`, `type: "sunburstRootToNode"`,
		`eachSeriesByType("treemap"`, `type: "treemapRootToNode"`,
	} {
		if !strings.Contains(themeRuntimeMarkup, want) {
			t.Errorf("immediate identity-preserving theme runtime missing %q", want)
		}
	}
}

func TestThemeRuntimeResizesCanvasAfterConsumerHostShrinks(t *testing.T) {
	t.Parallel()
	// Browser regression: shell layout produced an 847 px canvas in a 607 px
	// consumer host. Registration must resize immediately, then observe later
	// wrapper and responsive layout changes without replacing the chart instance.
	for _, want := range []string{
		`var resizeObserver = window.ResizeObserver ? new ResizeObserver`,
		`entries.forEach(function (entry) { scheduleResize(entry.target); });`,
		`if (resizeObserver && host) resizeObserver.observe(host);`,
		`scheduleResize(host);`,
		`window.echarts.getInstanceByDom(host)`,
		`chart.resize();`,
	} {
		if !strings.Contains(themeRuntimeMarkup, want) {
			t.Errorf("responsive theme runtime missing %q", want)
		}
	}
	register := strings.Index(themeRuntimeMarkup, `register: function (figure)`)
	if register < 0 {
		t.Fatal("interactive runtime missing shared registration")
	}
	registration := themeRuntimeMarkup[register:]
	observe := strings.Index(registration, `resizeObserver.observe(host)`)
	if observe < 0 {
		t.Fatal("interactive registration must observe chart host size")
	}
	resize := strings.Index(registration[observe:], `scheduleResize(host)`)
	apply := strings.Index(registration[observe:], `apply(figure)`)
	if resize < 0 || apply < 0 || resize > apply {
		t.Fatal("interactive registration must observe and schedule resize before applying theme options")
	}
	if strings.Contains(themeRuntimeMarkup, `echarts.init(`) {
		t.Fatal("shared runtime must resize the existing instance without reinitializing it")
	}
}

func TestEChartMarksExplicitColorsForRuntimePrecedence(t *testing.T) {
	t.Parallel()
	bar := charts.NewBar()
	bar.SetXAxis([]string{"Mon"}).AddSeries("Signups", []opts.BarData{{Value: 12}})
	instance := newInstance(chartcomponents.KindInteractiveBar, renderConfig{
		Label: "Weekly signups", Chart: bar, Style: charttheme.Style{Colors: []string{"#123456"}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(output.String(), `data-goshtoso-charts-explicit-colors="true"`) {
		t.Fatal("rendered markup did not preserve explicit color precedence")
	}
}

func TestEChartRejectsMissingContract(t *testing.T) {
	t.Parallel()
	for _, cfg := range []renderConfig{{}, {Label: "Missing chart"}} {
		var output bytes.Buffer
		if err := newInstance(chartcomponents.KindInteractiveBar, cfg).Render(context.Background(), &output); err == nil {
			t.Fatalf("Render(%+v) error = nil", cfg)
		}
	}
}
