package candlestick

import (
	"bytes"
	"context"
	"math"
	"reflect"
	"strings"
	"testing"
)

func multipleSeriesConfig() Config {
	data := func(base float64) []Datum {
		return []Datum{
			{Label: "Day 1", Open: base, High: base + 10, Low: base - 5, Close: base + 5},
			{Label: "Day 2", Open: base + 5, High: base + 15, Low: base, Close: base + 12},
			{Label: "Day 3", Open: base + 12, High: base + 18, Low: base + 8, Close: base + 15},
			// Upstream sets Low above Close. Correct the invalid range while preserving the bearish body.
			{Label: "Day 4", Open: base + 15, High: base + 20, Low: base + 4, Close: base + 8},
			{Label: "Day 5", Open: base + 8, High: base + 13, Low: base + 5, Close: base + 9},
		}
	}
	return Config{
		Label: "Three stock series",
		Series: []Series{
			{Name: "Stock A", Data: data(100), BodyStyle: BodyStyleFilled},
			{Name: "Stock B", Data: data(150), BodyStyle: BodyStyleTraditional},
			{Name: "Stock C", Data: data(200), BodyStyle: BodyStyleOutline},
		},
		Options: Options{YUnit: 1, Padding: Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}},
		Width:   1000, Height: 700,
	}
}

func TestCandlestickMapsPinnedMultipleSeriesAndBodyStyles(t *testing.T) {
	t.Parallel()
	cfg := multipleSeriesConfig()
	options := candlestickOptions(cfg)
	if options.Title.Text != "" || len(options.SeriesList) != 3 || options.XAxis.Labels[4] != "Day 5" {
		t.Fatalf("multiple-series shape = %#v", options)
	}
	if got := []string{options.SeriesList[0].Name, options.SeriesList[1].Name, options.SeriesList[2].Name}; !reflect.DeepEqual(got, []string{"Stock A", "Stock B", "Stock C"}) {
		t.Fatalf("series names = %v", got)
	}
	if got := []string{options.SeriesList[0].CandleStyle, options.SeriesList[1].CandleStyle, options.SeriesList[2].CandleStyle}; !reflect.DeepEqual(got, []string{"filled", "traditional", "outline"}) {
		t.Fatalf("body styles = %v", got)
	}
	if options.SeriesList[2].Data[3].Open != 215 || options.SeriesList[2].Data[3].High != 220 || options.SeriesList[2].Data[3].Low != 204 || options.SeriesList[2].Data[3].Close != 208 {
		t.Fatalf("Stock C day 4 drifted: %#v", options.SeriesList[2].Data[3])
	}
	if options.YAxis[0].Unit != 1 {
		t.Fatalf("Y unit = %v", options.YAxis[0].Unit)
	}
	if options.Padding.Top != 20 || options.Padding.Right != 20 || options.Padding.Bottom != 20 || options.Padding.Left != 20 {
		t.Fatalf("padding = %#v", options.Padding)
	}
}

func TestCandlestickMapsFiniteGeometryAndPerSeriesWickOverride(t *testing.T) {
	t.Parallel()
	show, hide, margin := true, false, .1
	cfg := multipleSeriesConfig()
	cfg.Options.Geometry = Geometry{CandleWidth: .55, WickWidth: 2.5, SeriesGap: &margin, ShowWicks: &hide}
	cfg.Series[1].ShowWicks = &show
	options := candlestickOptions(cfg)
	if options.CandleWidth != .55 || options.WickWidth != 2.5 || options.CandleMargin == nil || *options.CandleMargin != .1 || options.ShowWicks == nil || *options.ShowWicks {
		t.Fatalf("geometry = width %v wick %v margin %v show %v", options.CandleWidth, options.WickWidth, options.CandleMargin, options.ShowWicks)
	}
	if options.SeriesList[0].ShowWicks != nil || options.SeriesList[1].ShowWicks == nil || !*options.SeriesList[1].ShowWicks {
		t.Fatalf("per-series wick overrides = %#v", options.SeriesList)
	}
}

func TestCandlestickFinitePatternVocabularyAndThresholds(t *testing.T) {
	t.Parallel()
	for selection, want := range map[PatternSelection][]PatternType{
		PatternSelectionBearish:  {PatternTypeShootingStar, PatternTypeGravestoneDoji, PatternTypeBearishMarubozu, PatternTypeBearishEngulfing, PatternTypeDarkCloudCover, PatternTypeEveningStar},
		PatternSelectionReversal: {PatternTypeHammer, PatternTypeShootingStar, PatternTypeDragonflyDoji, PatternTypeGravestoneDoji, PatternTypeBullishEngulfing, PatternTypeBearishEngulfing, PatternTypePiercingLine, PatternTypeDarkCloudCover, PatternTypeMorningStar, PatternTypeEveningStar},
		PatternSelectionTrend:    {PatternTypeBullishMarubozu, PatternTypeBearishMarubozu},
	} {
		if got := patternSelections[selection]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s patterns = %v, want %v", selection, got, want)
		}
	}
	options := PatternOptions{
		Enabled:       []PatternType{PatternTypeDoji, PatternTypeHammer},
		DojiThreshold: .08, ShadowTolerance: .02, ShadowRatio: 2.5, EngulfingMinSize: 1.2,
	}
	mapped := chartPatternConfig(options)
	if !reflect.DeepEqual(mapped.EnabledPatterns, []string{"doji", "hammer"}) || mapped.DojiThreshold != .08 || mapped.ShadowTolerance != .02 || mapped.ShadowRatio != 2.5 || mapped.EngulfingMinSize != 1.2 {
		t.Fatalf("pattern mapping = %#v", mapped)
	}
	data := patternConfig().Data
	results, err := DetectPatterns(data, options)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if result.Type != PatternTypeDoji && result.Type != PatternTypeHammer {
			t.Fatalf("custom pattern selection leaked %q", result.Type)
		}
	}
}

