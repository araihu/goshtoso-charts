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
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "weekly-signups"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`-fullscreen-action`,
		`data-goshtoso-chart-expand`,
		`exportFromMenu($el, &#34;png&#34;)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("markup missing %q", want)
		}
	}
	if strings.Contains(markup, `data-goshtoso-chart-export="svg"`) {
		t.Fatal("canvas interactive chart exposed unsupported SVG export")
	}
	if strings.Contains(markup, `-chart-expand-export"`) {
		t.Fatal("single interactive PNG capability rendered a dropdown")
	}
}

func TestEChartRendersTrustedSnippet(t *testing.T) {
	t.Parallel()
	bar := charts.NewBar()
	bar.SetXAxis([]string{"Mon", "Tue"}).AddSeries("Signups", []opts.BarData{{Value: 12}, {Value: 18}})
	instance := newInstance(chartcomponents.KindInteractiveBar, renderConfig{Label: "Weekly signups", Caption: "Interactive example.", Chart: bar, ResponsiveWidth: true})
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
		`data-goshtoso-charts-responsive-width="true"`,
		"data-goshtoso-charts-theme-runtime", `--color-chart-text-strong`,
		`backgroundColor: surface`, `title: repeat`, `legend: repeat`, `xAxis: repeat`,
		`yAxis: repeat`, `radar: repeat`, `visualMap: themedVisualMaps`, `tooltip: repeat`,
		`series: themedSeries`, `MutationObserver`, `attributeFilter: ["class", "data-theme"]`,
		`subtree: false`, `animationDurationUpdate: 0`,
		`--color-chart-series-`, `themed.color = seriesColors`,
		`--color-chart-scale-low`, `--color-chart-scale-mid`, `--color-chart-scale-high`,
		`getImageData`, `rendererColor`,
		`ResizeObserver`, `responsiveFigures`, `targetFigures`, `data-goshtoso-charts-gauge-scale`,
		`data-goshtoso-charts-candlestick-styles`, `series.type === "boxplot"`, `series.type === "candlestick"`, `series.type === "gauge"`,
		`themedVisualMaps`, `current.color`, `themeSeriesItems`,
		`matchMedia("(prefers-color-scheme: dark)")`,
		`matchMedia("(prefers-reduced-motion: reduce)")`,
		`data-goshtoso-charts-pie-auto-emphasis`, `syncPieAutoEmphasis`,
		`clearInterval(state.timer)`, `type: "highlight"`, `type: "showTip"`,
		`data-goshtoso-charts-line3d-auto-rotate`, `series.type === "line3D"`,
		`autoRotate: !reduceMotion`,
		`ResizeObserver`, `chart.resize({ width: width, height: height, animation: { duration: 0 } })`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestResponsiveWidthOnlyDefaultsOmittedAndFullWidthHosts(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		width string
		want  bool
	}{
		{want: true},
		{width: "100%", want: true},
		{width: " 100% ", want: true},
		{width: "720px", want: false},
		{width: "48rem", want: false},
	} {
		if got := responsiveWidth(test.width); got != test.want {
			t.Errorf("responsiveWidth(%q) = %t, want %t", test.width, got, test.want)
		}
	}
}

func TestThemeRuntimeUsesIdentityPreservingImmediateSilentMerge(t *testing.T) {
	t.Parallel()
	for _, unwanted := range []string{
		`chart.setOption(themed, false, true)`,
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
		`resizeObserver.observe(target)`,
		`requestAnimationFrame(function ()`,
		`if (!explicitVisualMapColors) visualMap.inRange = { color: [scaleLow, scaleMid, scaleHigh] }`,
		`var themedCalendars = (current.calendar || []).map`, `var calendarItemStyle = calendar.itemStyle || {}`,
		`color: calendarItemStyle.color ? rendererColor(calendarItemStyle.color, surfaceAlt) : surfaceAlt`,
		`borderColor: calendarItemStyle.borderColor ? rendererColor(calendarItemStyle.borderColor, outline) : outline`,
		`monthLabel: Object.assign({}, calendar.monthLabel || {}, { color: muted })`,
		`series.type === "map"`, `themedItem.showLegendSymbol = false`, `item.sourceColor`, `item.className`,
		`gaugeScale.stops.map`, `observeTarget`,
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
		`observeTarget(figure, host);`,
		`observeTarget(figure, host && host.parentElement);`,
		`scheduleResize(figure);`,
		`window.echarts.getInstanceByDom(host)`,
		`chart.resize({ width: width, height: height, animation: { duration: 0 } });`,
		`window.addEventListener("resize", scheduleAll);`,
		`document.addEventListener("goshtoso-charts:resize",`,
		`document.addEventListener("goshtoso-charts:export-request",`,
		`detail.dataURL = chart.getDataURL({`,
	} {
		if !strings.Contains(themeRuntimeMarkup, want) {
			t.Errorf("responsive theme runtime missing %q", want)
		}
	}
	register := strings.Index(themeRuntimeMarkup, `var registerResponsive = function (figure)`)
	if register < 0 {
		t.Fatal("interactive runtime missing responsive registration")
	}
	registration := themeRuntimeMarkup[register:]
	observe := strings.Index(registration, `observeTarget(figure, host)`)
	if observe < 0 {
		t.Fatal("interactive registration must observe chart host geometry")
	}
	resize := strings.Index(registration[observe:], `scheduleResize(figure)`)
	apply := strings.Index(registration[observe:], `apply(figure)`)
	if resize < 0 || apply < 0 || resize > apply {
		t.Fatal("interactive registration must observe and schedule resize before applying theme options")
	}
	if strings.Contains(themeRuntimeMarkup, `echarts.init(`) {
		t.Fatal("shared runtime must resize the existing instance without reinitializing it")
	}
}

func TestThemeRuntimeCleansResponsiveObserversWithoutDuplicateRegistration(t *testing.T) {
	t.Parallel()
	for _, want := range []string{
		`var responsiveFigures = new Map();`,
		`if (responsiveFigures.has(figure))`,
		`resizeObserver.unobserve(target);`,
		`responsiveFigures.delete(figure);`,
		`if (!figure.isConnected) unregister(figure);`,
		`childList: true`,
		`subtree: true`,
	} {
		if !strings.Contains(themeRuntimeMarkup, want) {
			t.Errorf("responsive lifecycle missing %q", want)
		}
	}
}

func TestThemeRuntimeSettlesGeometryWithoutRestartingAnimation(t *testing.T) {
	t.Parallel()
	for _, want := range []string{
		`stableFrames`,
		`frameCount`,
		`maxSettleFrames`,
		`settleFrames`,
		`requestAnimationFrame(step);`,
		`animation: { duration: 0 }`,
	} {
		if !strings.Contains(themeRuntimeMarkup, want) {
			t.Errorf("responsive settling runtime missing %q", want)
		}
	}
	for _, unwanted := range []string{`echarts.init(`, `echarts.dispose(`} {
		if strings.Contains(themeRuntimeMarkup, unwanted) {
			t.Errorf("responsive runtime must preserve instance, found %q", unwanted)
		}
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
