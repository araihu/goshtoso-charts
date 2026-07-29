package pages

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/candlestick"
)

func TestStaticCandlestickDedicatedSourceInventoryIsExhaustive(t *testing.T) {
	t.Parallel()
	if staticCandlestickUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" {
		t.Fatalf("revision = %q", staticCandlestickUpstreamRevision)
	}
	want := map[string]string{
		"examples/1-Painter/candlestick_chart-1-basic/main.go":           "44c216955ae850b824a0e3f3ee2bbaf67a23ca185d8faea77335d048cd19c26b",
		"examples/1-Painter/candlestick_chart-2-multiple_series/main.go": "f132f40ac3e920a891782c5ab6f80e681f8e7ee87a0a05405bad75ee161e964f",
		"examples/1-Painter/candlestick_chart-3-bollinger_bands/main.go": "cc3b347d5faea1a15ca22554dcc46a35beed74e49da56701659a1a7d1f000202",
		"examples/1-Painter/candlestick_chart-4-patterns/main.go":        "ab5891e744bc8ec40fbead6b16af5642ea94c738369469b392ac7acf1e0055ec",
		"examples/1-Painter/candlestick_chart-5-aggregation/main.go":     "ba7d1d31fef54f792e53840d969c4a3d791309a6059b2c5997dd2e509e1cbde1",
		"examples/2-OptionFunc/candlestick_chart-1-basic/main.go":        "aad7ab0297061baac358b63b19b15e6dca48734d2be607d11679bd284263423c",
	}
	got := map[string]string{
		staticCandlestickBasicUpstreamPath:       staticCandlestickBasicUpstreamSHA256,
		staticCandlestickMultipleUpstreamPath:    staticCandlestickMultipleUpstreamSHA256,
		staticCandlestickBollingerUpstreamPath:   staticCandlestickBollingerUpstreamSHA256,
		staticCandlestickPatternsUpstreamPath:    staticCandlestickPatternsUpstreamSHA256,
		staticCandlestickAggregationUpstreamPath: staticCandlestickAggregationUpstreamSHA256,
		staticCandlestickOptionFuncUpstreamPath:  staticCandlestickOptionFuncUpstreamSHA256,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dedicated source ledger = %#v, want %#v", got, want)
	}
}

func TestStaticCandlestickAttributionNamesEveryDedicatedSource(t *testing.T) {
	t.Parallel()
	if len(chartAttributions) == 0 {
		t.Fatal("chart attributions are empty")
	}
	attribution := chartAttributions[0]
	if attribution.Name != "go-analyze/charts" || !strings.Contains(attribution.Version, staticCandlestickUpstreamRevision) {
		t.Fatalf("static attribution source = %#v", attribution)
	}
	for _, source := range []string{
		staticCandlestickBasicUpstreamPath,
		staticCandlestickMultipleUpstreamPath,
		staticCandlestickBollingerUpstreamPath,
		staticCandlestickPatternsUpstreamPath,
		staticCandlestickAggregationUpstreamPath,
		staticCandlestickOptionFuncUpstreamPath,
	} {
		if !strings.Contains(attribution.UsedFor, source) {
			t.Errorf("static Candlestick attribution missing %q", source)
		}
	}
}

