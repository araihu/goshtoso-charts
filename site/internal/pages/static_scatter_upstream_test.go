package pages

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/scatter"
)

func TestStaticScatterUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	want := []staticScatterCoverageEntry{
		{Path: "examples/1-Painter/scatter_chart-1-basic/main.go", SHA256: "6bd838c49fc38d6b50be1b2c26e1845348de6a5bce3a4a7e637497b78ad61818", Treatment: "Basic categorical scatter with a missing observation"},
		{Path: "examples/1-Painter/scatter_chart-2-symbols/main.go", SHA256: "2667f6f260c63d56dcc22cb036b6b0408ea9da0f943757909d1436be7b9ad515", Treatment: "Per-series circle, diamond, square, and dot symbols"},
		{Path: "examples/1-Painter/scatter_chart-3-dense_data/main.go", SHA256: "0a50b43ccad6a96b3248d3e45e83add46e33b8b6ff98133e1f2597bdd46f49bb", Treatment: "Dense multi-value random walks with trends and maximum references"},
		{Path: "examples/1-Painter/scatter_chart-4-top_n_labels/main.go", SHA256: "cf92798819fbc010f44eaa406acabd337f16b52eec00793a10679e9c3b7cda81", Treatment: "Top-five value labels"},
		{Path: "examples/2-OptionFunc/scatter_chart-1-basic/main.go", SHA256: "a4528b8943edac99ab99f1632d328a34c64013e7551e3c13f61d1aa45844afd1", Treatment: "Basic data with circle symbols and integer formatting"},
	}
	if staticScatterUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" {
		t.Fatalf("static Scatter upstream revision = %q", staticScatterUpstreamRevision)
	}
	if got := staticScatterUpstreamCoverage(); !reflect.DeepEqual(got, want) {
		t.Fatalf("static Scatter upstream coverage = %#v, want %#v", got, want)
	}
}

func TestStaticScatterBasicAndSymbolTreatmentsPreservePinnedData(t *testing.T) {
	t.Parallel()
	basic := sampleBasicScatter()
	if basic.Title.Text != "Scatter" || basic.Title.FontSize != 16 || basic.Legend.Padding.Left != 100 || basic.Options.Symbol != scatter.SymbolDot || basic.Options.Size != 4 || basic.Width != 600 || basic.Height != 400 {
		t.Fatalf("basic Scatter presentation drifted: %#v", basic)
	}
	wantNames := []string{"Email", "Union Ads", "Video Ads", "Direct", "Search Engine"}
	if len(basic.Series) != len(wantNames) || !reflect.DeepEqual(basic.Categories, staticScatterWeekLabels) {
		t.Fatalf("basic Scatter shape drifted: %#v", basic)
	}
	for index, name := range wantNames {
		if basic.Series[index].Name != name || len(basic.Series[index].Values) != 7 {
			t.Errorf("basic Scatter series %d = %#v", index, basic.Series[index])
		}
	}
	if len(basic.Series[0].Values[3]) != 0 || !reflect.DeepEqual(basic.Series[0].Values[2], []float64{101}) || !reflect.DeepEqual(basic.Series[0].Values[4], []float64{90}) {
		t.Fatalf("basic Email missing observation drifted: %#v", basic.Series[0].Values)
	}
	if basic.Export == nil || basic.Export.Background != chartcontrol.ExportBackgroundTransparent {
		t.Fatal("basic Scatter no longer exercises transparent SVG and PNG export")
	}

	symbols := sampleSymbolScatter()
	wantSymbols := []scatter.Symbol{scatter.SymbolCircle, scatter.SymbolDiamond, scatter.SymbolSquare, scatter.SymbolDot}
	if symbols.Options.Size != 4 || len(symbols.Series) != len(wantSymbols) || symbols.Series[0].Values[3][0] != 95 {
		t.Fatalf("symbol Scatter shape drifted: %#v", symbols)
	}
	for index, symbol := range wantSymbols {
		if symbols.Series[index].Options.Symbol != symbol {
			t.Errorf("symbol Scatter series %d symbol = %q, want %q", index, symbols.Series[index].Options.Symbol, symbol)
		}
	}
}

