package pages

import (
	"reflect"
	"testing"

	"github.com/araihu/goshtoso-charts/components/interactive"
)

func TestInteractiveFunnelCoverageInventoriesEveryPinnedBehaviorAndSourceSpan(t *testing.T) {
	t.Parallel()
	if interactiveFunnelUpstreamPath != "examples/funnel.go" || interactiveFunnelUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" || interactiveFunnelUpstreamSHA256 != "c532e6490bad284b4b6a5dec20825359abc795a8ee9f3bb5febbcfb4e0cd2d55" {
		t.Fatal("interactive Funnel source pin changed")
	}
	wantCoverage := []interactiveFunnelCoverageEntry{
		{Name: "funnelBase", Treatment: "basic five-stage funnel"},
		{Name: "funnelShowLabel", Treatment: "visible stage labels positioned left"},
	}
	if got := interactiveFunnelUpstreamCoverage(); !reflect.DeepEqual(got, wantCoverage) {
		t.Fatalf("interactiveFunnelUpstreamCoverage() = %#v, want %#v", got, wantCoverage)
	}
	wantSpans := []interactiveFunnelSourceSpan{
		{Name: "dimensions", Lines: "13", SHA256: "bd5b4e6a9c429f461f802686cfe0539b660b85c12fc382656d0097d07d1f7e83", Role: "ordered stage labels"},
		{Name: "genFunnelKvItems", Lines: "15–21", SHA256: "9bf2059b03ae41f499ad99ac7ece0ac9deec70fe0fc3df8052310b1f8a64ac1c", Role: "random data helper adapted to a local fixed seed"},
		{Name: "funnelBase", Lines: "22–30", SHA256: "88c9efbc1bdda11af5c7cbad673b7cd568941ed9248a6428ec5ea5c45184fd45", Role: "basic example"},
		{Name: "funnelShowLabel", Lines: "33–47", SHA256: "ed198502f5b56653897070b7ba9b7fb862a8ceb4255ac4242cd817d0be0e23d7", Role: "left-label example"},
		{Name: "FunnelExamples.Examples", Lines: "51–63", SHA256: "6c55c7a63e033a5b0f6de045c16c31d16eee67a134000a1e240da88ffbb1ca97", Role: "page composition only"},
	}
	if got := interactiveFunnelSourceSpans(); !reflect.DeepEqual(got, wantSpans) {
		t.Fatalf("interactiveFunnelSourceSpans() = %#v, want %#v", got, wantSpans)
	}
}

func TestFixedInteractiveFunnelDataPreservesOrderDomainAndHelperCallSequence(t *testing.T) {
	t.Parallel()
	want := [][]float64{{31, 37, 47, 9, 31}, {18, 25, 40, 6, 0}}
	for callIndex, wantValues := range want {
		first := fixedInteractiveFunnelData(callIndex)
		second := fixedInteractiveFunnelData(callIndex)
		if !reflect.DeepEqual(first, second) {
			t.Fatalf("fixedInteractiveFunnelData(%d) is not deterministic", callIndex)
		}
		if len(first) != len(interactiveFunnelDimensions) {
			t.Fatalf("fixedInteractiveFunnelData(%d) length = %d", callIndex, len(first))
		}
		for index, point := range first {
			if point.Name != interactiveFunnelDimensions[index] || point.Value != wantValues[index] {
				t.Errorf("call %d point %d = %#v, want %q %v", callIndex, index, point, interactiveFunnelDimensions[index], wantValues[index])
			}
			if point.Value < 0 || point.Value >= 50 {
				t.Errorf("call %d point %d value %v outside [0,50)", callIndex, index, point.Value)
			}
		}
	}
}

func TestInteractiveFunnelExamplesMapBothPinnedBehaviors(t *testing.T) {
	t.Parallel()
	base := sampleInteractiveFunnel()
	if base.Label != "Basic five-stage funnel" || base.Width != "100%" || base.Height != "420px" || base.Options.Title.Text != "basic funnel example" {
		t.Fatalf("base Funnel presentation = %#v", base)
	}
	if len(base.Series) != 1 || base.Series[0].Name != "Analytics" || !reflect.DeepEqual(base.Series[0].Data, fixedInteractiveFunnelData(0)) {
		t.Fatalf("base Funnel series = %#v", base.Series)
	}
	if base.Style.Class != "max-w-5xl mx-auto" || !base.Options.Controls.Fullscreen || base.Options.Export == nil {
		t.Fatalf("base Funnel layout or controls = %#v %#v", base.Style, base.Options)
	}

	labels := sampleInteractiveFunnelLabels()
	if labels.Options.Title.Text != "show label" || !reflect.DeepEqual(labels.Series[0].Data, fixedInteractiveFunnelData(1)) {
		t.Fatalf("label Funnel data or title = %#v", labels)
	}
	if labels.SeriesOptions.Label == nil || labels.SeriesOptions.Label.Show == nil || !*labels.SeriesOptions.Label.Show || labels.SeriesOptions.Label.Position != "left" {
		t.Fatalf("label Funnel options = %#v", labels.SeriesOptions.Label)
	}
	for _, cfg := range []interactive.FunnelConfig{base, labels} {
		if cfg.Options.Legend == nil || cfg.Options.Legend.Bottom != "0" || cfg.Options.Tooltip == nil || cfg.Options.Tooltip.Trigger != "item" {
			t.Errorf("contained Funnel options = %#v", cfg.Options)
		}
	}
}
