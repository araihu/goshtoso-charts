package pages

import (
	"reflect"
	"testing"

	"github.com/araihu/goshtoso-charts/components/chart"
	interactivepie "github.com/araihu/goshtoso-charts/components/interactive/pie"
)

func TestInteractivePieUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	if interactivePieUpstreamPath != "examples/pie.go" ||
		interactivePieUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" ||
		interactivePieUpstreamSHA256 != "a59bb6f11818d4175d033f025f00a58e6a191eff5acf30f0e0cd5f98cd493ada" {
		t.Fatal("interactive Pie upstream source pin changed")
	}
	wantNames := []string{
		"pieBase", "pieShowLabel", "pieRadius", "pieRadiusWithPadAngle", "pieRoseArea",
		"pieRoseRadius", "pieRoseAreaRadius", "pieInPie", "pieWithDispatchAction",
	}
	coverage := interactivePieUpstreamCoverage()
	if len(coverage) != len(wantNames) {
		t.Fatalf("interactive Pie coverage count = %d, want %d", len(coverage), len(wantNames))
	}
	for index, name := range wantNames {
		if coverage[index].Name != name || coverage[index].Status != pieCoverageExample || coverage[index].Treatment == "" {
			t.Errorf("coverage[%d] = %#v, want supported %q", index, coverage[index], name)
		}
	}
	if got := interactivePieSupplementarySources(); len(got) != 3 {
		t.Fatalf("supplementary source count = %d, want 3", len(got))
	}
}

func TestInteractivePieSamplesPreserveUpstreamSemantics(t *testing.T) {
	t.Parallel()
	values := []struct {
		cfg  interactivepie.Config
		want [][]float64
	}{
		{sampleInteractivePie(), [][]float64{{81, 87, 47, 59}}},
		{sampleInteractivePieLabels(), [][]float64{{81, 18, 25, 40}}},
		{sampleInteractivePieRadius(), [][]float64{{56, 0, 94, 11}}},
		{sampleInteractivePiePadded(), [][]float64{{62, 89, 28, 74}}},
		{sampleInteractivePieRoseArea(), [][]float64{{11, 45, 37, 6}}},
		{sampleInteractivePieRoseRadius(), [][]float64{{95, 66, 28, 58}}},
		{sampleInteractivePiePairedRoses(), [][]float64{{47, 47, 87, 88}, {90, 15, 41, 8}}},
		{sampleInteractivePieNested(), [][]float64{{87, 31, 29, 56}, {37, 31, 85, 26}}},
		{sampleInteractivePieAutoEmphasis(), [][]float64{{13, 90, 94, 63}}},
	}
	for _, example := range values {
		if len(example.cfg.Series) != len(example.want) {
			t.Fatalf("%q series count = %d, want %d", example.cfg.Label, len(example.cfg.Series), len(example.want))
		}
		for index := range example.cfg.Series {
			got := make([]float64, len(example.cfg.Series[index].Data))
			for point := range got {
				got[point] = example.cfg.Series[index].Data[point].Value
			}
			if !reflect.DeepEqual(got, example.want[index]) {
				t.Errorf("%q series %d values = %v, want %v", example.cfg.Label, index, got, example.want[index])
			}
		}
	}

	padded := sampleInteractivePiePadded()
	if padded.Series[0].PadAngle != 5 || !reflect.DeepEqual(padded.Series[0].Center, &interactivepie.Center{X: 40, Y: 50}) ||
		padded.Options.Legend == nil || !reflect.DeepEqual(padded.Options.Legend.Padding, &chart.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1}) ||
		padded.TooltipContent != interactivepie.TooltipNameAndShare {
		t.Errorf("padded Pie semantics = %#v", padded)
	}
	paired := sampleInteractivePiePairedRoses()
	if !reflect.DeepEqual(paired.Series[0].Center, &interactivepie.Center{X: 25, Y: 50}) ||
		!reflect.DeepEqual(paired.Series[1].Center, &interactivepie.Center{X: 75, Y: 50}) ||
		paired.Series[0].RoseMode != interactivepie.RoseArea || paired.Series[1].RoseMode != interactivepie.RoseRadius {
		t.Errorf("paired rose semantics = %#v", paired.Series)
	}
	auto := sampleInteractivePieAutoEmphasis()
	if auto.AutoEmphasis == nil || auto.AutoEmphasis.IntervalMilliseconds != 1000 || auto.Series[0].Options.Emphasis == nil {
		t.Errorf("auto-emphasis semantics = %#v", auto)
	}
	selected := sampleInteractivePieSelected()
	if !selected.Series[0].Selectable || !selected.Series[0].Data[0].Selected {
		t.Errorf("selected Pie API example = %#v", selected.Series[0])
	}
}