func TestStaticScatterDenseTreatmentPreservesPinnedSemantics(t *testing.T) {
	t.Parallel()
	cfg := sampleDenseScatter()
	if len(cfg.Categories) != 1000 || cfg.Categories[0] != "foo 0" || cfg.Categories[999] != "foo 999" || len(cfg.Series) != 3 || cfg.Width != 600 || cfg.Height != 400 {
		t.Fatalf("dense Scatter shape drifted: %d categories, %d series, %dx%d", len(cfg.Categories), len(cfg.Series), cfg.Width, cfg.Height)
	}
	if cfg.Options.Size != .5 || cfg.Options.Trend.Kind != scatter.TrendSimpleMovingAverage || cfg.Options.Trend.Period != 100 || cfg.Options.ValueFormat != scatter.ValueFormatHumanized {
		t.Fatalf("dense Scatter series options drifted: %#v", cfg.Options)
	}
	if cfg.Series[0].Options.ReferenceLine != scatter.ReferenceLineMaximum || cfg.Series[1].Options.ReferenceLine != scatter.ReferenceLineMaximum || cfg.Series[2].Options.ReferenceLine != scatter.ReferenceLineNone {
		t.Fatalf("dense Scatter maximum references drifted: %#v", cfg.Series)
	}
	if cfg.Title.Text != "Dense Scatter Chart Demo" || cfg.Title.Placement != scatter.PlacementCenter || cfg.Legend.Orientation != scatter.LegendVertical || cfg.Legend.Placement != scatter.PlacementRight || cfg.Legend.Alignment != scatter.AlignmentRight || cfg.Legend.FontSize != 6 {
		t.Fatalf("dense Scatter title or legend drifted: %#v / %#v", cfg.Title, cfg.Legend)
	}
	if cfg.XAxis.LabelCount != 10 || cfg.XAxis.LabelRotation != 45 || cfg.XAxis.LabelFontSize != 6 || cfg.XAxis.BoundaryGap == nil || *cfg.XAxis.BoundaryGap || cfg.YAxis.Min == nil || *cfg.YAxis.Min != 0 || cfg.YAxis.Max == nil || *cfg.YAxis.Max != 280 || cfg.YAxis.Unit != 10 || cfg.YAxis.LabelSkip != 1 || cfg.YAxis.LabelFontSize != 6 {
		t.Fatalf("dense Scatter axes drifted: %#v / %#v", cfg.XAxis, cfg.YAxis)
	}
	if cfg.Padding != (scatter.Padding{Top: 16, Right: 32, Bottom: 16, Left: 16}) || cfg.RootAttrs["data-static-scatter-exhaustion"] != "1fe31b06" {
		t.Fatalf("dense Scatter padding or candidate marker drifted: %#v / %#v", cfg.Padding, cfg.RootAttrs)
	}
}

func TestStaticScatterOptionFunctionTreatmentUsesTypedIntegerFormat(t *testing.T) {
	t.Parallel()
	cfg := sampleIntegerScatter()
	if cfg.Options.Symbol != scatter.SymbolCircle || cfg.Options.Size != 0 || cfg.Options.ValueFormat != scatter.ValueFormatInteger || cfg.Width != 600 || cfg.Height != 400 {
		t.Fatalf("integer-format Scatter treatment drifted: %#v", cfg)
	}
	if !reflect.DeepEqual(cfg.Series, sampleBasicScatter().Series) {
		t.Fatal("integer-format treatment no longer reuses exact basic dataset")
	}
}

func TestStaticScatterPageDocumentsDecisionsAccessibilityAndCanonicalAPI(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read pages.templ: %v", err)
	}
	page := string(source)
	start := strings.Index(page, "templ scatterContent")
	if start < 0 {
		t.Fatal("cannot isolate static Scatter page")
	}
	end := strings.Index(page[start:], "templ radarContent")
	if end < 0 {
		t.Fatal("cannot isolate static Scatter page")
	}
	scatterPage := page[start : start+end]
	for _, want := range []string{
		"AbovePreview: visualizationGuide(", "Use when", "Avoid", "equivalent caller-owned table", "do not depend on color",
		"Per-series symbols", "Dense multi-value observations", "Top value labels", "Whole-number formatting",
		"Static/vector behavior", "inline SVG", "no-JavaScript", "chart controls", "chart modes", `chartDocumentation(`, `"scatter"`,
	} {
		if !strings.Contains(scatterPage, want) {
			t.Errorf("static Scatter docs missing %q", want)
		}
	}
	for _, forbidden := range []string{"componentContract(", "Component contract", "Primitive", "Kind", "Configuration", "go-analyze", "go-echarts", "Collapse"} {
		if strings.Contains(scatterPage, forbidden) {
			t.Errorf("static Scatter docs retain redundant or renderer-specific copy %q", forbidden)
		}
	}
}

func TestStaticScatterCoverageUsesCanonicalLedgerWithoutLostEntries(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/upstream-example-coverage.md")
	if err != nil {
		t.Fatalf("read canonical upstream coverage ledger: %v", err)
	}
	ledger := string(data)
	if !strings.Contains(ledger, "## Static/vector Scatter") || !strings.Contains(ledger, staticScatterUpstreamRevision) {
		t.Fatal("canonical ledger is missing pinned static Scatter section or revision")
	}
	for _, entry := range staticScatterUpstreamCoverage() {
		if count := strings.Count(ledger, "`"+entry.Path+"`"); count != 1 {
			t.Errorf("static Scatter coverage row %q occurs %d times, want 1", entry.Path, count)
		}
		if !strings.Contains(ledger, "`"+entry.SHA256+"`") {
			t.Errorf("static Scatter coverage row %q missing SHA-256 %s", entry.Path, entry.SHA256)
		}
	}
	for _, want := range []string{"all five dedicated Scatter-family files", "Unsupported dedicated Scatter-family behaviors: none", "fixed local seed", "cross-family composition"} {
		if !strings.Contains(ledger, want) {
			t.Errorf("static Scatter ledger missing boundary statement %q", want)
		}
	}
}
