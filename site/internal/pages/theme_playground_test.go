package pages

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestThemePlaygroundReusesCentralizedExamples(t *testing.T) {
	t.Parallel()
	// The static source paths are centralized in the Line/Pie coverage records;
	// the interactive packages expose one central pinned path each.
	for _, path := range []string{
		"examples/1-Painter/line_chart-1-basic/main.go",
		"examples/1-Painter/pie_chart-1-basic/main.go",
		interactiveBarUpstreamPath,
		interactiveRadarUpstreamPath,
	} {
		if path == "" {
			t.Fatal("theme playground upstream source path is empty")
		}
	}
	if interactiveBarUpstreamPath != "examples/bar.go" || interactiveRadarUpstreamPath != "examples/radar.go" {
		t.Fatal("theme playground interactive source paths drifted")
	}

	staticLine, upstreamLine := themePlaygroundStaticLine(), sampleBasicLine()
	if !reflect.DeepEqual(staticLine.Labels, upstreamLine.Labels) || !reflect.DeepEqual(staticLine.Series, upstreamLine.Series) {
		t.Fatal("theme playground static Line drifted from centralized basic treatment")
	}
	staticPie, upstreamPie := themePlaygroundStaticPie(), sampleBasicPie()
	if !reflect.DeepEqual(staticPie.Slices, upstreamPie.Slices) {
		t.Fatal("theme playground static Pie drifted from centralized basic treatment")
	}
	interactiveBar, upstreamBar := themePlaygroundInteractiveBar(), sampleInteractiveBar()
	if !reflect.DeepEqual(interactiveBar.XAxis, upstreamBar.XAxis) || !reflect.DeepEqual(interactiveBar.Series, upstreamBar.Series) {
		t.Fatal("theme playground interactive Bar drifted from centralized basic treatment")
	}
	interactiveRadar, upstreamRadar := themePlaygroundInteractiveRadar(), sampleInteractiveRadarBase()
	if !reflect.DeepEqual(interactiveRadar.Indicators, upstreamRadar.Indicators) || !reflect.DeepEqual(interactiveRadar.Series, upstreamRadar.Series) {
		t.Fatal("theme playground interactive Radar drifted from centralized basic treatment")
	}
}

func TestThemePlaygroundParentDisablesShellThemeSelector(t *testing.T) {
	t.Parallel()
	if !shellConfig().Appearance.DisableThemeSelector {
		t.Fatal("documentation shell theme selector remains enabled")
	}
	if shellConfig().Appearance.PersistPreferences {
		t.Fatal("documentation shell still reads or writes stale theme preferences")
	}

	body := renderThemePlayground(t, ThemePlaygroundPage(false))
	for _, want := range []string{
		`href="/docs/theme-playground"`,
		`data-sidebar-icon="theme-playground"`,
		`src="/docs/theme-playground/frame"`,
		`title="Theme playground chart preview"`,
		`data-search="theme playground general theme-playground themes appearance picker isolated iframe live static vector interactive"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("theme playground parent missing %q", want)
		}
	}
	if strings.Contains(body, `id="componentdocshell-theme-trigger"`) {
		t.Fatal("documentation shell still renders the topbar theme selector")
	}
}

func TestThemePlaygroundFrameOwnsNonPersistentPickerAndFourCharts(t *testing.T) {
	t.Parallel()
	body := renderThemePlayground(t, ThemePlaygroundFrame())
	for _, want := range []string{
		`<html lang="en" data-theme="araihu"`,
		`componentDocShell({&#34;persist&#34;:false,&#34;theme&#34;:&#34;araihu&#34;,&#34;colorScheme&#34;:&#34;light&#34;})`,
		`id="theme-playground-theme-trigger"`,
		`aria-label="Theme"`,
		`data-theme-playground-grid`,
		`Static line`,
		`Static pie`,
		`Interactive bar`,
		`Interactive radar`,
		`data-goshtoso-charts-theme-runtime`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("theme playground frame missing %q", want)
		}
	}
	if got := strings.Count(body, `data-theme-playground-chart=`); got != 4 {
		t.Errorf("theme playground charts = %d, want 4", got)
	}
	if got := strings.Count(body, `data-theme-playground-chart="vector"`); got != 2 {
		t.Errorf("static/vector playground charts = %d, want 2", got)
	}
	if got := strings.Count(body, `data-theme-playground-chart="interactive"`); got != 2 {
		t.Errorf("interactive playground charts = %d, want 2", got)
	}
	for _, unwanted := range []string{"localStorage", "component-doc-shell__header", "componentdocshell-sidebar"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("isolated frame contains parent-shell concern %q", unwanted)
		}
	}
}

func renderThemePlayground(t *testing.T, component Page) string {
	t.Helper()
	var output bytes.Buffer
	if err := component.Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	return output.String()
}
