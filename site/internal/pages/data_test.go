package pages

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/araihu/goshtoso-charts/components/bar"
	"github.com/araihu/goshtoso-charts/components/candlestick"
	"github.com/araihu/goshtoso-charts/components/interactive"
	"github.com/araihu/goshtoso-charts/components/pie"
)

func TestDualAxisLineSampleMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if dualAxisLineUpstreamPath != "examples/1-Painter/line_chart-8-dual_y_axis/main.go" ||
		dualAxisLineUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		dualAxisLineUpstreamSHA256 != "78a3edd9aa356dc798c367b40cc5abecdb765b634795c38767f34bf266b805af" {
		t.Fatalf("dual-axis upstream source = %s@%s SHA-256 %s", dualAxisLineUpstreamPath, dualAxisLineUpstreamRevision, dualAxisLineUpstreamSHA256)
	}
	cfg := sampleDualAxisLine()
	if cfg.Label != "Dual Axis Line" || cfg.Title.Text != "Dual Axis Line" || cfg.Width != 600 || cfg.Height != 400 {
		t.Fatalf("dual-axis title/geometry = %q %#v %dx%d", cfg.Label, cfg.Title, cfg.Width, cfg.Height)
	}
	if !reflect.DeepEqual(cfg.Labels, []string{"A", "B", "C", "D", "E", "F", "G"}) {
		t.Fatalf("dual-axis labels = %#v", cfg.Labels)
	}
	if len(cfg.Series) != 2 || len(cfg.YAxes) != 2 ||
		cfg.Series[0].Name != "Left Series" || cfg.Series[1].Name != "Right Series" ||
		cfg.Series[0].YAxisIndex != 0 || cfg.Series[1].YAxisIndex != 1 {
		t.Fatalf("dual-axis series/axes = %#v / %#v", cfg.Series, cfg.YAxes)
	}
	if !reflect.DeepEqual(cfg.Series[0].Values, []float64{120, 132, 101, 134, 90, 230, 210}) ||
		!reflect.DeepEqual(cfg.Series[1].Values, []float64{820, 932, 901, 934, 1290, 1330, 1320}) {
		t.Fatalf("dual-axis values = %#v", cfg.Series)
	}
}

func TestAreaLineSampleMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if areaLineUpstreamPath != "examples/1-Painter/line_chart-5-area/main.go" ||
		areaLineUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		areaLineUpstreamSHA256 != "b2d7b87ff675f437dbc95f2d7a0447c2040e18c5b873256a5808987dfc6131d0" {
		t.Fatalf("area upstream source = %s@%s SHA-256 %s", areaLineUpstreamPath, areaLineUpstreamRevision, areaLineUpstreamSHA256)
	}
	cfg := sampleAreaLine()
	if cfg.Label != "Line" || cfg.Title.Text != "Line" || cfg.Width != 600 || cfg.Height != 400 {
		t.Fatalf("area title/geometry = %q %#v %dx%d", cfg.Label, cfg.Title, cfg.Width, cfg.Height)
	}
	if !reflect.DeepEqual(cfg.Labels, []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}) || len(cfg.Series) != 1 ||
		cfg.Series[0].Name != "Email" || !reflect.DeepEqual(cfg.Series[0].Values, []float64{120, 132, 101, 134, 90, 230, 210}) {
		t.Fatalf("area labels/series = %#v / %#v", cfg.Labels, cfg.Series)
	}
	if !cfg.Area.Enabled || cfg.Area.Opacity != 150.0/255.0 || cfg.XAxis.BoundaryGap == nil || *cfg.XAxis.BoundaryGap ||
		cfg.Legend.Padding.Top != 5 || cfg.Legend.Padding.Bottom != 10 || len(cfg.YAxes) != 1 || cfg.YAxes[0].Min == nil || *cfg.YAxes[0].Min != 0 {
		t.Fatalf("area options = %#v %#v %#v %#v", cfg.Area, cfg.XAxis, cfg.Legend, cfg.YAxes)
	}
}

func TestInteractiveComponentPagesRemainRendererNeutral(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"pages.templ", "interactive_components.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		lower := strings.ToLower(string(source))
		for _, forbidden := range []string{"go-echarts", "apache echarts", "window.echarts"} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("%s exposes private renderer name %q", path, forbidden)
			}
		}
	}
}

