package pages

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/heatmap"
)

func TestStaticHeatMapUpstreamInventoryIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	if staticHeatMapUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" {
		t.Fatalf("static HeatMap upstream revision = %q", staticHeatMapUpstreamRevision)
	}
	wantCoverage := []staticHeatMapCoverageEntry{{
		Path: "examples/1-Painter/heat_map-1-basic/main.go", SHA256: "c39a3d85a0df126da5d099a60e1491ae424d0768260ee738b7932288f1bf687f", Treatment: "Basic five-by-five matrix",
	}}
	if got := staticHeatMapUpstreamCoverage(); !reflect.DeepEqual(got, wantCoverage) {
		t.Fatalf("static HeatMap coverage = %#v, want %#v", got, wantCoverage)
	}
	wantSpans := []staticHeatMapSourceSpan{
		{Path: wantCoverage[0].Path, Lines: "14-22", Name: "writeFile", SHA256: "4582ad5f11c031d2b70604df82f08d30ff21d977b6affb87bf0ade5eb7ae88ce", Role: "Filesystem output helper; no chart behavior"},
		{Path: wantCoverage[0].Path, Lines: "24-51", Name: "main", SHA256: "227e252486af0ec1e89a2c6920253c42bc26ff2f0366907409ca9af320319fd9", Role: "Five-by-five data, centered title, named axes, and 600x400 presentation"},
		{Path: wantCoverage[0].Path, Lines: "25-31", Name: "matrix literal", SHA256: "85c0b5b518e8d8d8689b6579e1377224dba3f87f844e69a685203740d59e4705", Role: "Exact twenty-five-value matrix"},
		{Path: wantCoverage[0].Path, Lines: "33-37", Name: "chart options", SHA256: "44f2aeb070244f692526b134d79f35d364f4cf588dc30cb044d2d8f10f635182", Role: "Data binding, centered title, and X/Y axis titles"},
		{Path: wantCoverage[0].Path, Lines: "39-50", Name: "painter and output", SHA256: "a593149de21fcd6622c896759f082626f4557d2c3fb7d158a350d8bd15b4bea7", Role: "600x400 PNG rendering and filesystem delivery"},
	}
	if got := staticHeatMapSourceSpans(); !reflect.DeepEqual(got, wantSpans) {
		t.Fatalf("static HeatMap spans = %#v, want %#v", got, wantSpans)
	}
	wantAPI := []staticHeatMapAPISource{
		{Path: "heat_map.go", SHA256: "dd9b80660b9e0c0b11e5ec9ef00f5f10005a7edc1061da08f928cea15324b23c"},
		{Path: "series_label.go", SHA256: "d7b176bea3679542e878c4c5703db3711d1e258b64efb7d19614a16fc5722611"},
		{Path: "title.go", SHA256: "e85f6a0fe2e8fd7c253ac226d780164beb0cc7214e94979241d1e9fbce824b26"},
		{Path: "axis.go", SHA256: "72d8ed4c1253122a3ac49e76fbf00ff905a75aff26bcc9e4b6aea48325372a07"},
		{Path: "chart_option.go", SHA256: "0b298fcd45fab6bbe476514d90e5107d65d32bd8ba985f2411e79ec88fd2b858"},
		{Path: "painter.go", SHA256: "f4ac102e9b21623765e2fdfe4c0910a03265bc751b9f5d019ae41e80611be959"},
	}
	if got := staticHeatMapAPISources(); !reflect.DeepEqual(got, wantAPI) {
		t.Fatalf("static HeatMap API sources = %#v, want %#v", got, wantAPI)
	}
}

func TestStaticHeatMapExamplesPreserveSinglePinnedTreatmentBoundary(t *testing.T) {
	t.Parallel()
	basic := sampleBasicHeatMap()
	wantRows := [][]float64{
		{4.4, 4.9, 7.0, 7.5, 4.3}, {2.6, 5.9, 9.0, 6.4, 2.3},
		{3.3, 6.4, 7.0, 4.9, 3.2}, {1.9, 6.0, 9.0, 5.9, 2.6},
		{4.4, 5.9, 7.0, 6.4, 4.6},
	}
	if !reflect.DeepEqual(basic.Rows, wantRows) || basic.Title != "Heat Map Chart" ||
		basic.XAxis.Title != "X-Axis" || basic.YAxis.Title != "Y-Axis" || basic.Width != 600 || basic.Height != 400 {
		t.Fatalf("basic HeatMap treatment drifted: %#v", basic)
	}
	if basic.RootAttrs["data-static-heatmap-source"] != "c39a3d85a0df126d" || basic.RootAttrs["data-static-heatmap-exhaustion"] != "1fe31b06" {
		t.Fatalf("basic source markers = %#v", basic.RootAttrs)
	}
	override := sampleBasicHeatMapOverride()
	if !reflect.DeepEqual(override.Rows, wantRows) || override.Gradient.Reverse || len(override.Gradient.Stops) != 3 ||
		!override.ValueLabels.Show || override.ValueLabels.Format != heatmap.ValueFormatExact ||
		override.RootAttrs["data-static-heatmap-api"] != "dd9b80660b9e0c0b" {
		t.Fatalf("caller-style override drifted or implied another source: %#v", override)
	}
}

