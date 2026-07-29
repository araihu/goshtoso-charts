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
		"chartcontrol.WrapperMode", "chartcontrol.WrapperModeHidden", "goshtoso-charts:set-wrapper-mode", "goshtoso-charts:wrapper-mode-change",
		`hx-swap="outerHTML"`, "window.__goshtosoChartsControls.setWrapperMode",
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

func TestGuideTemplateDocumentsEveryWrapperLifecycleBoundary(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("guides.templ")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, want := range []string{
		"WrapperModeEnabled", "WrapperModeDisabled", "WrapperModeHidden", "WrapperModeOmitted",
		"Default. Wrapper, chart, available actions", "Chart stays visible, live, and theme-aware",
		"subtree is hidden, inert", "Server-only", "Only the chart renders", "focusReturn",
		"previousMode", "latest live snapshot", "Omitted skips wrapper-only export validation",
		"Interactive charts still require their chart runtime regardless of wrapper mode",
		"reject unknown values instead of casting arbitrary input",
		"one closed state", "overlapping wrapper booleans and precedence rules do not exist",
		"data-goshtoso-chart-wrapper-mode", "observable state, not a mutation API",
		"Disabled affects shared wrapper actions only", "enabled-to-omitted and omitted-to-non-omitted",
		"actionless non-omitted wrapper", "versioned same-origin external runtime", "preserves any action that was already unavailable",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("guide template missing wrapper lifecycle boundary %q", want)
		}
	}
}