func TestInteractivePageSourceUsesVisualizationGuidance(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read pages.templ: %v", err)
	}
	interactive := string(source)
	start := strings.Index(interactive, "templ InteractiveBarPage")
	end := strings.Index(interactive, "templ pieContent")
	if start < 0 || end < 0 || start >= end {
		t.Fatal("cannot isolate interactive component pages")
	}
	interactive = interactive[start:end]
	if got := strings.Count(interactive, "AbovePreview: visualizationGuidance("); got != 24 {
		t.Fatalf("interactive guidance count = %d, want 24 canonical component routes", got)
	}
	for _, forbidden := range []string{"AbovePreview: componentContract(", `"Primitive"`, `"Kind"`, `"Configuration"`, `"Component contract"`} {
		if strings.Contains(interactive, forbidden) {
			t.Errorf("interactive component pages retain %q", forbidden)
		}
	}
}

func TestInteractiveSunburstSourceExplainsAccessibleHierarchyReading(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read pages.templ: %v", err)
	}
	page := string(source)
	start := strings.Index(page, "templ interactiveSunburstContent")
	if start < 0 {
		t.Fatal("cannot isolate sunburst page")
	}
	end := strings.Index(page[start:], "templ interactiveTreeContent")
	if end < 0 {
		t.Fatal("cannot isolate sunburst page")
	}
	sunburst := page[start : start+end]
	for _, want := range []string{"shallow hierarchy", "Deep hierarchies", "keyboard navigation", "Hierarchy contract"} {
		if want == "Hierarchy contract" {
			if strings.Contains(sunburst, want) {
				t.Errorf("sunburst retains %q", want)
			}
			continue
		}
		if !strings.Contains(sunburst, want) {
			t.Errorf("sunburst guidance missing %q", want)
		}
	}
	if !strings.Contains(sunburst, "not the accessible navigation path") {
		t.Error("sunburst must not present visual drill-down as accessible navigation")
	}
}

func TestGoAPIReferenceUsesGoshtosoButtonAppearanceLink(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read pages.templ: %v", err)
	}
	page := string(source)
	start := strings.Index(page, "templ goAPIReference")
	if start < 0 {
		t.Fatal("cannot find Go API footer")
	}
	footer := page[start:]
	if !strings.Contains(footer, "link.Link(") || !strings.Contains(footer, "link.WithAppearance(link.AppearanceButton)") {
		t.Error("Go API footer must use the Goshtoso button-appearance Link")
	}
	if strings.Contains(footer, "<a href={ templ.URL(\"https://pkg.go.dev") {
		t.Error("Go API footer must not duplicate raw button-link markup")
	}
}

func TestDoughnutSampleMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if doughnutUpstreamPath != "examples/1-Painter/doughnut_chart-1-basic/main.go" ||
		doughnutUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		doughnutUpstreamSHA256 != "b97bca2322e90e2f03ab49aa77f683d0c58e027846b939e5a61100602dad1ebf" {
		t.Fatalf("doughnut upstream source = %s@%s SHA-256 %s", doughnutUpstreamPath, doughnutUpstreamRevision, doughnutUpstreamSHA256)
	}
	cfg := sampleDoughnutChart()
	if cfg.Label != "Doughnut Chart" || cfg.Title.Text != "Doughnut Chart" || cfg.Title.Subtitle != "(Fake Data)" {
		t.Fatalf("doughnut titles = label %q title %#v", cfg.Label, cfg.Title)
	}
	if cfg.Width != 600 || cfg.Height != 400 || cfg.InnerRadiusPercent != 60 {
		t.Fatalf("doughnut geometry = %dx%d inner %v", cfg.Width, cfg.Height, cfg.InnerRadiusPercent)
	}
	if cfg.Padding.Top != 20 || cfg.Padding.Right != 20 || cfg.Padding.Bottom != 20 || cfg.Padding.Left != 20 {
		t.Fatalf("doughnut padding = %#v", cfg.Padding)
	}
	if cfg.Legend.LeftPercent != 80 || cfg.Legend.VerticalPlacement != pie.VerticalPlacementBottom ||
		cfg.Legend.Orientation != pie.LegendVertical || cfg.Legend.FontSize != 10 {
		t.Fatalf("doughnut legend = %#v", cfg.Legend)
	}
	names := make([]string, len(cfg.Slices))
	values := make([]float64, len(cfg.Slices))
	for index, slice := range cfg.Slices {
		names[index], values[index] = slice.Name, slice.Value
	}
	if want := []string{"Search Engine", "Direct", "Email", "Union Ads", "Video Ads"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("doughnut names = %v, want %v", names, want)
	}
	if want := []float64{1048, 735, 580, 484, 300}; !reflect.DeepEqual(values, want) {
		t.Fatalf("doughnut values = %v, want %v", values, want)
	}
}

func TestHorizontalBarMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if horizontalBarUpstreamPath != "examples/1-Painter/horizontal_bar_chart-1-basic/main.go" ||
		horizontalBarUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		horizontalBarUpstreamSHA256 != "735240dd8433bd2494ae019f272840a8ff2fcf5572166b78269e23cbff7111a0" {
		t.Fatalf("horizontal bar upstream source = %s@%s SHA-256 %s", horizontalBarUpstreamPath, horizontalBarUpstreamRevision, horizontalBarUpstreamSHA256)
	}
	cfg := sampleHorizontalWorldPopulation()
	if cfg.Title != "World Population" || cfg.Width != 600 || cfg.Height != 400 {
		t.Fatalf("title/dimensions drifted: %q %dx%d", cfg.Title, cfg.Width, cfg.Height)
	}
	if cfg.Padding.Top != 20 || cfg.Padding.Right != 40 || cfg.Padding.Bottom != 20 || cfg.Padding.Left != 20 {
		t.Fatalf("padding drifted: %#v", cfg.Padding)
	}
	wantLabels := []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"}
	wantNames := []string{"2011", "2012"}
	wantValues := [][]float64{{10, 30, 50, 70, 90, 110, 130}, {20, 40, 60, 80, 100, 120, 140}}
	if !reflect.DeepEqual(cfg.Labels, wantLabels) || len(cfg.Series) != 2 {
		t.Fatalf("category/series shape drifted: labels %v series %d", cfg.Labels, len(cfg.Series))
	}
	for index, series := range cfg.Series {
		if series.Name != wantNames[index] || !reflect.DeepEqual(series.Values, wantValues[index]) {
			t.Fatalf("series %d = %q %v, want %q %v", index, series.Name, series.Values, wantNames[index], wantValues[index])
		}
	}
}

func TestBarReferencesMechanicallyMatchPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if barReferencesUpstreamPath != "examples/1-Painter/bar_chart-4-mark/main.go" || barReferencesUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" || barReferencesUpstreamSHA256 != "544fea22c29db4225c7b10bb6d12137d484a4ca9b6c647dc29730a61ce4ced4c" {
		t.Fatalf("bar reference upstream source = %s@%s SHA-256 %s", barReferencesUpstreamPath, barReferencesUpstreamRevision, barReferencesUpstreamSHA256)
	}
	cfg := sampleBarReferences()
	if cfg.Width != 600 || cfg.Height != 400 || !reflect.DeepEqual(cfg.Labels, []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}) {
		t.Fatalf("reference bar geometry/categories drifted: %dx%d %v", cfg.Width, cfg.Height, cfg.Labels)
	}
	want := [][]float64{{2.0, 4.9, 7.0, 23.2, 25.6, 76.7, 135.6, 162.2, 32.6, 20.0, 6.4, 3.3}, {2.6, 5.9, 9.0, 26.4, 28.7, 70.7, 175.6, 182.2, 48.7, 18.8, 6.0, 2.3}}
	for index, series := range cfg.Series {
		if !reflect.DeepEqual(series.Values, want[index]) || !series.References.Average || !series.References.Minimum || !series.References.Maximum || series.References.Format != bar.ValueFormatHumanized {
			t.Fatalf("series %d drifted: %#v", index, series)
		}
	}
}

func TestViolinSamplesAreDeterministicAndPreserveUpstreamGenerator(t *testing.T) {
	t.Parallel()
	first, second := sampleDistributionShapes(), sampleDistributionShapes()
	if !reflect.DeepEqual(first.Series, second.Series) {
		t.Fatal("fixed LCG seed did not reproduce violin samples")
	}
	if first.Title != "Distribution Shapes" || first.Width != 1200 || first.Height != 800 || first.Density.Points != 80 {
		t.Fatalf("sample config = title %q, %dx%d, %d points", first.Title, first.Width, first.Height, first.Density.Points)
	}
	wantNames := []string{"Normal", "Right Skewed", "Bimodal", "Tight"}
	wantFirst := []float64{51.632672269835695, 34.032704172611375, 81.03976781672685, 44.259116252823056}
	if len(first.Series) != len(wantNames) {
		t.Fatalf("series count = %d", len(first.Series))
	}
	for index, series := range first.Series {
		if series.Name != wantNames[index] || len(series.Samples) != 200 {
			t.Fatalf("series %d = %q with %d samples", index, series.Name, len(series.Samples))
		}
		if math.Abs(series.Samples[0]-wantFirst[index]) > 1e-12 {
			t.Errorf("series %q first sample = %.15f, want %.15f", series.Name, series.Samples[0], wantFirst[index])
		}
		if !series.Marks.Mean || !series.Marks.Median || !reflect.DeepEqual(series.Statistics.Quantiles, []float64{.25, .75}) {
			t.Errorf("series %q statistics = marks %#v, summary %#v", series.Name, series.Marks, series.Statistics)
		}
	}
}

func TestBasicFunnelDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	cfg := sampleBasicFunnel()
	wantLabels := []string{"Show", "Click", "Visit", "Inquiry", "Order", "Pay", "Cancel"}
	wantValues := []float64{100, 80, 60, 40, 20, 10, 2}
	if cfg.Title != "Funnel" || cfg.Width != 0 || cfg.Height != 0 || cfg.Options.Legend.Padding.Left != 100 {
		t.Fatalf("title/dimensions/legend geometry drifted: %#v", cfg)
	}
	if len(cfg.Stages) != len(wantLabels) {
		t.Fatalf("stage count = %d, want %d", len(cfg.Stages), len(wantLabels))
	}
	for index, stage := range cfg.Stages {
		if stage.Label != wantLabels[index] || stage.Value != wantValues[index] {
			t.Fatalf("stage %d = (%q, %g), want (%q, %g)", index, stage.Label, stage.Value, wantLabels[index], wantValues[index])
		}
	}
}

func TestStaticCandlestickBollingerDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if staticCandlestickBollingerUpstreamPath != "examples/1-Painter/candlestick_chart-3-bollinger_bands/main.go" ||
		staticCandlestickBollingerUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		staticCandlestickBollingerUpstreamSHA256 != "cc3b347d5faea1a15ca22554dcc46a35beed74e49da56701659a1a7d1f000202" {
		t.Fatalf("Bollinger upstream source = %s@%s (%s)", staticCandlestickBollingerUpstreamPath, staticCandlestickBollingerUpstreamRevision, staticCandlestickBollingerUpstreamSHA256)
	}
	cfg := sampleCandlestickBollingerBands()
	if cfg.Title != "Candlestick Chart with Bollinger Bands" || cfg.Width != 800 || cfg.Height != 600 ||
		cfg.Options.TitleFontSize != 18 || cfg.Options.YUnit != 1 ||
		cfg.Options.Padding != (candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}) {
		t.Fatalf("Bollinger presentation drifted: %#v", cfg)
	}
	if len(cfg.Data) != 20 || len(cfg.TrendLines) != 3 {
		t.Fatalf("Bollinger shape = %d data / %d trends", len(cfg.Data), len(cfg.TrendLines))
	}
	hash := sha256.New()
	for _, datum := range cfg.Data {
		fmt.Fprintf(hash, "%s|%g|%g|%g|%g\n", datum.Label, datum.Open, datum.High, datum.Low, datum.Close)
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "fc218c7fedf84c7ac739016015e2508439a9a31d2c68a46fa1bfa84dc0e8f1ef" {
		t.Fatalf("normalized pinned Bollinger OHLC SHA-256 = %s", got)
	}
	wantTypes := []candlestick.TrendType{
		candlestick.TrendTypeBollingerUpper,
		candlestick.TrendTypeSimpleMovingAverage,
		candlestick.TrendTypeBollingerLower,
	}
	for index, trend := range cfg.TrendLines {
		if trend.Type != wantTypes[index] || trend.Period != 5 {
			t.Fatalf("Bollinger trend %d = %#v", index, trend)
		}
	}
}

func TestStaticCandlestickPatternsDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if staticCandlestickPatternsUpstreamPath != "examples/1-Painter/candlestick_chart-4-patterns/main.go" ||
		staticCandlestickPatternsUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		staticCandlestickPatternsUpstreamSHA256 != "ab5891e744bc8ec40fbead6b16af5642ea94c738369469b392ac7acf1e0055ec" {
		t.Fatalf("patterns upstream source = %s@%s (%s)", staticCandlestickPatternsUpstreamPath, staticCandlestickPatternsUpstreamRevision, staticCandlestickPatternsUpstreamSHA256)
	}
	cfg := sampleCandlestickPatterns()
	if cfg.Title != "Candlestick Patterns" || cfg.SeriesName != "Stock Price with Patterns" || cfg.Width != 900 || cfg.Height != 650 ||
		cfg.Patterns.Selection != candlestick.PatternSelectionAll || len(cfg.Patterns.References) != 2 ||
		cfg.Patterns.References[0] != candlestick.CloseReferenceAverage || cfg.Patterns.References[1] != candlestick.CloseReferenceMinimum {
		t.Fatalf("patterns configuration drifted: %#v", cfg)
	}
	want := [][4]float64{{100, 110, 95, 105}, {105, 108, 102, 105.1}, {108, 109, 98, 107}, {107, 108, 103, 104}, {102, 115, 101, 113}, {113, 125, 112, 114}, {114, 118, 113, 117}, {119, 120, 108, 110}, {110, 113, 107, 109.9}, {109, 118, 108, 116}}
	if len(cfg.Data) != len(want) {
		t.Fatalf("patterns datum count = %d, want %d", len(cfg.Data), len(want))
	}
	for index, datum := range cfg.Data {
		if got := [4]float64{datum.Open, datum.High, datum.Low, datum.Close}; got != want[index] {
			t.Fatalf("patterns datum %d = %v, want %v", index+1, got, want[index])
		}
	}
	if sampleCandlestickCorePatterns().Patterns.Selection != candlestick.PatternSelectionCore || sampleCandlestickBullishPatterns().Patterns.Selection != candlestick.PatternSelectionBullish || !sampleCandlestickPatternLabels().Patterns.PreferLabels {
		t.Fatal("pattern variants drifted")
	}
}

func TestStaticCandlestickAggregationDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if staticCandlestickAggregationUpstreamPath != "examples/1-Painter/candlestick_chart-5-aggregation/main.go" ||
		staticCandlestickAggregationUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		staticCandlestickAggregationUpstreamSHA256 != "ba7d1d31fef54f792e53840d969c4a3d791309a6059b2c5997dd2e509e1cbde1" {
		t.Fatalf("aggregation upstream source = %s@%s (%s)", staticCandlestickAggregationUpstreamPath, staticCandlestickAggregationUpstreamRevision, staticCandlestickAggregationUpstreamSHA256)
	}
	cfg := sampleCandlestickAggregation()
	if cfg.Title != "1-Minute Candles (Before Aggregation)" || cfg.SeriesName != "1-Minute" || cfg.Width != 1200 || cfg.Height != 800 ||
		cfg.Aggregation.WindowSize != 5 || cfg.Aggregation.Title != "5-Minute Aggregated Candles" || cfg.Aggregation.SeriesName != "5-Minute" ||
		cfg.Options.TitleFontSize != 16 || cfg.Options.YUnit != 1 || cfg.Options.Padding != (candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}) {
		t.Fatalf("aggregation presentation drifted: %#v", cfg)
	}
	want := []candlestick.Datum{
		{Label: "1", Open: 100, High: 102, Low: 99, Close: 101}, {Label: "2", Open: 101, High: 103, Low: 100, Close: 102},
		{Label: "3", Open: 102, High: 105, Low: 101, Close: 104}, {Label: "4", Open: 104, High: 106, Low: 103, Close: 105},
		{Label: "5", Open: 105, High: 107, Low: 104, Close: 106}, {Label: "6", Open: 106, High: 108, Low: 105, Close: 107},
		{Label: "7", Open: 107, High: 109, Low: 106, Close: 108}, {Label: "8", Open: 108, High: 110, Low: 107, Close: 109},
		{Label: "9", Open: 109, High: 111, Low: 108, Close: 110}, {Label: "10", Open: 110, High: 112, Low: 109, Close: 111},
		{Label: "11", Open: 111, High: 113, Low: 110, Close: 112}, {Label: "12", Open: 112, High: 114, Low: 111, Close: 113},
		{Label: "13", Open: 113, High: 115, Low: 112, Close: 114}, {Label: "14", Open: 114, High: 116, Low: 113, Close: 115},
		{Label: "15", Open: 115, High: 117, Low: 114, Close: 116},
	}
	if !reflect.DeepEqual(cfg.Data, want) {
		t.Fatalf("aggregation source data = %#v, want %#v", cfg.Data, want)
	}
}

func TestInteractiveCandlestickDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if interactiveCandlestickUpstreamPath != "examples/kline.go" {
		t.Fatalf("upstream path = %q", interactiveCandlestickUpstreamPath)
	}
	if interactiveCandlestickUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" {
		t.Fatalf("upstream revision = %q", interactiveCandlestickUpstreamRevision)
	}
	if len(interactiveCandlestickUpstreamData) != 88 {
		t.Fatalf("candlestick datum count = %d, want 88", len(interactiveCandlestickUpstreamData))
	}
	hash := sha256.New()
	for _, datum := range interactiveCandlestickUpstreamData {
		fmt.Fprintf(hash, "%s|%g|%g|%g|%g\n", datum.Category, datum.Candle.Open, datum.Candle.Close, datum.Candle.Low, datum.Candle.High)
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "03fd8007530e739fbfce31cbe3e3e2f59e174ddaece06a7a144a30ec225f3c4f" {
		t.Fatalf("normalized pinned upstream OHLC SHA-256 = %s", got)
	}
}

