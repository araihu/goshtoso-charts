package pages

import (
	"reflect"
	"strings"
	"testing"
)

func TestInteractiveBoxPlotSiteUsesCanonicalChildPackage(t *testing.T) {
	t.Parallel()

	config := sampleInteractiveBoxPlot()
	if got := reflect.TypeOf(config).PkgPath(); got != "github.com/araihu/goshtoso-charts/components/interactive/boxplot" {
		t.Fatalf("site BoxPlot config package = %q", got)
	}
	code := interactiveChartBoxPlotCode()
	for _, want := range []string{
		"@interactiveboxplot.BoxPlot(interactiveboxplot.Config{",
		"[]interactiveboxplot.Series",
		"[]interactiveboxplot.Data",
		"chart.ChartOptions",
		`[]string{"Development", "Staging", "Production"}`,
		`{Min: 22, Q1: 38, Median: 51, Q3: 73, Max: 116}`,
	} {
		if !strings.Contains(code, want) {
			t.Errorf("BoxPlot snippet missing canonical API %q", want)
		}
	}
	if strings.Contains(code, "interactive.BoxPlot") {
		t.Fatal("BoxPlot snippet still teaches compatibility facade")
	}
}
