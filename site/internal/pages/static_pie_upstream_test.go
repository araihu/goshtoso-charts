package pages

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/pie"
)

func TestStaticPieUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	want := []staticPieCoverageEntry{
		{Path: "examples/1-Painter/doughnut_chart-1-basic/main.go", SHA256: "b97bca2322e90e2f03ab49aa77f683d0c58e027846b939e5a61100602dad1ebf", Treatment: "Basic doughnut presentation"},
		{Path: "examples/1-Painter/doughnut_chart-2-styles/main.go", SHA256: "5816db5dd035c8607b2929779353c32d2bca78ed5f6244b3fc04e65292ac3610", Treatment: "Outside labels, inside labels, and center total"},
		{Path: "examples/1-Painter/pie_chart-1-basic/main.go", SHA256: "06183e92e75445d89917af5dfd318c8b45f624c4efa6565b626a6aff6b3b128f", Treatment: "Basic pie presentation"},
		{Path: "examples/1-Painter/pie_chart-2-series_radius/main.go", SHA256: "54d85c6420a5e8f4fca7691c4969be80cc6bc52f8d4f10cbe5e499715875cbf6", Treatment: "Area-scaled slice radii"},
		{Path: "examples/1-Painter/pie_chart-3-gap/main.go", SHA256: "2392d1fd1a7644158626a261344e79b18bef2c3d802fa1cea8c3add413b980f6", Treatment: "Segment gap and hidden legend"},
		{Path: "examples/2-OptionFunc/doughnut_chart-1-basic/main.go", SHA256: "1936ff4508d6ef3967185e4076804bf53dc0bf8c64a254a569081fb1d399b453", Treatment: "Duplicate basic doughnut through option functions"},
		{Path: "examples/2-OptionFunc/pie_chart-1-basic/main.go", SHA256: "d09222d5febf104f07a81e05a4235d96004b61e5c032dd3513a501a840bbe9b7", Treatment: "Duplicate basic pie through option functions"},
	}
	if staticPieUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" {
		t.Fatalf("static Pie upstream revision = %q", staticPieUpstreamRevision)
	}
	if got := staticPieUpstreamCoverage(); !reflect.DeepEqual(got, want) {
		t.Fatalf("static Pie upstream coverage = %#v, want %#v", got, want)
	}
}

func TestStaticPieBasicRadiusAndGapTreatmentsPreservePinnedData(t *testing.T) {
	t.Parallel()
	wantNames := []string{"Search Engine", "Direct", "Email", "Union Ads", "Video Ads"}
	wantValues := []float64{1048, 735, 580, 484, 300}
	for _, cfg := range []pie.Config{sampleBasicPie(), sampleAreaScaledPie(), sampleSegmentGapPie(), sampleBasicDoughnutChart()} {
		if names, values := pieSliceData(cfg.Slices); !reflect.DeepEqual(names, wantNames) || !reflect.DeepEqual(values, wantValues) {
			t.Fatalf("basic dataset drifted for %q: %v / %v", cfg.Label, names, values)
		}
		if cfg.Width != 600 || cfg.Height != 400 {
			t.Fatalf("basic geometry for %q = %dx%d", cfg.Label, cfg.Width, cfg.Height)
		}
	}
	radius := sampleAreaScaledPie()
	if radius.Radius.OuterPixels != 120 || radius.Radius.Scale != pie.RadiusScaleArea {
		t.Fatalf("radius treatment = %#v", radius.Radius)
	}
	if sampleBasicPie().Export == nil || sampleBasicPie().Export.Background != chartcontrol.ExportBackgroundTransparent {
		t.Fatal("basic Pie no longer exercises shared transparent SVG and PNG export")
	}
	gap := sampleSegmentGapPie()
	if gap.SegmentGap != 16 || !gap.Legend.Hidden || gap.Title.Text != "Pie Chart With Segment Gap" {
		t.Fatalf("gap treatment = %#v", gap)
	}
	basicDoughnut := sampleBasicDoughnutChart()
	if basicDoughnut.Variant != pie.VariantDoughnut || basicDoughnut.InnerRadiusPercent != 60 || basicDoughnut.Legend.LeftPercent != 80 || basicDoughnut.Title.Subtitle != "(Fake Data)" {
		t.Fatalf("basic doughnut treatment = %#v", basicDoughnut)
	}
}