func TestInteractiveMapDatasetsAndVariantsMatchPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if interactiveMapUpstreamPath != "examples/map.go" || interactiveMapUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" {
		t.Fatalf("map upstream source = %s@%s", interactiveMapUpstreamPath, interactiveMapUpstreamRevision)
	}
	if interactiveMapUpstreamSHA256 != "3b59b5cb7ed392f3fa436d51fd420704ab2e82e439c95b226d35d12b913cf9da" {
		t.Fatalf("map upstream SHA-256 = %s", interactiveMapUpstreamSHA256)
	}
	if interactiveMapGeometryRevision != "IBGE-MMD-2025" || interactiveMapGeometrySHA256 != "1b3719c82f6e2278a3e6ea8b7fc2e195460ee6a7de1546d0a8e05e6d0174bb3d" {
		t.Fatalf("map geometry source = %s / %s", interactiveMapGeometryRevision, interactiveMapGeometrySHA256)
	}
	want := []struct{ name, code string }{
		{"Rondônia", "RO"}, {"Acre", "AC"}, {"Amazonas", "AM"}, {"Roraima", "RR"}, {"Pará", "PA"}, {"Amapá", "AP"}, {"Tocantins", "TO"},
		{"Maranhão", "MA"}, {"Piauí", "PI"}, {"Ceará", "CE"}, {"Rio Grande do Norte", "RN"}, {"Paraíba", "PB"}, {"Pernambuco", "PE"}, {"Alagoas", "AL"}, {"Sergipe", "SE"}, {"Bahia", "BA"},
		{"Minas Gerais", "MG"}, {"Espírito Santo", "ES"}, {"Rio de Janeiro", "RJ"}, {"São Paulo", "SP"}, {"Paraná", "PR"}, {"Santa Catarina", "SC"}, {"Rio Grande do Sul", "RS"},
		{"Mato Grosso do Sul", "MS"}, {"Mato Grosso", "MT"}, {"Goiás", "GO"}, {"Distrito Federal", "DF"},
	}
	if len(interactiveMapBrazilRegions) != len(want) {
		t.Fatalf("Brazil region count = %d, want %d", len(interactiveMapBrazilRegions), len(want))
	}
	for index, region := range interactiveMapBrazilRegions {
		if region.Name != want[index].name || region.Code != want[index].code || region.Value < 0 || region.Value >= 150 {
			t.Errorf("Brazil region %d = %#v, want %s/%s and upstream [0,150) value domain", index, region, want[index].name, want[index].code)
		}
	}
	variants := sampleInteractiveMaps()
	wantVariants := []interactive.MapVariant{interactive.MapVariantBasic, interactive.MapVariantLabels, interactive.MapVariantScale, interactive.MapVariantRegional, interactive.MapVariantTheme}
	if len(variants) != len(wantVariants) {
		t.Fatalf("map variant count = %d, want %d", len(variants), len(wantVariants))
	}
	for index := range variants {
		if variants[index].variant != wantVariants[index] {
			t.Errorf("map variant %d = %q, want %q", index, variants[index].variant, wantVariants[index])
		}
	}
}

