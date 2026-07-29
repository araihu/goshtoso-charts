package pages

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/radar"
)

func TestStaticRadarUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	want := []staticRadarCoverageEntry{
		{Path: "examples/1-Painter/radar_chart-1-basic/main.go", SHA256: "0cf8dbdd72f6a398b7c560b544a0d800570d17a620549753b726171f996254d4", Treatment: "Basic budget comparison through the Painter API"},
		{Path: "examples/2-OptionFunc/radar_chart-1-basic/main.go", SHA256: "39a9427d6bb3bcff7d7627943210e2b19b934785c03896e9beffb1a35462c78e", Treatment: "Same basic budget comparison through option functions"},
	}
	if staticRadarUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" {
		t.Fatalf("static Radar upstream revision = %q", staticRadarUpstreamRevision)
	}
	if got := staticRadarUpstreamCoverage(); !reflect.DeepEqual(got, want) {
		t.Fatalf("static Radar upstream coverage = %#v, want %#v", got, want)
	}
	wantAPISources := []staticRadarAPISource{
		{Path: "radar_chart.go", SHA256: "50f9e29787665a03ab744be3081b9582cbb2d4064025245818665923849e79d7"},
		{Path: "series.go", SHA256: "953f4e5d555701348ebcb8eb0bfe1753a6df56eb4f94a86403c6dc6cecf79217"},
		{Path: "series_label.go", SHA256: "d7b176bea3679542e878c4c5703db3711d1e258b64efb7d19614a16fc5722611"},
		{Path: "title.go", SHA256: "e85f6a0fe2e8fd7c253ac226d780164beb0cc7214e94979241d1e9fbce824b26"},
		{Path: "legend.go", SHA256: "eaad1144ff1c5af84049a2968e952caba2dac3970558d542c18c6a98ba6d515b"},
		{Path: "chart_option.go", SHA256: "0b298fcd45fab6bbe476514d90e5107d65d32bd8ba985f2411e79ec88fd2b858"},
	}
	if got := staticRadarAPISources(); !reflect.DeepEqual(got, wantAPISources) {
		t.Fatalf("static Radar API sources = %#v, want %#v", got, wantAPISources)
	}
}

func TestStaticRadarBasicTreatmentPreservesPinnedDataAndPresentation(t *testing.T) {
	t.Parallel()
	cfg := sampleBasicRadar()
	if cfg.Title.Text != "Basic Radar Chart" || cfg.Title.FontSize != 16 || cfg.Legend.Horizontal != radar.PlacementEnd || cfg.Width != 600 || cfg.Height != 400 {
		t.Fatalf("basic Radar presentation drifted: %#v", cfg)
	}
	wantIndicators := []radar.Indicator{
		{Name: "Sales", Max: 6500},
		{Name: "Administration", Max: 16000},
		{Name: "Information Technology", Max: 30000},
		{Name: "Customer Support", Max: 38000},
		{Name: "Development", Max: 52000},
		{Name: "Marketing", Max: 25000},
	}
	wantSeries := []radar.Series{
		{Name: "Allocated Budget", Values: []float64{4200, 3000, 20000, 35000, 50000, 18000}},
		{Name: "Actual Spending", Values: []float64{5000, 14000, 28000, 26000, 42000, 21000}},
	}
	if !reflect.DeepEqual(cfg.Indicators, wantIndicators) || !reflect.DeepEqual(cfg.Series, wantSeries) {
		t.Fatalf("basic Radar data drifted: %#v / %#v", cfg.Indicators, cfg.Series)
	}
	if cfg.RootAttrs["data-static-radar-exhaustion"] != "1fe31b06" || cfg.RootAttrs["data-goshtoso-candidate"] != "radar-basic-0cf8dbdd72f6a398" {
		t.Fatalf("basic Radar provenance markers = %#v", cfg.RootAttrs)
	}
}