func TestStaticHeatMapSnippetsMatchBothPreviews(t *testing.T) {
	t.Parallel()
	for _, want := range []string{
		`Label: "Basic heat map"`, `Title: "Heat Map Chart"`, `{4.4, 4.9, 7.0, 7.5, 4.3}`, `Width: 600`, `Height: 400`,
	} {
		if !strings.Contains(heatMapCode(), want) {
			t.Errorf("basic HeatMap snippet missing %q", want)
		}
	}
	for _, want := range []string{
		`TitleOptions`, `Padding`, `ValueLabelOptions`, `ValueFormatExact`, `Decimals: 1`, `#0e7490`, `#e11d48`,
	} {
		if !strings.Contains(heatMapOverrideCode(), want) {
			t.Errorf("caller-style HeatMap snippet missing %q", want)
		}
	}
}

func TestStaticHeatMapAttributionLedgerAndPageExplainBoundary(t *testing.T) {
	t.Parallel()
	if len(chartAttributions) == 0 {
		t.Fatal("chart attributions are empty")
	}
	attribution := chartAttributions[0]
	if attribution.Name != "go-analyze/charts" || !strings.Contains(attribution.Version, staticHeatMapUpstreamRevision) ||
		!strings.Contains(attribution.UsedFor, staticHeatMapUpstreamCoverage()[0].Path) {
		t.Fatalf("static HeatMap attribution incomplete: %#v", attribution)
	}
	data, err := os.ReadFile("../../../docs/upstream-example-coverage.md")
	if err != nil {
		t.Fatalf("read coverage ledger: %v", err)
	}
	section := canonicalLedgerSection(t, string(data), "## Static/vector Heat map")
	for _, entry := range staticHeatMapUpstreamCoverage() {
		for _, want := range []string{"`" + entry.Path + "`", "`" + entry.SHA256 + "`"} {
			if !strings.Contains(section, want) {
				t.Errorf("static HeatMap ledger missing %q", want)
			}
		}
	}
	for _, span := range staticHeatMapSourceSpans() {
		for _, want := range []string{span.Lines, "`" + span.SHA256 + "`"} {
			if !strings.Contains(section, want) {
				t.Errorf("static HeatMap source span missing %q", want)
			}
		}
	}
	for _, source := range staticHeatMapAPISources() {
		for _, want := range []string{"`" + source.Path + "`", "`" + source.SHA256 + "`"} {
			if !strings.Contains(section, want) {
				t.Errorf("static HeatMap API source missing %q", want)
			}
		}
	}
	for _, want := range []string{
		"exactly one dedicated HeatMap-family file", "one distinct upstream treatment",
		"not counted as a second upstream example", "Unsupported dedicated HeatMap-family behaviors: none",
		"Arbitrary label callbacks", "raw theme", "painter", "output-encoder",
	} {
		if !strings.Contains(section, want) {
			t.Errorf("static HeatMap ledger missing boundary %q", want)
		}
	}

	source, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read pages.templ: %v", err)
	}
	page := string(source)
	start := strings.Index(page, "templ heatMapContent")
	if start < 0 {
		t.Fatal("cannot find static HeatMap page")
	}
	end := strings.Index(page[start:], "templ tableContent")
	if end < 0 {
		t.Fatal("cannot isolate static HeatMap page")
	}
	heatMapPage := page[start : start+end]
	for _, want := range []string{
		"AbovePreview: visualizationGuide(", "concentration, gaps, and broad patterns", "exact cell-by-cell comparison",
		"Typed presentation options", "same pinned upstream matrix", "not counted as a second upstream example",
		"Static/vector behavior", "inline 600 by 400 SVG", "print", "without JavaScript", "chart controls", "chart modes", `"heatmap"`,
	} {
		if !strings.Contains(heatMapPage, want) {
			t.Errorf("static HeatMap page missing %q", want)
		}
	}
	for _, forbidden := range []string{"Component contract", "Primitive", "Kind", "Configuration", "go-analyze", "go-echarts", "infrastructure", "operations"} {
		if strings.Contains(heatMapPage, forbidden) {
			t.Errorf("static HeatMap page retains redundant or renderer-specific copy %q", forbidden)
		}
	}
}
