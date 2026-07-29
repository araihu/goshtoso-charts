package pages

import (
	"reflect"
	"testing"

	"github.com/araihu/goshtoso-charts/components/bar"
)

func TestStaticBarUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	want := []staticBarCoverageEntry{
		{Path: "examples/1-Painter/bar_chart-1-basic/main.go", SHA256: "30f03d99e6f1394096d18ebed5edc04210096945c163719ce38fdd4c97368383", Treatment: "Basic vertical comparison"},
		{Path: "examples/1-Painter/bar_chart-2-size_margin/main.go", SHA256: "fc0d426db8e09dc2032bd0c8f6bf67dbcdd66668bac0547821e4094006a55c7b", Treatment: "Vertical thickness and group gap comparison"},
		{Path: "examples/1-Painter/bar_chart-3-label_position-round_caps/main.go", SHA256: "b29d5387ab867885d627c565ce11c2284dd046272a43a86dabb83549363696ce", Treatment: "Rounded caps and start/end value labels"},
		{Path: "examples/1-Painter/bar_chart-4-mark/main.go", SHA256: "544fea22c29db4225c7b10bb6d12137d484a4ca9b6c647dc29730a61ce4ced4c", Treatment: "Average lines and minimum/maximum points"},
		{Path: "examples/1-Painter/bar_chart-5-stacked/main.go", SHA256: "d35e6866b6e4d4071d3e09e26db77165fbd9aecefecf0e6e880bb35752e23119", Treatment: "Stacked totals, maximum line, and global maximum point"},
		{Path: "examples/1-Painter/horizontal_bar_chart-1-basic/main.go", SHA256: "735240dd8433bd2494ae019f272840a8ff2fcf5572166b78269e23cbff7111a0", Treatment: "Basic horizontal comparison"},
		{Path: "examples/1-Painter/horizontal_bar_chart-2-size_margin/main.go", SHA256: "34d9f682f5168830cfac75a4f35a4bcc2e8216ea25024142be892bc7784597da", Treatment: "Horizontal thickness and group gap comparison"},
		{Path: "examples/1-Painter/horizontal_bar_chart-3-mark/main.go", SHA256: "c2bd6eaf3f47d8bce333186aa0212204fe3c6ff67cc66236fc6fe1040d180cb4", Treatment: "Horizontal maximum reference lines"},
		{Path: "examples/1-Painter/horizontal_bar_chart-4-stacked/main.go", SHA256: "82ff73196a355aeda6ddb12b4e6d6cfc2d191dc32dfcd6e1830eb69a321b7b24", Treatment: "Stacked horizontal bars with exact value labels"},
		{Path: "examples/2-OptionFunc/bar_chart-1-basic/main.go", SHA256: "eeff9689b6279ecbbdf9475a8034e8f57aa582c29e8d74db74f34cf78b385ca8", Treatment: "Basic vertical comparison and references through option functions"},
		{Path: "examples/2-OptionFunc/horizontal_bar_chart-1-basic/main.go", SHA256: "84268867a32a2a4ff81cf29d2a56a48174d7c1d46c5edc8ac20a82c07e13fea4", Treatment: "Basic horizontal comparison through option functions"},
	}
	if staticBarUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" {
		t.Fatalf("static Bar upstream revision = %q", staticBarUpstreamRevision)
	}
	if got := staticBarUpstreamCoverage(); !reflect.DeepEqual(got, want) {
		t.Fatalf("static Bar upstream coverage = %#v, want %#v", got, want)
	}
}

func TestStaticBarGeometryAndLabelExamplesPreservePinnedTreatments(t *testing.T) {
	t.Parallel()
	defaultChart, thin, noGap := sampleBarGeometryComparison(false)
	for _, cfg := range []bar.Config{defaultChart, thin, noGap} {
		if !reflect.DeepEqual(cfg.Labels, []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}) || len(cfg.Series) != 2 || cfg.Width != 400 || cfg.Height != 400 {
			t.Fatalf("geometry sample shape drifted: %#v", cfg)
		}
	}
	if defaultChart.Geometry != (bar.GeometryOptions{}) || thin.Geometry.ThicknessRatio != 0.15 || noGap.Geometry.GapRatio == nil || *noGap.Geometry.GapRatio != 0 {
		t.Fatalf("geometry treatments = default %#v thin %#v no-gap %#v", defaultChart.Geometry, thin.Geometry, noGap.Geometry)
	}
	horizontal, horizontalThin, horizontalNoGap := sampleBarGeometryComparison(true)
	if horizontal.Orientation != bar.OrientationHorizontal || horizontalThin.Orientation != bar.OrientationHorizontal || horizontalNoGap.Orientation != bar.OrientationHorizontal {
		t.Fatalf("horizontal geometry comparison lost orientation")
	}
	end, start := sampleRoundedBarLabels(bar.DataLabelPositionEnd), sampleRoundedBarLabels(bar.DataLabelPositionStart)
	for _, cfg := range []bar.Config{end, start} {
		if !cfg.Geometry.RoundedCaps || cfg.Geometry.GapRatio == nil || *cfg.Geometry.GapRatio != .02 || len(cfg.Series) != 2 || !cfg.Series[0].Labels.Show || cfg.Series[0].Labels.Format != bar.ValueFormatHumanized {
			t.Fatalf("rounded label treatment drifted: %#v", cfg)
		}
	}
	if start.LabelPosition != bar.DataLabelPositionStart || end.LabelPosition != bar.DataLabelPositionEnd {
		t.Fatalf("label positions = %q/%q", end.LabelPosition, start.LabelPosition)
	}
}

func TestStaticBarStackedAndHorizontalExamplesPreservePinnedReferences(t *testing.T) {
	t.Parallel()
	stacked := sampleStackedBar()
	if !stacked.Stacked || stacked.Padding.Right != 45 || !stacked.Series[0].References.MaximumLine || !stacked.Series[1].References.GlobalMaximum || stacked.Series[1].References.PointPrefix != "Sum:" || stacked.Series[1].References.PointSize != 32 {
		t.Fatalf("stacked treatment drifted: %#v", stacked)
	}
	horizontalMarked := sampleHorizontalBarReferences()
	if horizontalMarked.Orientation != bar.OrientationHorizontal || !horizontalMarked.Series[0].References.MaximumLine || !horizontalMarked.Series[1].References.MaximumLine {
		t.Fatalf("horizontal references drifted: %#v", horizontalMarked)
	}
	horizontalStacked := sampleHorizontalStackedBar()
	if !horizontalStacked.Stacked || horizontalStacked.Orientation != bar.OrientationHorizontal || !horizontalStacked.ValueAxis.Hidden || !horizontalStacked.Series[0].Labels.Show || !horizontalStacked.Series[1].Labels.Show {
		t.Fatalf("horizontal stacked treatment drifted: %#v", horizontalStacked)
	}
}