func TestStaticRadarTypedOptionsExerciseRelevantUpstreamSurface(t *testing.T) {
	t.Parallel()
	cfg := sampleReadableRadar()
	if cfg.Options.RadiusPercent != 44 || cfg.Options.ValueLabels != radar.ValueLabelsShown || cfg.Options.ValueFormat != radar.ValueFormatHumanized {
		t.Fatalf("Radar options = %#v", cfg.Options)
	}
	if cfg.Title.Horizontal != radar.PlacementCenter || cfg.Title.FontSize != 16 || cfg.Title.SubtextFontSize != 12 || cfg.Legend.Orientation != radar.LegendVertical || cfg.Legend.Horizontal != radar.PlacementEnd || cfg.Legend.Alignment != radar.AlignmentEnd || !cfg.Legend.Overlay {
		t.Fatalf("Radar title/legend = %#v / %#v", cfg.Title, cfg.Legend)
	}
	if cfg.Padding != (radar.Padding{Top: 24, Right: 84, Bottom: 24, Left: 24}) || cfg.Indicators[0].Min != 1000 || cfg.Indicators[0].Label.FontSize != 10 || cfg.Series[0].Options.LabelFontSize != 9 || cfg.Series[1].Options.ValueLabels != radar.ValueLabelsHidden || cfg.Series[1].Options.ValueFormat != radar.ValueFormatInteger {
		t.Fatalf("Radar typed surface = %#v / %#v / %#v", cfg.Padding, cfg.Indicators[0], cfg.Series[1])
	}
}

func TestStaticRadarPageDocumentsDecisionsAccessibilityAndCanonicalAPI(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read pages.templ: %v", err)
	}
	page := string(source)
	start := strings.Index(page, "templ radarContent")
	if start < 0 {
		t.Fatal("cannot isolate static Radar page")
	}
	end := strings.Index(page[start:], "templ candlestickContent")
	if end < 0 {
		t.Fatal("cannot isolate static Radar page")
	}
	radarPage := page[start : start+end]
	for _, want := range []string{
		"AbovePreview: visualizationGuide(", "small number of profiles", "shared bounded dimensions", "aligned bars or a table", "shape and color alone",
		"Readable values and compact layout", "minimums", "indicator labels", "Static/vector behavior", "inline SVG", "no-JavaScript",
		"chart controls", "chart modes", `chartDocumentation(`, `"radar"`,
	} {
		if !strings.Contains(radarPage, want) {
			t.Errorf("static Radar docs missing %q", want)
		}
	}
	for _, forbidden := range []string{"componentContract(", "Component contract", "Primitive", "Kind", "Configuration", "go-analyze", "go-echarts", "Collapse", "Hierarchy contract"} {
		if strings.Contains(radarPage, forbidden) {
			t.Errorf("static Radar docs retain redundant or renderer-specific copy %q", forbidden)
		}
	}
}

func TestStaticRadarCoverageUsesCanonicalLedgerWithoutLostEntries(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/upstream-example-coverage.md")
	if err != nil {
		t.Fatalf("read canonical upstream coverage ledger: %v", err)
	}
	ledger := string(data)
	section := canonicalLedgerSection(t, ledger, "## Static/vector Radar")
	for _, entry := range staticRadarUpstreamCoverage() {
		if count := strings.Count(section, "`"+entry.Path+"`"); count != 1 {
			t.Errorf("static Radar coverage row %q occurs %d times, want 1", entry.Path, count)
		}
		if !strings.Contains(section, "`"+entry.SHA256+"`") {
			t.Errorf("static Radar row %q missing SHA-256 %s", entry.Path, entry.SHA256)
		}
	}
	for _, source := range staticRadarAPISources() {
		for _, want := range []string{"`" + source.Path + "`", "`" + source.SHA256 + "`"} {
			if count := strings.Count(section, want); count != 1 {
				t.Errorf("static Radar API evidence %q occurs %d times, want 1", want, count)
			}
		}
	}
	for _, want := range []string{"both dedicated Radar-family files", "one distinct visual", "Unsupported dedicated Radar-family behaviors: none", "typed renderer-neutral configuration", "arbitrary formatter callbacks"} {
		if !strings.Contains(section, want) {
			t.Errorf("static Radar ledger missing boundary statement %q", want)
		}
	}
}
