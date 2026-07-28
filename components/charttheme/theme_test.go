package charttheme

import "testing"

func TestStylePrecedenceAndClasses(t *testing.T) {
	style := Style{Palette: PalettePastel, Colors: []string{"#123456"}, Class: "ring-2 custom-chart"}
	if got := style.SeriesColor(0); got != "#123456" {
		t.Fatalf("explicit color = %q", got)
	}
	if got := style.SeriesColor(1); got != "var(--goshtoso-charts-series-2)" {
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
