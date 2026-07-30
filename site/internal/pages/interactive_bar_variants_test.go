package pages

import (
	"reflect"
	"testing"

	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
)

func TestInteractiveBarUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	if interactiveBarUpstreamPath != "examples/bar.go" ||
		interactiveBarUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" ||
		interactiveBarUpstreamSHA256 != "dcda545f978fdd055ecff5a6050b2ad9dc8cf9fe350bd7e4768952e8068fc9f9" {
		t.Fatal("interactive Bar upstream source pin changed")
	}
	wantNames := []string{
		"barBasic", "barTitle", "barTooltip", "barSetToolbox", "barShowLabel", "barXYName", "barXYFormatter",
		"barColor", "barSplitLine", "barGap", "barDataZoomInside", "barDataZoomSlider", "barReverse", "barStack",
		"barMarkPoints", "barMarkLines", "barOverlap", "barSize", "barWidth",
	}
	coverage := interactiveBarUpstreamCoverage()
	if len(coverage) != len(wantNames) {
		t.Fatalf("interactive Bar coverage count = %d, want %d", len(coverage), len(wantNames))
	}
	for index, name := range wantNames {
		if coverage[index].Name != name {
			t.Errorf("coverage[%d].Name = %q, want %q", index, coverage[index].Name, name)
		}
		if name == "barOverlap" {
			if coverage[index].Status != barCoverageUnsupported || coverage[index].Reason != "mixed Bar, Line, and Scatter composition requires a renderer-neutral composite chart API" {
				t.Errorf("barOverlap coverage = %#v", coverage[index])
			}
		} else if coverage[index].Status != barCoverageExample || coverage[index].Treatment == "" {
			t.Errorf("%s coverage = %#v", name, coverage[index])
		}
	}
	if got := interactiveBarSupplementarySources(); len(got) != 6 {
		t.Fatalf("supplementary source count = %d, want 6", len(got))
	}
}

func TestInteractiveBarSamplesPreserveUpstreamShapesDeterministically(t *testing.T) {
	t.Parallel()
	if got := interactiveBarCategories(); !reflect.DeepEqual(got, []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}) {
		t.Fatalf("categories = %#v", got)
	}
	first, second := fixedInteractiveBarData(11), fixedInteractiveBarData(11)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("fixed Bar values are nondeterministic")
	}
	for index, point := range first {
		if point.Value < 0 || point.Value >= 300 || point.Value != float64(int(point.Value)) {
			t.Errorf("point %d = %v, want integer in [0,300)", index, point.Value)
		}
	}

	tests := []interactivebar.Config{
		sampleInteractiveBar(), sampleInteractiveBarLabels(), sampleInteractiveBarAxes(), sampleInteractiveBarColors(),
		sampleInteractiveBarWidthsAndGap(), sampleInteractiveBarHorizontal(), sampleInteractiveBarStacked(),
		sampleInteractiveBarZoom(interactivebar.ZoomInside), sampleInteractiveBarZoom(interactivebar.ZoomSlider),
		sampleInteractiveBarMarkPoints(), sampleInteractiveBarMarkLines(), sampleInteractiveBarLargeCanvas(),
	}
	for _, cfg := range tests {
		if len(cfg.XAxis) != 7 || len(cfg.Series) != 2 {
			t.Errorf("%q shape = %d categories, %d series", cfg.Label, len(cfg.XAxis), len(cfg.Series))
		}
		for _, series := range cfg.Series {
			if len(series.Data) != 7 {
				t.Errorf("%q series %q value count = %d", cfg.Label, series.Name, len(series.Data))
			}
		}
	}

	if sampleInteractiveBarHorizontal().Orientation != interactivebar.OrientationHorizontal {
		t.Error("horizontal sample lost axis reversal")
	}
	if sampleInteractiveBarStacked().SeriesOptions.Stack != "stackA" {
		t.Error("stacked sample lost shared stack")
	}
	if zoom := sampleInteractiveBarZoom(interactivebar.ZoomSlider).Zoom; zoom == nil || zoom.Mode != interactivebar.ZoomSlider || zoom.StartPercent != 10 || zoom.EndPercent != 50 {
		t.Errorf("slider zoom = %#v", zoom)
	}
	if cfg := sampleInteractiveBarLargeCanvas(); cfg.Width != "100%" || cfg.Height != "600px" {
		t.Errorf("large canvas = %q by %q", cfg.Width, cfg.Height)
	}
}