func TestInteractiveGeoDatasetsAndVariantsMatchPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if interactiveGeoUpstreamPath != "examples/geo.go" || interactiveGeoUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" {
		t.Fatalf("geo upstream source = %s@%s", interactiveGeoUpstreamPath, interactiveGeoUpstreamRevision)
	}
	if interactiveGeoUpstreamSHA256 != "3a6dbe86c34e5ea478b1dea5430c10cac9f7c4905264e12fc37654f0f5d4550a" {
		t.Fatalf("geo upstream SHA-256 = %s", interactiveGeoUpstreamSHA256)
	}
	if interactiveGeoGeometryRevision != "IBGE-MMD-2025" || interactiveGeoBrazilSHA256 != interactiveMapGeometrySHA256 || interactiveGeoSaoPauloSHA256 != "657dee960c4c4d991f5b0e6d59681152d5e2b9c48091e5094085a666c97ff317" {
		t.Fatalf("geo geometry pins = %s / %s / %s", interactiveGeoGeometryRevision, interactiveGeoBrazilSHA256, interactiveGeoSaoPauloSHA256)
	}
	wantNational := []interactive.GeoPoint{
		{Name: "Manaus", Longitude: -60.02, Latitude: -3.12, Value: 81},
		{Name: "Recife", Longitude: -34.88, Latitude: -8.05, Value: 27},
		{Name: "Brasília", Longitude: -47.88, Latitude: -15.79, Value: 47},
		{Name: "Rio de Janeiro", Longitude: -43.17, Latitude: -22.91, Value: 59},
		{Name: "São Paulo", Longitude: -46.63, Latitude: -23.55, Value: 18},
		{Name: "Porto Alegre", Longitude: -51.23, Latitude: -30.03, Value: 63},
	}
	wantSaoPaulo := []interactive.GeoPoint{
		{Name: "São Paulo", Longitude: -46.63, Latitude: -23.55, Value: 12},
		{Name: "Campinas", Longitude: -47.06, Latitude: -22.91, Value: 76},
		{Name: "Ribeirão Preto", Longitude: -47.81, Latitude: -21.18, Value: 41},
	}
	if !reflect.DeepEqual(interactiveGeoBrazilPoints, wantNational) {
		t.Fatalf("Brazil points = %#v", interactiveGeoBrazilPoints)
	}
	if !reflect.DeepEqual(interactiveGeoSaoPauloPoints, wantSaoPaulo) {
		t.Fatalf("São Paulo points = %#v", interactiveGeoSaoPauloPoints)
	}
	for _, points := range [][]interactive.GeoPoint{interactiveGeoBrazilPoints, interactiveGeoSaoPauloPoints} {
		for _, point := range points {
			if point.Value < 0 || point.Value >= 100 {
				t.Errorf("fixed literal value %q=%g outside upstream [0,100) domain", point.Name, point.Value)
			}
		}
	}
	variants := sampleInteractiveGeos()
	if len(variants) != 2 || variants[0].name != "effect-scatter" || variants[1].name != "scatter" {
		t.Fatalf("geo variants = %#v", variants)
	}
}

func TestDenseScatterValuesAreDeterministicAndPreserveUpstreamDistribution(t *testing.T) {
	t.Parallel()
	first := denseScatterValues(rand.New(rand.NewSource(20260728)), 3, 1000, 10)
	second := denseScatterValues(rand.New(rand.NewSource(20260728)), 3, 1000, 10)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("fixed local seed did not reproduce dense data")
	}
	if len(first) != 3 {
		t.Fatalf("series count = %d", len(first))
	}
	for seriesIndex, series := range first {
		if len(series) != 1000 {
			t.Fatalf("series %d category count = %d", seriesIndex, len(series))
		}
		for index, samples := range series {
			want := 1
			if index > 0 && index%2 == 0 {
				want++
			}
			if index > 0 && index%10 == 0 {
				want++
			}
			if len(samples) != want {
				t.Fatalf("series %d category %d samples = %d, want %d", seriesIndex, index, len(samples), want)
			}
			if index > 0 {
				previous := series[index-1][0]
				minimum, maximum := previous*.9, previous*1.1
				for _, sample := range samples {
					if sample < minimum || sample > maximum {
						t.Fatalf("series %d category %d value %f outside 10%% walk [%f,%f]", seriesIndex, index, sample, minimum, maximum)
					}
				}
			}
		}
	}
}

func TestTopNScatterMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if topNScatterUpstreamPath != "examples/1-Painter/scatter_chart-4-top_n_labels/main.go" ||
		topNScatterUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		topNScatterUpstreamSHA256 != "cf92798819fbc010f44eaa406acabd337f16b52eec00793a10679e9c3b7cda81" {
		t.Fatalf("top N scatter upstream source = %s@%s SHA-256 %s", topNScatterUpstreamPath, topNScatterUpstreamRevision, topNScatterUpstreamSHA256)
	}
	cfg := sampleTopNScatter()
	if cfg.Title.Text != "Website Traffic Over 30 Days - Peak Days Highlighted" || cfg.Title.Subtext != "(Only top 5 traffic days show labels)" || !cfg.Legend.Hidden || cfg.Width != 800 || cfg.Height != 500 || cfg.Padding.Top != 20 || cfg.Padding.Right != 20 || cfg.Padding.Bottom != 20 || cfg.Padding.Left != 20 {
		t.Fatalf("top N scatter presentation = %#v", cfg)
	}
	if len(cfg.Categories) != 30 || cfg.Categories[0] != "Day 1" || cfg.Categories[29] != "Day 30" || len(cfg.Series) != 1 || cfg.Series[0].Name != "Daily Visitors (k)" {
		t.Fatalf("top N scatter categories/series = %#v", cfg)
	}
	want := []float64{15.2, 18.5, 22.1, 19.8, 25.4, 21.3, 17.9, 32.6, 28.1, 24.7, 31.5, 29.3, 26.8, 35.2, 41.7, 38.9, 33.1, 29.6, 27.4, 30.8, 36.3, 42.1, 39.5, 44.8, 48.3, 45.6, 40.2, 37.9, 34.5, 26.1}
	for index, value := range want {
		if got := cfg.Series[0].Values[index]; len(got) != 1 || got[0] != value {
			t.Fatalf("top N scatter value %d = %v, want %g", index, got, value)
		}
	}
	if cfg.Options.TopNLabels.Count != 5 || cfg.Options.TopNLabels.FontSize != 16 || cfg.YAxis.Min == nil || *cfg.YAxis.Min != 0 || cfg.YAxis.Max == nil || *cfg.YAxis.Max != 50 {
		t.Fatalf("top N scatter options = %#v %#v", cfg.Options, cfg.YAxis)
	}
}

func TestThemeRiverDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	streams := sampleThemeRiverStreams()
	wantNames := []string{"DQ", "TY", "SS", "QG", "SY", "DD"}
	wantValues := [][]float64{
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
		{35, 36, 37, 22, 24, 26, 34, 21, 18, 45, 32, 35, 30, 28, 27, 26, 15, 30, 35, 42, 42},
		{21, 25, 27, 23, 24, 21, 35, 39, 40, 36, 33, 43, 40, 34, 28, 26, 37, 41, 46, 47, 41},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 4, 32, 26, 22, 16, 22, 10},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 4, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
	}
	if len(streams) != len(wantNames) {
		t.Fatalf("stream count = %d, want %d", len(streams), len(wantNames))
	}
	for streamIndex, stream := range streams {
		if stream.Name != wantNames[streamIndex] {
			t.Fatalf("stream %d name = %q, want %q", streamIndex, stream.Name, wantNames[streamIndex])
		}
		if len(stream.Points) != 21 {
			t.Fatalf("stream %q point count = %d", stream.Name, len(stream.Points))
		}
		for pointIndex, point := range stream.Points {
			wantDate := time.Date(2015, time.November, 8+pointIndex, 0, 0, 0, 0, time.UTC)
			if !point.Time.Equal(wantDate) || point.Value != wantValues[streamIndex][pointIndex] {
				t.Fatalf("stream %q point %d = (%s, %g), want (%s, %g)", stream.Name, pointIndex, point.Time, point.Value, wantDate, wantValues[streamIndex][pointIndex])
			}
		}
	}
}

func TestWordCloudDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	words := sampleWordCloudWords()
	want := []struct {
		name  string
		value float64
	}{
		{"Sam S Club", 10000}, {"Macys", 6181}, {"Amy Schumer", 4386}, {"Jurassic World", 4055},
		{"Charter Communications", 2467}, {"Chick Fil A", 2244}, {"Planet Fitness", 1898},
		{"Pitch Perfect", 1484}, {"Express", 1689}, {"Home", 1112}, {"Johnny Depp", 985},
		{"Lena Dunham", 847}, {"Lewis Hamilton", 582}, {"KXAN", 555}, {"Mary Ellen Mark", 550},
		{"Farrah Abraham", 462}, {"Rita Ora", 366}, {"Serena Williams", 282},
		{"NCAA baseball tournament", 273}, {"Point Break", 265},
	}
	if len(words) != len(want) {
		t.Fatalf("word count = %d, want %d", len(words), len(want))
	}
	for index, word := range words {
		if word.Name != want[index].name || word.Value != want[index].value {
			t.Fatalf("word %d = (%q, %g), want (%q, %g)", index, word.Name, word.Value, want[index].name, want[index].value)
		}
	}
}

func TestLiquidGaugeDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	data := sampleLiquidGaugeData()
	want := []float64{.3, .4, .5}
	if len(data) != len(want) {
		t.Fatalf("liquid reading count = %d, want %d", len(data), len(want))
	}
	for index := range want {
		if data[index].Name != fmt.Sprintf("Wave %d", index+1) || data[index].Value != want[index] {
			t.Fatalf("liquid reading %d = (%q, %g), want (Wave %d, %g)", index, data[index].Name, data[index].Value, index+1, want[index])
		}
	}
}