func TestStaticCandlestickBasicDatasetsMatchBothPinnedSources(t *testing.T) {
	t.Parallel()
	basic := sampleBasicCandlestick()
	extended := sampleCandlestickOptionFuncBasic()
	want := []candlestick.Datum{
		{Label: "Day 1", Open: 100, High: 110, Low: 95, Close: 105},
		{Label: "Day 2", Open: 105, High: 115, Low: 100, Close: 112},
		{Label: "Day 3", Open: 112, High: 118, Low: 108, Close: 115},
		{Label: "Day 4", Open: 115, High: 120, Low: 104, Close: 108},
		{Label: "Day 5", Open: 108, High: 113, Low: 105, Close: 109},
		{Label: "Day 6", Open: 109, High: 116, Low: 106, Close: 114},
		{Label: "Day 7", Open: 114, High: 121, Low: 111, Close: 119},
	}
	if !reflect.DeepEqual(basic.Data, want) {
		t.Fatalf("painter basic data = %#v, want %#v", basic.Data, want)
	}
	wantExtended := append(append([]candlestick.Datum(nil), want...), candlestick.Datum{Label: "Day 8", Open: 119, High: 125, Low: 116, Close: 122})
	if !reflect.DeepEqual(extended.Data, wantExtended) {
		t.Fatalf("option-function basic data = %#v, want %#v", extended.Data, wantExtended)
	}
	if extended.Title != "Basic Candlestick Chart" || extended.Width != 800 || extended.Height != 600 || extended.Options.TitleFontSize != 18 || extended.Options.YUnit != 1 || extended.Options.Padding != (candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}) {
		t.Fatalf("extended basic presentation drifted: %#v", extended)
	}
	for index, candidate := range []candlestick.Config{basic, extended} {
		hash := sha256.New()
		for _, datum := range candidate.Data {
			fmt.Fprintf(hash, "%s|%g|%g|%g|%g\n", datum.Label, datum.Open, datum.High, datum.Low, datum.Close)
		}
		wantHash := []string{"217fe6f6861eedfb6afac7868513ae70c1a3e16d5071f2f6ffab2a7a1d5feb60", "a8bdcf0a7766dd3748a7c97359056f04f8fbf671011c08e54a19c7a39c5d5487"}[index]
		if got := fmt.Sprintf("%x", hash.Sum(nil)); got != wantHash {
			t.Fatalf("normalized source hash = %q, want %q", got, wantHash)
		}
	}
}

func TestStaticCandlestickMultipleSeriesPreservesPinnedDataStylesAndDefectCorrection(t *testing.T) {
	t.Parallel()
	cfg := sampleCandlestickMultipleSeries()
	if cfg.Title != "" || cfg.Width != 1000 || cfg.Height != 700 || len(cfg.Series) != 3 || cfg.Options.YUnit != 1 || cfg.Options.Padding != (candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}) {
		t.Fatalf("multiple-series presentation drifted: %#v", cfg)
	}
	wantNames := []string{"Stock A", "Stock B", "Stock C"}
	wantStyles := []candlestick.BodyStyle{candlestick.BodyStyleFilled, candlestick.BodyStyleTraditional, candlestick.BodyStyleOutline}
	for index, series := range cfg.Series {
		if series.Name != wantNames[index] || series.BodyStyle != wantStyles[index] || len(series.Data) != 5 {
			t.Fatalf("series %d = %#v", index, series)
		}
		base := float64(100 + index*50)
		want := []candlestick.Datum{
			{Label: "Day 1", Open: base, High: base + 10, Low: base - 5, Close: base + 5},
			{Label: "Day 2", Open: base + 5, High: base + 15, Low: base, Close: base + 12},
			{Label: "Day 3", Open: base + 12, High: base + 18, Low: base + 8, Close: base + 15},
			{Label: "Day 4", Open: base + 15, High: base + 20, Low: base + 4, Close: base + 8},
			{Label: "Day 5", Open: base + 8, High: base + 13, Low: base + 5, Close: base + 9},
		}
		if !reflect.DeepEqual(series.Data, want) {
			t.Fatalf("series %d data = %#v, want %#v", index, series.Data, want)
		}
	}
	if !strings.Contains(cfg.Caption, "invalid upstream Day 4 lows") {
		t.Fatalf("defect correction is not disclosed: %q", cfg.Caption)
	}
}