func TestStaticPieDoughnutStylesPreservePinnedTreatments(t *testing.T) {
	t.Parallel()
	wantNames := []string{"Direct", "Search Engine", "Referral", "Email", "Video Ads"}
	wantValues := []float64{1048, 735, 580, 484, 300}
	outside, inside, total := sampleDoughnutOutsideLabels(), sampleDoughnutInsideLabels(), sampleDoughnutCenterTotal()
	for _, cfg := range []pie.Config{outside, inside, total} {
		if names, values := pieSliceData(cfg.Slices); !reflect.DeepEqual(names, wantNames) || !reflect.DeepEqual(values, wantValues) {
			t.Fatalf("style dataset drifted for %q: %v / %v", cfg.Label, names, values)
		}
	}
	if outside.SegmentGap != 24 || !outside.Legend.Hidden || outside.Padding.Bottom != 15 || outside.Width != 600 || outside.Height != 400 {
		t.Fatalf("outside-label treatment = %#v", outside)
	}
	if inside.Labels.Placement != pie.LabelPlacementInside || inside.InnerRadiusPercent != 80 || inside.Width != 400 || inside.Height != 400 {
		t.Fatalf("inside-label treatment = %#v", inside)
	}
	if total.SegmentGap != 8 || !total.Labels.Hidden || total.Center.Content != pie.CenterContentTotal || total.Center.Prefix != "Total Response: " || total.Center.Format != pie.ValueFormatHumanized || total.Center.Decimals != 2 || total.Center.FontSize != 12 || !total.Legend.Overlay {
		t.Fatalf("center-total treatment = %#v", total)
	}
}

func TestStaticPiePageDocumentsDecisionsAccessibilityAndCanonicalAPI(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read pages.templ: %v", err)
	}
	page := string(source)
	start := strings.Index(page, "templ pieContent")
	end := strings.Index(page[start:], "templ scatterContent")
	if start < 0 || end < 0 {
		t.Fatal("cannot isolate static Pie page")
	}
	piePage := page[start : start+end]
	for _, want := range []string{
		"AbovePreview: visualizationGuide(", "Use when", "Avoid", "exact-value and share table",
		"Area-scaled radii", "Segment spacing", "Doughnut treatment", "Outside and inside labels", "Center total and overlay legend",
		"Static/vector behavior", "inline SVG", "chart controls", "chart modes", `chartDocumentation(`, `"pie"`,
	} {
		if !strings.Contains(piePage, want) {
			t.Errorf("static Pie docs missing %q", want)
		}
	}
	for _, forbidden := range []string{"componentContract(", "Component contract", "Primitive", "Kind", "Configuration", "go-analyze", "go-echarts", "Collapse"} {
		if strings.Contains(piePage, forbidden) {
			t.Errorf("static Pie docs retain redundant or renderer-specific copy %q", forbidden)
		}
	}
}

func TestStaticPieCoverageUsesCanonicalLedgerWithoutLostEntries(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/upstream-example-coverage.md")
	if err != nil {
		t.Fatalf("read canonical upstream coverage ledger: %v", err)
	}
	ledger := string(data)
	if !strings.Contains(ledger, "## Static/vector Pie") || !strings.Contains(ledger, staticPieUpstreamRevision) {
		t.Fatal("canonical ledger is missing pinned static Pie section or revision")
	}
	for _, entry := range staticPieUpstreamCoverage() {
		if count := strings.Count(ledger, "`"+entry.Path+"`"); count != 1 {
			t.Errorf("static Pie coverage row %q occurs %d times, want 1", entry.Path, count)
		}
		if !strings.Contains(ledger, "`"+entry.SHA256+"`") {
			t.Errorf("static Pie coverage row %q missing SHA-256 %s", entry.Path, entry.SHA256)
		}
	}
	for _, want := range []string{"seven dedicated", "seven distinct visual treatments", "Unsupported dedicated Pie-family behaviors: none", "cross-family composition", "No dedicated Pie-family source defines statistical references"} {
		if !strings.Contains(ledger, want) {
			t.Errorf("static Pie ledger missing boundary statement %q", want)
		}
	}
}

func pieSliceData(slices []pie.Slice) ([]string, []float64) {
	names := make([]string, len(slices))
	values := make([]float64, len(slices))
	for index := range slices {
		names[index], values[index] = slices[index].Name, slices[index].Value
	}
	return names, values
}
