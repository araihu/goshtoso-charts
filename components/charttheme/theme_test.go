package charttheme

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestStylePrecedenceAndClasses(t *testing.T) {
	style := Style{Palette: PalettePastel, Colors: []string{"#123456"}, Class: "ring-2 custom-chart"}
	if got := style.SeriesColor(0); got != "#123456" {
		t.Fatalf("explicit color = %q", got)
	}
	if got := style.SeriesColor(1); got != "var(--color-chart-series-2)" {
		t.Fatalf("fallback token = %q", got)
	}
	if got := style.RootClasses("chart"); got != "chart goshtoso-charts-palette goshtoso-charts-palette-pastel ring-2 custom-chart" {
		t.Fatalf("classes = %q", got)
	}
	colors := style.ResolvedColors()
	if len(colors) != 8 || colors[0] != "#123456" || colors[1] != "#fca5a5" {
		t.Fatalf("resolved colors = %#v", colors)
	}
}

func TestAutoUsesBoldLiteralFallback(t *testing.T) {
	colors := (Style{}).ResolvedColors()
	if len(colors) != 8 || colors[0] != "#2563eb" {
		t.Fatalf("auto fallback = %#v", colors)
	}
	if got := (Style{}).RootClasses("chart"); got != "chart goshtoso-charts-palette goshtoso-charts-palette-auto" {
		t.Fatalf("auto classes = %q", got)
	}
}

func TestStatusPaletteUsesSemanticOrderAndRootClass(t *testing.T) {
	t.Parallel()
	style := Style{Palette: PaletteStatus, Colors: []string{"#123456"}, Class: "custom-status-chart"}
	colors := style.ResolvedColors()
	if len(colors) != 8 || colors[0] != "#123456" || colors[1] != "#d97706" || colors[2] != "#dc2626" {
		t.Fatalf("status colors = %#v", colors)
	}
	if got := style.RootClasses("chart"); got != "chart goshtoso-charts-palette goshtoso-charts-palette-status custom-status-chart" {
		t.Fatalf("status classes = %q", got)
	}
}

func TestStylesExposeLightAndDarkSemanticChartTokens(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := Styles().Render(context.Background(), &output); err != nil {
		t.Fatalf("Styles().Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`--color-chart-surface: var(--color-surface)`,
		`--color-chart-series-1: var(--color-blue-600, #2563eb)`,
		`--goshtoso-charts-series-1: var(--color-chart-series-1)`,
		`--goshtoso-charts-surface: var(--color-chart-surface)`,
		`--color-chart-surface-alt: var(--color-surface-alt, var(--color-surface))`,
		`--color-chart-outline: var(--color-outline)`,
		`--color-chart-grid: color-mix`,
		`--color-chart-text: var(--color-on-surface)`,
		`--color-chart-text-strong: var(--color-on-surface-strong`,
		`--color-chart-text-muted: var(--color-on-surface-muted`,
		`--color-chart-scale-low:`,
		`--color-chart-scale-mid:`,
		`--color-chart-scale-high:`,
		`--color-chart-scale-low: var(--color-cyan-300, #67e8f9)`,
		`--color-chart-scale-mid: var(--color-amber-400, #fbbf24)`,
		`--color-chart-scale-high: var(--color-red-600, #dc2626)`,
		`--color-chart-scale-low: var(--color-cyan-700, #0e7490)`,
		`--color-chart-scale-high: var(--color-red-400, #f87171)`,
		`.goshtoso-charts-palette-status`,
		`--color-chart-series-1: color-mix(in srgb, var(--color-success, var(--color-green-600, #16a34a)) 80%`,
		`--color-chart-series-2: color-mix(in srgb, var(--color-warning, var(--color-amber-600, #d97706)) 80%`,
		`--color-chart-series-3: color-mix(in srgb, var(--color-danger, var(--color-red-600, #dc2626)) 80%`,
		`.dark .goshtoso-charts-palette-status`,
		`--color-chart-series-1: color-mix(in srgb, var(--color-success, var(--color-green-400, #4ade80)) 80%`,
		`.dark .goshtoso-charts-palette`,
		`--color-chart-surface: var(--color-surface-dark)`,
		`--color-chart-text-strong: var(--color-on-surface-dark-strong`,
		`--color-chart-text-muted: var(--color-on-surface-dark-muted`,
		`[data-theme="araihu"] .goshtoso-charts-palette-auto`,
		`--color-chart-scale-low: var(--color-sky-400`,
		`--color-chart-scale-mid: var(--color-amber-500`,
		`--color-chart-scale-high: var(--color-rose-600`,
		`--color-chart-scale-high: var(--color-rose-400`,
		`--color-chart-scale-mid: var(--color-amber-200`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("theme styles missing %q", want)
		}
	}
}
