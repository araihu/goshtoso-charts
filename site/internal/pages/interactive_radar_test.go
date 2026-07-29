package pages

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/interactive"
)

func TestInteractiveRadarUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	if interactiveRadarUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" ||
		interactiveRadarUpstreamPath != "examples/radar.go" ||
		interactiveRadarUpstreamSHA256 != "f6b8e26399826e7f979717fbb4a30b48a8c8d10e8f496da60c430aaadc0e8ffb" {
		t.Fatal("interactive Radar upstream source pin changed")
	}
	wantBehaviors := []string{"radarBase", "radarStyle", "radarLegendMulti", "radarLegendSingle"}
	coverage := interactiveRadarUpstreamCoverage()
	if len(coverage) != len(wantBehaviors) {
		t.Fatalf("coverage count = %d, want %d", len(coverage), len(wantBehaviors))
	}
	for index, name := range wantBehaviors {
		if coverage[index].Name != name || coverage[index].Treatment == "" {
			t.Errorf("coverage[%d] = %#v, want %q", index, coverage[index], name)
		}
	}
	wantFunctions := []interactiveRadarSourceFunction{
		{Name: "generateRadarItems", SHA256: "f906e8292d6830bb7983f954d67de496fd21b78dc709dbc86633f1f38d6435ae", Role: "data adaptation"},
		{Name: "radarBase", SHA256: "e897284229b2e8a01ac1a57a65a4e779c9375083bb6fad2899ab785ce7e808d7", Role: "example"},
		{Name: "radarStyle", SHA256: "45738725345bf456020df60ea819afbb8931d079cfba95153dd9cfb77b529eaa", Role: "example"},
		{Name: "radarLegendMulti", SHA256: "692e14a0b753d77b4ea4bf47aa95192d69d42ebc82319e2e47378406b507cb59", Role: "example"},
		{Name: "radarLegendSingle", SHA256: "503fc8e155fc5a43597ed9fa8a015be1f743ffb865e94fb705f4eb9b2d48a528", Role: "example"},
		{Name: "RadarExamples.Examples", SHA256: "e5c8ddab877b5227eec0975bcdf4b36531b5212b20d058e4d785d6f99b5e91d8", Role: "page composition only"},
	}
	if got := interactiveRadarSourceFunctions(); !reflect.DeepEqual(got, wantFunctions) {
		t.Fatalf("source function inventory = %#v, want %#v", got, wantFunctions)
	}
}

func TestInteractiveRadarSamplesPreserveEveryUpstreamDatasetAndTreatment(t *testing.T) {
	t.Parallel()
	base := sampleInteractiveRadarBase()
	style := sampleInteractiveRadarStyle()
	multiple := sampleInteractiveRadarLegendMulti()
	single := sampleInteractiveRadarLegendSingle()

	for _, config := range []interactive.RadarConfig{base, style, multiple, single} {
		if !reflect.DeepEqual(config.Indicators, interactiveRadarIndicators) {
			t.Errorf("%q indicators = %#v", config.Label, config.Indicators)
		}
		if config.Width != "100%" || config.Height != "520px" || config.Style.Class != "max-w-5xl mx-auto" {
			t.Errorf("%q responsive geometry = %#v", config.Label, config)
		}
		if !config.Options.Controls.Fullscreen || config.Options.Export == nil {
			t.Errorf("%q wrapper controls = %#v", config.Label, config.Options)
		}
	}
	if len(base.Series) != 1 || base.Series[0].Name != "Beijing" || len(base.Series[0].Data) != 21 {
		t.Fatalf("base series = %#v", base.Series)
	}
	if base.Coordinate.SplitArea == nil || !*base.Coordinate.SplitArea || base.Coordinate.SplitLine == nil || base.Coordinate.SplitLine.Show == nil || !*base.Coordinate.SplitLine.Show {
		t.Fatalf("base coordinate = %#v", base.Coordinate)
	}
	if style.Coordinate.Shape != interactive.RadarShapeCircle || style.Coordinate.SplitNumber != 5 ||
		style.Coordinate.SplitArea != nil ||
		style.SeriesOptions.LineStyle == nil || style.SeriesOptions.LineStyle.Opacity == nil || *style.SeriesOptions.LineStyle.Opacity != 0.5 ||
		style.SeriesOptions.AreaStyle == nil || style.SeriesOptions.AreaStyle.Opacity == nil || *style.SeriesOptions.AreaStyle.Opacity != 0.1 {
		t.Fatalf("style treatment = %#v", style)
	}
	if len(multiple.Series) != 3 || multiple.Options.Legend == nil || multiple.Options.Legend.SelectionMode != interactive.LegendSelectionMultiple {
		t.Fatalf("multiple legend treatment = %#v", multiple)
	}
	if len(single.Series) != 3 || single.Options.Legend == nil || single.Options.Legend.SelectionMode != interactive.LegendSelectionSingle ||
		single.SeriesOptions.AreaStyle == nil || single.SeriesOptions.AreaStyle.Opacity == nil || *single.SeriesOptions.AreaStyle.Opacity != 0.5 {
		t.Fatalf("single legend treatment = %#v", single)
	}

	for city, values := range map[string][][]float64{"Beijing": interactiveRadarBeijing, "Guangzhou": interactiveRadarGuangzhou, "Shanghai": interactiveRadarShanghai} {
		if len(values) != 21 {
			t.Errorf("%s observation count = %d, want 21", city, len(values))
		}
		for index, value := range values {
			if len(value) != 6 {
				t.Errorf("%s day %d dimension count = %d, want 6", city, index+1, len(value))
			}
		}
	}
	if !reflect.DeepEqual(interactiveRadarBeijing[0], []float64{55, 9, 56, 0.46, 18, 6}) ||
		!reflect.DeepEqual(interactiveRadarBeijing[20], []float64{39, 15, 36, 0.61, 29, 13}) ||
		!reflect.DeepEqual(interactiveRadarGuangzhou[0], []float64{26, 37, 27, 1.163, 27, 13}) ||
		!reflect.DeepEqual(interactiveRadarShanghai[20], []float64{87, 63, 101, 0.9, 56, 41}) {
		t.Fatal("Radar source datasets changed")
	}
}

func TestInteractiveRadarCoverageUsesCanonicalLedger(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/upstream-example-coverage.md")
	if err != nil {
		t.Fatalf("read canonical upstream coverage ledger: %v", err)
	}
	section := canonicalLedgerSection(t, string(data), "## Interactive Radar")
	for _, want := range []string{
		interactiveRadarUpstreamRevision, interactiveRadarUpstreamPath, interactiveRadarUpstreamSHA256,
		"all four upstream behavior functions", "renderer-neutral `interactive.Radar` component",
		"seventh value", "day index", "Goshtoso theme tokens", "Unsupported dedicated Radar-family behaviors: none",
	} {
		if !strings.Contains(section, want) {
			t.Errorf("canonical ledger missing interactive Radar evidence %q", want)
		}
	}
	for _, entry := range interactiveRadarUpstreamCoverage() {
		if count := strings.Count(section, "| `"+entry.Name+"` | Example |"); count != 1 {
			t.Errorf("Radar behavior row %q occurs %d times, want 1", entry.Name, count)
		}
	}
	for _, function := range interactiveRadarSourceFunctions() {
		if count := strings.Count(section, "`"+function.SHA256+"`"); count != 1 {
			t.Errorf("Radar function hash %q occurs %d times, want 1", function.SHA256, count)
		}
	}
}