func TestStaticCandlestickLedgerPinsEveryFileFunctionAndFiniteAPISource(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "upstream-example-coverage.md"))
	if err != nil {
		t.Fatal(err)
	}
	ledger := canonicalLedgerSection(t, string(data), "## Static/vector Candlestick")
	for _, want := range []string{
		"## Static/vector Candlestick", "all six dedicated Candlestick-family files", "eight source functions",
		staticCandlestickBasicUpstreamPath, staticCandlestickBasicUpstreamSHA256,
		staticCandlestickMultipleUpstreamPath, staticCandlestickMultipleUpstreamSHA256,
		staticCandlestickBollingerUpstreamPath, staticCandlestickBollingerUpstreamSHA256,
		staticCandlestickPatternsUpstreamPath, staticCandlestickPatternsUpstreamSHA256,
		staticCandlestickAggregationUpstreamPath, staticCandlestickAggregationUpstreamSHA256,
		staticCandlestickOptionFuncUpstreamPath, staticCandlestickOptionFuncUpstreamSHA256,
		"4f576ae2cde41f4d76dcd23babee4f7a14440a6c8686c892552d17375a44868b",
		"7de600489e87d2dcc88a0e56c4aa2417e97bb512d0e000bd615b407063b80c35",
		"2b56840d7ef57840f42670f2b4619cc899503822a3a2bbef51fc13aff38928fe",
		"2df1c6d6e980195d50db54e3d1410f977b21b60cb9b8316bd17d3222bb9e53d9",
		"191c4a25a756552c34775b854b902820934dd0c8f6e227e916f489041b168f4d",
		"7209718f4bfd6a1f7650f049251534fcd17c296dab039b82e88f21406d9fd9f8",
		"d79715bb8b1da24b8b0caa9cd9e3a69f089b30cb47d6a15e16462e4cab0d6ed0",
		"2fb069e158b4c28cd1a4ace193380557aadaab443338ca7f21be76f68aaab128",
		"candlestick_chart.go", "d70ee4b46e6d95de928e83d48ffcc43d3ea5516bbfc431131d1fd5bc61ab667d",
		"candlestick_patterns.go", "c9579221d8b97477d878baa898d22ae7c60d94e4458218565c762dc5a75d8092",
		"series.go", "953f4e5d555701348ebcb8eb0bfe1753a6df56eb4f94a86403c6dc6cecf79217",
		"painter.go", "f4ac102e9b21623765e2fdfe4c0910a03265bc751b9f5d019ae41e80611be959",
		"Unsupported dedicated Candlestick-family behaviors: none.",
	} {
		if !strings.Contains(ledger, want) {
			t.Errorf("source ledger missing %q", want)
		}
	}
	for _, hash := range []string{
		staticCandlestickBasicUpstreamSHA256,
		staticCandlestickMultipleUpstreamSHA256,
		staticCandlestickBollingerUpstreamSHA256,
		staticCandlestickPatternsUpstreamSHA256,
		staticCandlestickAggregationUpstreamSHA256,
		staticCandlestickOptionFuncUpstreamSHA256,
		"4f576ae2cde41f4d76dcd23babee4f7a14440a6c8686c892552d17375a44868b",
		"7de600489e87d2dcc88a0e56c4aa2417e97bb512d0e000bd615b407063b80c35",
		"2b56840d7ef57840f42670f2b4619cc899503822a3a2bbef51fc13aff38928fe",
		"2df1c6d6e980195d50db54e3d1410f977b21b60cb9b8316bd17d3222bb9e53d9",
		"191c4a25a756552c34775b854b902820934dd0c8f6e227e916f489041b168f4d",
		"7209718f4bfd6a1f7650f049251534fcd17c296dab039b82e88f21406d9fd9f8",
		"d79715bb8b1da24b8b0caa9cd9e3a69f089b30cb47d6a15e16462e4cab0d6ed0",
		"2fb069e158b4c28cd1a4ace193380557aadaab443338ca7f21be76f68aaab128",
		"d70ee4b46e6d95de928e83d48ffcc43d3ea5516bbfc431131d1fd5bc61ab667d",
		"c9579221d8b97477d878baa898d22ae7c60d94e4458218565c762dc5a75d8092",
		"953f4e5d555701348ebcb8eb0bfe1753a6df56eb4f94a86403c6dc6cecf79217",
		"f4ac102e9b21623765e2fdfe4c0910a03265bc751b9f5d019ae41e80611be959",
	} {
		if count := strings.Count(ledger, hash); count != 1 {
			t.Errorf("static Candlestick hash %q occurs %d times in scoped ledger, want 1", hash, count)
		}
	}
}
