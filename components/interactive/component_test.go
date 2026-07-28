package interactive

import (
	"bytes"
	"context"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	charts "github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

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
		"data-goshtoso-charts-theme-runtime", `--color-chart-text-strong`,
		`backgroundColor: surface`, `title: repeat`, `legend: repeat`, `xAxis: repeat`,
		`yAxis: repeat`, `radar: repeat`, `visualMap: repeat`, `tooltip: repeat`,
		`series: themedSeries`, `MutationObserver`, `attributeFilter: ["class", "data-theme"]`,
		`--color-chart-series-`, `themed.color = seriesColors`,
		`matchMedia("(prefers-color-scheme: dark)")`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
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
