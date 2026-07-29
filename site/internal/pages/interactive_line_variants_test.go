package pages

import (
	"reflect"
	"testing"
)

func TestInteractiveLineUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	if interactiveLineUpstreamPath != "examples/line.go" ||
		interactiveLineUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" ||
		interactiveLineUpstreamSHA256 != "1f36444bd373eafde876af19746d6b0115a776fd7c019e5996bdf2d00ecd7b1c" {
		t.Fatal("interactive Line upstream source pin changed")
	}
	want := []interactiveLineCoverageEntry{
		{Name: "lineBase", Status: lineCoverageExample},
		{Name: "lineShowLabel", Status: lineCoverageExample},
		{Name: "lineMarkPoint", Status: lineCoverageExample},
		{Name: "lineSplitLine", Status: lineCoverageExample},
		{Name: "lineNumerical", Status: lineCoverageExample},
		{Name: "lineTime", Status: lineCoverageExample},
		{Name: "lineStep", Status: lineCoverageExample},
		{Name: "lineSmooth", Status: lineCoverageExample},
		{Name: "lineArea", Status: lineCoverageExample},
		{Name: "lineSmoothArea", Status: lineCoverageExample},
		{Name: "lineOverlap", Status: lineCoverageUnsupported, Reason: "mixed-series composition requires a renderer-neutral composite chart API"},
		{Name: "lineMulti", Status: lineCoverageExample},
		{Name: "lineDemo", Status: lineCoverageExample},
		{Name: "lineSymbols", Status: lineCoverageExample},
	}
	if got := interactiveLineUpstreamCoverage(); !reflect.DeepEqual(got, want) {
		t.Fatalf("interactive Line coverage = %#v, want %#v", got, want)
	}
}

func TestInteractiveLineSamplesPreserveUpstreamShapesDeterministically(t *testing.T) {
	t.Parallel()
	if got := interactiveLineCategories(); !reflect.DeepEqual(got, []string{"Apple", "Banana", "Peach", "Lemon", "Pear", "Cherry"}) {
		t.Fatalf("categories = %#v", got)
	}
	for name, sample := range map[string]int{
		"base":           len(sampleInteractiveLineBase().Series[0].Data),
		"labels":         len(sampleInteractiveLineLabels().Series[0].Data),
		"symbols first":  len(sampleInteractiveLineSymbols().Series[0].Data),
		"symbols second": len(sampleInteractiveLineSymbols().Series[1].Data),
		"numerical":      len(sampleInteractiveLineNumerical().Series[0].Data),
		"temporal":       len(sampleInteractiveLineTime().Series[0].Data),
	} {
		want := 6
		if name == "numerical" {
			want = 30
		} else if name == "temporal" {
			want = 50
		}
		if sample != want {
			t.Errorf("%s data count = %d, want %d", name, sample, want)
		}
	}
	first := sampleInteractiveLineNumerical().Series[0].Data
	second := sampleInteractiveLineNumerical().Series[0].Data
	if !reflect.DeepEqual(first, second) {
		t.Fatal("numerical substitute is nondeterministic")
	}
}
