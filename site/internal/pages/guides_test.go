package pages

import (
	"os"
	"strings"
	"testing"
)

func TestGuideSourceUsesActualExportAndControlTypes(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("guides.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, want := range []string{
		"chartcontrol.Options", "chartcontrol.Bool(false)", "chartcontrol.ExportOptions", "chartcontrol.ExportSVG", "chartcontrol.ExportPNG",
		"chartcontrol.ExportBackgroundTransparent", "interactive.ChartOptions", "interactive.TooltipOptions", "interactive.Bool(true)",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("guide source missing actual API %q", want)
		}
	}
	for _, forbidden := range []string{"go-echarts", "Apache ECharts", "go-analyze/charts", "opts.", "charts."} {
		if strings.Contains(text, forbidden) {
			t.Errorf("guide source exposes private renderer %q", forbidden)
		}
	}
}