func TestCandlestickMultipleSeriesRendersOneExactTablePerSeries(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := Candlestick(multipleSeriesConfig()).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	markup := output.String()
	for _, want := range []string{
		`aria-label="Stock A exact OHLC values"`, `aria-label="Stock B exact OHLC values"`, `aria-label="Stock C exact OHLC values"`,
		"Stock A observations", "Stock B observations", "Stock C observations", "Traditional", "Outline",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("multiple-series output missing %q", want)
		}
	}
}

func TestCandlestickNewOptionValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{"mixed legacy and series", func(cfg *Config) { cfg.Series = multipleSeriesConfig().Series }, "cannot combine"},
		{"empty series", func(cfg *Config) { cfg.SeriesName, cfg.Data = "", nil; cfg.Series = []Series{} }, "at least one series"},
		{"series name", func(cfg *Config) { *cfg = multipleSeriesConfig(); cfg.Series[0].Name = "" }, "series 1 name"},
		{"duplicate series name", func(cfg *Config) { *cfg = multipleSeriesConfig(); cfg.Series[1].Name = " Stock A " }, `series name "Stock A" is duplicated`},
		{"series labels", func(cfg *Config) { *cfg = multipleSeriesConfig(); cfg.Series[1].Data[1].Label = "Other" }, "labels must align"},
		{"body style", func(cfg *Config) { *cfg = multipleSeriesConfig(); cfg.Series[0].BodyStyle = "gradient" }, "body style"},
		{"trend lines with multiple series", func(cfg *Config) {
			*cfg = multipleSeriesConfig()
			cfg.TrendLines = []TrendLine{{Type: TrendTypeSimpleMovingAverage, Period: 2}}
		}, "cannot combine"},
		{"patterns with multiple series", func(cfg *Config) {
			*cfg = multipleSeriesConfig()
			cfg.Patterns.Selection = PatternSelectionAll
		}, "cannot combine"},
		{"aggregation multi", func(cfg *Config) { *cfg = multipleSeriesConfig(); cfg.Aggregation.WindowSize = 2 }, "single series"},
		{"candle width", func(cfg *Config) { cfg.Options.Geometry.CandleWidth = 1.1 }, "candle width"},
		{"wick width", func(cfg *Config) { cfg.Options.Geometry.WickWidth = math.NaN() }, "wick width"},
		{"series gap", func(cfg *Config) { value := -.1; cfg.Options.Geometry.SeriesGap = &value }, "series gap"},
		{"pattern selection and enabled", func(cfg *Config) {
			cfg.Patterns.Selection = PatternSelectionAll
			cfg.Patterns.Enabled = []PatternType{PatternTypeDoji}
		}, "either selection or enabled"},
		{"pattern duplicate", func(cfg *Config) {
			cfg.Patterns.Selection = ""
			cfg.Patterns.Enabled = []PatternType{PatternTypeDoji, PatternTypeDoji}
		}, "duplicated"},
		{"pattern unsupported", func(cfg *Config) { cfg.Patterns.Selection = ""; cfg.Patterns.Enabled = []PatternType{"gap"} }, "unsupported"},
		{"doji threshold", func(cfg *Config) { cfg.Patterns.Selection, cfg.Patterns.DojiThreshold = PatternSelectionAll, 1.1 }, "doji threshold"},
		{"shadow tolerance", func(cfg *Config) { cfg.Patterns.Selection, cfg.Patterns.ShadowTolerance = PatternSelectionAll, -.1 }, "shadow tolerance"},
		{"shadow ratio", func(cfg *Config) { cfg.Patterns.Selection, cfg.Patterns.ShadowRatio = PatternSelectionAll, math.Inf(1) }, "shadow ratio"},
		{"engulfing minimum", func(cfg *Config) { cfg.Patterns.Selection, cfg.Patterns.EngulfingMinSize = PatternSelectionAll, -.1 }, "engulfing minimum"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.Data = append([]Datum(nil), cfg.Data...)
			test.edit(&cfg)
			if err := cfg.validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate = %v, want %q", err, test.want)
			}
		})
	}
}
