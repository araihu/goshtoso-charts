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
		`.dark .goshtoso-charts-palette`,
		`--color-chart-surface: var(--color-surface-dark)`,
		`--color-chart-text-strong: var(--color-on-surface-dark-strong`,
		`--color-chart-text-muted: var(--color-on-surface-dark-muted`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("theme styles missing %q", want)
		}
	}
}
