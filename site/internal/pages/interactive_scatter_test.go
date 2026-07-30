package pages

import (
	"reflect"
	"testing"

	"github.com/araihu/goshtoso-charts/components/chart"
	interactivescatter "github.com/araihu/goshtoso-charts/components/interactive/scatter"
)

func TestInteractiveScatterUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	if interactiveScatterUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" ||
		interactiveScatterUpstreamPath != "examples/scatter.go" || interactiveScatterUpstreamSHA256 != "a77ddbf7580210a842a3e1d3966ab62c3f229fdb1a33df8f319ef029bd4188b5" ||
		interactiveEffectScatterUpstreamPath != "examples/effectscatter.go" || interactiveEffectScatterUpstreamSHA256 != "1bf49dc5fb02b248ff6794aa549836b4c8fa02ddb89be6adc0c4574327673f1a" {
		t.Fatal("interactive Scatter upstream source pin changed")
	}
	want := []string{"scatterBase", "scatterShowLabel", "scatterSplitLine", "esBase", "esEffectStyle"}
	coverage := interactiveScatterUpstreamCoverage()
	if len(coverage) != len(want) {
		t.Fatalf("coverage count = %d, want %d", len(coverage), len(want))
	}
	for index, name := range want {
		if coverage[index].Name != name || coverage[index].Treatment == "" {
			t.Errorf("coverage[%d] = %#v, want %q", index, coverage[index], name)
		}
	}
	if len(interactiveScatterSourceFunctions()) != 9 {
		t.Fatalf("source function inventory count = %d, want 9", len(interactiveScatterSourceFunctions()))
	}
	for _, function := range interactiveScatterSourceFunctions() {
		if function.Path == "" || function.Name == "" || len(function.SHA256) != 64 || function.Role == "" {
			t.Errorf("incomplete source function inventory entry: %#v", function)
		}
	}
	if len(interactiveScatterSupplementarySources()) != 3 {
		t.Fatalf("supplementary source count = %d, want 3", len(interactiveScatterSupplementarySources()))
	}
}

func TestInteractiveScatterSamplesPreserveUpstreamSemantics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		config  interactivescatter.Config
		variant interactivescatter.Variant
		values  [][]float64
	}{
		{sampleInteractiveScatter(), interactivescatter.VariantStandard, [][]float64{{81, 87, 47, 59, 81, 18}, {25, 40, 56, 0, 94, 11}}},
		{sampleInteractiveScatterLabels(), interactivescatter.VariantStandard, [][]float64{{62, 89, 28, 74, 11, 45}, {37, 6, 95, 66, 28, 58}}},
		{sampleInteractiveScatterSplitLines(), interactivescatter.VariantStandard, [][]float64{{47, 47, 87, 88, 90, 15}, {41, 8, 87, 31, 29, 56}}},
		{sampleInteractiveEffectScatter(), interactivescatter.VariantEffect, [][]float64{{37, 31, 85, 26, 13, 90}}},
		{sampleInteractiveEffectScatterStyles(), interactivescatter.VariantEffect, [][]float64{{94, 63, 33, 47, 78, 24}, {59, 53, 57, 21, 89, 99}}},
	}
	for _, test := range tests {
		if test.config.Variant != test.variant {
			t.Errorf("%q variant = %q, want %q", test.config.Label, test.config.Variant, test.variant)
		}
		for seriesIndex, want := range test.values {
			got := make([]float64, len(test.config.Series[seriesIndex].Data))
			for pointIndex, point := range test.config.Series[seriesIndex].Data {
				got[pointIndex] = point.Value
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%q series %d values = %v, want %v", test.config.Label, seriesIndex, got, want)
			}
		}
	}
	base := sampleInteractiveScatter()
	if !reflect.DeepEqual(base.XAxis, interactiveScatterSports) || base.Series[0].Data[0].Symbol != "roundRect" || base.Series[0].Data[0].SymbolSize != 20 || base.Series[0].Data[0].SymbolRotate != 10 {
		t.Errorf("basic Scatter geometry = %#v", base)
	}
	labels := sampleInteractiveScatterLabels()
	if labels.SeriesOptions.Label == nil || labels.SeriesOptions.Label.Show == nil || !*labels.SeriesOptions.Label.Show || labels.SeriesOptions.Label.Position != "right" {
		t.Errorf("label treatment = %#v", labels.SeriesOptions.Label)
	}
	split := sampleInteractiveScatterSplitLines()
	if split.Options.XAxis == nil || split.Options.YAxis == nil || split.Options.XAxis.Name != "Sports" || split.Options.YAxis.Name != "Score" || split.Options.XAxis.ShowSplitLine == nil || !*split.Options.XAxis.ShowSplitLine {
		t.Errorf("split-line treatment = %#v", split.Options)
	}
	styles := sampleInteractiveEffectScatterStyles()
	if !reflect.DeepEqual(styles.Series[0].Ripple, &chart.RippleOptions{Period: 4, Scale: 10, BrushType: "stroke"}) || !reflect.DeepEqual(styles.Series[1].Ripple, &chart.RippleOptions{Period: 3, Scale: 6, BrushType: "fill"}) {
		t.Errorf("effect ripple treatments = %#v", styles.Series)
	}
}
