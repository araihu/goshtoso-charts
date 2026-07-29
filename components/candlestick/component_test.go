package candlestick

import (
	"bytes"
	"context"
	"crypto/sha256"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func validConfig() Config {
	return Config{
		Label: "Seven-day stock price", Title: "Candlestick Chart", SeriesName: "Stock Price",
		Data: []Datum{
			{Label: "Day 1", Open: 100, High: 110, Low: 95, Close: 105},
			{Label: "Day 2", Open: 105, High: 115, Low: 100, Close: 112},
			{Label: "Day 3", Open: 112, High: 118, Low: 108, Close: 115},
			{Label: "Day 4", Open: 115, High: 120, Low: 104, Close: 108},
		},
	}
}

func bollingerConfig() Config {
	return Config{
		Label: "Candlestick Chart with Bollinger Bands", Title: "Candlestick Chart with Bollinger Bands", SeriesName: "Price",
		Data: []Datum{
			{Label: "1", Open: 100, High: 110, Low: 95, Close: 105},
			{Label: "2", Open: 105, High: 115, Low: 100, Close: 112},
			{Label: "3", Open: 112, High: 118, Low: 108, Close: 115},
			{Label: "4", Open: 115, High: 125, Low: 110, Close: 120},
			{Label: "5", Open: 120, High: 130, Low: 115, Close: 125},
			{Label: "6", Open: 125, High: 135, Low: 120, Close: 130},
			{Label: "7", Open: 130, High: 140, Low: 125, Close: 135},
			{Label: "8", Open: 135, High: 145, Low: 130, Close: 140},
			{Label: "9", Open: 140, High: 150, Low: 135, Close: 145},
			{Label: "10", Open: 145, High: 155, Low: 140, Close: 150},
			{Label: "11", Open: 150, High: 160, Low: 145, Close: 148},
			{Label: "12", Open: 148, High: 153, Low: 143, Close: 146},
			{Label: "13", Open: 146, High: 151, Low: 141, Close: 144},
			{Label: "14", Open: 144, High: 149, Low: 139, Close: 142},
			{Label: "15", Open: 142, High: 147, Low: 137, Close: 140},
			{Label: "16", Open: 140, High: 145, Low: 135, Close: 138},
			{Label: "17", Open: 138, High: 143, Low: 133, Close: 136},
			{Label: "18", Open: 136, High: 141, Low: 131, Close: 134},
			{Label: "19", Open: 134, High: 139, Low: 129, Close: 132},
			{Label: "20", Open: 132, High: 137, Low: 127, Close: 130},
		},
		TrendLines: []TrendLine{
			{Type: TrendTypeBollingerUpper, Period: 5},
			{Type: TrendTypeSimpleMovingAverage, Period: 5},
			{Type: TrendTypeBollingerLower, Period: 5},
		},
		Options: Options{TitleFontSize: 18, YUnit: 1, Padding: Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}},
		Width:   800, Height: 600,
	}
}

func patternConfig() Config {
	return Config{
		Label: "Candlestick patterns", Title: "Candlestick Patterns", SeriesName: "Stock Price with Patterns",
		Data: []Datum{
			{Label: "1", Open: 100, High: 110, Low: 95, Close: 105}, {Label: "2", Open: 105, High: 108, Low: 102, Close: 105.1},
			{Label: "3", Open: 108, High: 109, Low: 98, Close: 107}, {Label: "4", Open: 107, High: 108, Low: 103, Close: 104},
			{Label: "5", Open: 102, High: 115, Low: 101, Close: 113}, {Label: "6", Open: 113, High: 125, Low: 112, Close: 114},
			{Label: "7", Open: 114, High: 118, Low: 113, Close: 117}, {Label: "8", Open: 119, High: 120, Low: 108, Close: 110},
			{Label: "9", Open: 110, High: 113, Low: 107, Close: 109.9}, {Label: "10", Open: 109, High: 118, Low: 108, Close: 116},
		},
		Patterns: PatternOptions{Selection: PatternSelectionAll, PreferLabels: true, Label: PatternLabelStyle{Text: PatternLabelTextNameWithCount, Color: "#ffffff", Class: "pattern-text", BackgroundColor: "#1d4ed8", FontSize: 8, CornerRadius: 2}, References: []CloseReferenceType{CloseReferenceAverage, CloseReferenceMinimum}},
		Width:    900, Height: 650,
	}
}

func aggregationConfig() Config {
	return Config{
		Label: "Candlestick aggregation", Title: "1-Minute Candles (Before Aggregation)", SeriesName: "1-Minute",
		Data: []Datum{
			{Label: "1", Open: 100, High: 102, Low: 99, Close: 101}, {Label: "2", Open: 101, High: 103, Low: 100, Close: 102},
			{Label: "3", Open: 102, High: 105, Low: 101, Close: 104}, {Label: "4", Open: 104, High: 106, Low: 103, Close: 105},
			{Label: "5", Open: 105, High: 107, Low: 104, Close: 106}, {Label: "6", Open: 106, High: 108, Low: 105, Close: 107},
			{Label: "7", Open: 107, High: 109, Low: 106, Close: 108}, {Label: "8", Open: 108, High: 110, Low: 107, Close: 109},
			{Label: "9", Open: 109, High: 111, Low: 108, Close: 110}, {Label: "10", Open: 110, High: 112, Low: 109, Close: 111},
			{Label: "11", Open: 111, High: 113, Low: 110, Close: 112}, {Label: "12", Open: 112, High: 114, Low: 111, Close: 113},
			{Label: "13", Open: 113, High: 115, Low: 112, Close: 114}, {Label: "14", Open: 114, High: 116, Low: 113, Close: 115},
			{Label: "15", Open: 115, High: 117, Low: 114, Close: 116},
		},
		Aggregation: AggregationOptions{WindowSize: 5, Title: "5-Minute Aggregated Candles", SeriesName: "5-Minute"},
		Options:     Options{TitleFontSize: 16, YUnit: 1, Padding: Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}},
		Width:       1200, Height: 800,
	}
}

func TestCandlestickRendersAccessibleSSRAndExactValues(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Caption = "Seven daily OHLC observations."
	cfg.XAxis = Axis{Title: "Day"}
	cfg.YAxis = Axis{Title: "Price"}
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "mx-auto"}
	cfg.RootAttrs = templ.Attributes{"id": "stock-price", "data-chart-purpose": "ohlc"}

	instance := Candlestick(cfg)
	if instance.Kind() != chartcomponents.KindCandlestickChart {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`<figure class="goshtoso-charts-candlestick goshtoso-charts-palette goshtoso-charts-palette-araihu mx-auto" role="img" aria-label="Seven-day stock price"`,
		`id="stock-price"`, `data-chart-purpose="ohlc"`, "<svg", "Candlestick Chart", "Stock Price", "Day", "Price",
		"Seven daily OHLC observations.", "Exact OHLC values", "Increase means close", "Day 4", "Decrease", "115", "120", "104", "108",
		"var(--color-chart-increasing)", "var(--color-chart-decreasing)", "min-width: 37.5rem",
		`data-goshtoso-chart-expand`, `-chart-expand-export"`,
		`>SVG</button>`, `>PNG</button>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"rgb(145,204,117)", "rgb(238,102,102)"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
}

func TestCandlestickMapsTypedOHLCAndPresentation(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.XAxis.Title = "Session"
	cfg.YAxis.Title = "Value"
	options := candlestickOptions(cfg)
	if options.Title.Text != "Candlestick Chart" || options.XAxis.Title != "Session" || options.YAxis[0].Title != "Value" {
		t.Fatalf("titles not mapped: %#v", options)
	}
	if got := options.XAxis.Labels; len(got) != 4 || got[3] != "Day 4" {
		t.Fatalf("labels = %v", got)
	}
	if got := options.SeriesList[0]; got.Name != "Stock Price" || got.Data[3].Open != 115 || got.Data[3].High != 120 || got.Data[3].Low != 104 || got.Data[3].Close != 108 {
		t.Fatalf("series = %#v", got)
	}
	if options.CandleWidth != 0.8 || options.ShowWicks != nil {
		t.Fatalf("upstream defaults changed: width=%v wicks=%v", options.CandleWidth, options.ShowWicks)
	}
}

func TestCandlestickDefaultsToUpstreamDimensions(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	if cfg.width() != 600 || cfg.height() != 400 {
		t.Fatalf("dimensions = %dx%d, want 600x400", cfg.width(), cfg.height())
	}
}

func TestCandlestickAggregationPreservesSourceAndWindowSemantics(t *testing.T) {
	t.Parallel()
	cfg := aggregationConfig()
	got := aggregateData(cfg.Data, cfg.Aggregation.WindowSize)
	want := []Datum{
		{Label: "1-5", Open: 100, High: 107, Low: 99, Close: 106},
		{Label: "6-10", Open: 106, High: 112, Low: 105, Close: 111},
		{Label: "11-15", Open: 111, High: 117, Low: 110, Close: 116},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("aggregated data = %#v, want %#v", got, want)
	}
	if len(cfg.Data) != 15 || cfg.Data[0].Open != 100 || cfg.Data[14].Close != 116 {
		t.Fatalf("source data mutated: %#v", cfg.Data)
	}
	partial := aggregateData(cfg.Data[:12], 5)
	if got, want := partial[2], (Datum{Label: "11-12", Open: 111, High: 114, Low: 110, Close: 113}); got != want {
		t.Fatalf("partial final window = %#v, want %#v", got, want)
	}
}

func TestCandlestickAggregationRendersBeforeAfterAndExactValues(t *testing.T) {
	t.Parallel()
	cfg := aggregationConfig()
	cfg.Caption = "Fifteen one-minute observations grouped into three five-minute candles."
	var output bytes.Buffer
	if err := Candlestick(cfg).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	markup := output.String()
	for _, want := range []string{
		`viewBox="0 0 1200 800"`, "1-Minute Candles (Before Aggregation)", "5-Minute Aggregated Candles",
		"1-Minute exact OHLC values", "5-Minute exact OHLC values", "1-5", "6-10", "11-15",
		"Fifteen one-minute observations grouped into three five-minute candles.", "var(--color-chart-surface)",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("aggregation markup missing %q", want)
		}
	}
}

func TestCandlestickDefaultSVGCompatibilityHash(t *testing.T) {
	t.Parallel()
	svg, err := renderSVG(validConfig())
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256.Sum256([]byte(svg)); got != [32]byte{0xcc, 0x9e, 0x22, 0xaf, 0xf7, 0xbc, 0x9d, 0xcf, 0x15, 0x63, 0xd4, 0x87, 0xfc, 0x63, 0xfa, 0x31, 0x82, 0xca, 0xd3, 0x9f, 0x6b, 0x70, 0xc0, 0xee, 0xb4, 0x5c, 0xbb, 0x22, 0x19, 0x8e, 0x93, 0x05} {
		t.Fatalf("default SVG compatibility SHA-256 = %x", got)
	}
}

func TestCandlestickMapsBollingerTreatmentAndPresentation(t *testing.T) {
	t.Parallel()
	cfg := bollingerConfig()
	options := candlestickOptions(cfg)
	if len(options.SeriesList[0].Data) != 20 || options.XAxis.Labels[0] != "1" || options.XAxis.Labels[19] != "20" {
		t.Fatalf("source shape drifted: data=%d labels=%v", len(options.SeriesList[0].Data), options.XAxis.Labels)
	}
	if options.Title.Text != cfg.Title || options.Title.FontStyle.FontSize != 18 || options.YAxis[0].Unit != 1 {
		t.Fatalf("title/Y presentation = %#v / %#v", options.Title, options.YAxis[0])
	}
	if options.Legend.Show == nil || !*options.Legend.Show ||
		options.Padding.Top != 20 || options.Padding.Right != 20 || options.Padding.Bottom != 20 || options.Padding.Left != 20 {
		t.Fatalf("legend/padding = %#v / %#v", options.Legend.Show, options.Padding)
	}
	wantTypes := []string{"bollinger_upper", "sma", "bollinger_lower"}
	for index, trend := range options.SeriesList[0].CloseTrendLine {
		if string(trend.Type) != wantTypes[index] || trend.Period != 5 {
			t.Fatalf("trend %d = %#v", index, trend)
		}
	}
}

func TestCandlestickComputesPinnedPeriodFiveBands(t *testing.T) {
	t.Parallel()
	cfg := bollingerConfig()
	rows := computedTrendValues(cfg)
	want := [][3]float64{
		{119.04653672665103, 110.66666666666667, 102.28679660668232},
		{123.86278049120021, 113, 102.13721950879979},
		{129.0586968631711, 115.4, 101.7413031368289},
		{133.4598621738516, 120.4, 107.34013782614839},
		{139.14213562373095, 125, 110.85786437626905},
		{144.14213562373095, 130, 115.85786437626905},
		{149.14213562373095, 135, 120.85786437626905},
		{154.14213562373095, 140, 125.85786437626905},
		{154.52520022699812, 143.6, 132.67479977300187},
		{152.54091981854108, 145.8, 139.05908018145894},
		{150.9081318457076, 146.6, 142.2918681542924},
		{151.65685424949237, 146, 140.34314575050763},
		{149.65685424949237, 144, 138.34314575050763},
		{147.65685424949237, 142, 136.34314575050763},
		{145.65685424949237, 140, 134.34314575050763},
		{143.65685424949237, 138, 132.34314575050763},
		{141.65685424949237, 136, 130.34314575050763},
		{139.65685424949237, 134, 128.34314575050763},
		{137.47213595499957, 133, 128.52786404500043},
		{135.2659863237109, 132, 128.7340136762891},
	}
	for row := range want {
		if len(rows[row]) != 3 {
			t.Fatalf("row %d trend count = %d", row, len(rows[row]))
		}
		for column := range want[row] {
			if math.Abs(rows[row][column].Value-want[row][column]) > 1e-12 {
				t.Fatalf("row %d trend %d = %.15f, want %.15f", row+1, column, rows[row][column].Value, want[row][column])
			}
		}
	}
}

func TestCandlestickAppliesCallerColorsAndClassesToCandlesAndTrends(t *testing.T) {
	t.Parallel()
	cfg := bollingerConfig()
	cfg.Options.Increasing.Color = "#14532d"
	cfg.Options.Decreasing.Class = "caller-decreasing"
	cfg.TrendLines[0].Color = "#1d4ed8"
	cfg.TrendLines[1].Class = "caller-middle"
	svg, err := renderSVG(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"#14532d", "caller-decreasing", "#1d4ed8", "caller-middle", "var(--color-chart-bollinger-lower)"} {
		if !strings.Contains(svg, want) {
			t.Errorf("SVG missing %q", want)
		}
	}
	for _, unwanted := range []string{"rgb(5,5,5)", "rgb(6,6,6)", "rgb(7,7,7)"} {
		if strings.Contains(svg, unwanted) {
			t.Errorf("SVG leaked sentinel %q", unwanted)
		}
	}
}

func TestCandlestickPatternsUseTypedVocabularyDeterministicOrderAndAccessibleEvidence(t *testing.T) {
	t.Parallel()
	cfg := patternConfig()
	results, err := DetectPatterns(cfg.Data, cfg.Patterns)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]PatternType, len(results))
	for index, result := range results {
		got[index] = result.Type
	}
	want := []PatternType{PatternTypeDoji, PatternTypeHammer, PatternTypeBullishEngulfing, PatternTypeShootingStar, PatternTypeInvertedHammer, PatternTypeBearishEngulfing, PatternTypeDoji, PatternTypeBullishEngulfing}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pattern order = %v, want %v", got, want)
	}
	if results[3].Label != "6" || results[3].Name != "Shooting Star" || results[4].Name != "Inverted Hammer" {
		t.Fatalf("multi-match = %#v", results[3:5])
	}
	var output bytes.Buffer
	if err := Candlestick(cfg).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Detected patterns", "Shooting Star, Inverted Hammer", "Bullish Engulfing", "Bearish Engulfing", "#ffffff", "#1d4ed8", "pattern-text"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("rendered patterns missing %q", want)
		}
	}
}

func TestCandlestickPatternSelectionsAndValidation(t *testing.T) {
	t.Parallel()
	cfg := patternConfig()
	for selection, wantCount := range map[PatternSelection]int{PatternSelectionAll: 8, PatternSelectionCore: 5, PatternSelectionBullish: 4} {
		cfg.Patterns.Selection = selection
		results, err := DetectPatterns(cfg.Data, cfg.Patterns)
		if err != nil || len(results) != wantCount {
			t.Errorf("%s = %d, %v; want %d", selection, len(results), err, wantCount)
		}
	}
	for _, test := range []struct {
		name, want string
		edit       func(*Config)
	}{
		{"selection", "unsupported", func(c *Config) { c.Patterns.Selection = "bearish" }},
		{"label text", "label text", func(c *Config) { c.Patterns.Label.Text = "raw" }},
		{"font", "font size", func(c *Config) { c.Patterns.Label.FontSize = math.NaN() }},
		{"reference", "close reference", func(c *Config) { c.Patterns.References = []CloseReferenceType{"median"} }},
		{"duplicate reference", "duplicated", func(c *Config) {
			c.Patterns.References = []CloseReferenceType{CloseReferenceMinimum, CloseReferenceMinimum}
		}},
		{"unsafe color", "unsafe", func(c *Config) { c.Patterns.Label.Color = "red;fill:blue" }},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			cfg := patternConfig()
			test.edit(&cfg)
			if err := cfg.validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCandlestickValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "label", edit: func(c *Config) { c.Label = "" }, want: "label is required"},
		{name: "series", edit: func(c *Config) { c.SeriesName = "" }, want: "series name is required"},
		{name: "data", edit: func(c *Config) { c.Data = nil }, want: "at least one datum"},
		{name: "datum label", edit: func(c *Config) { c.Data[0].Label = "" }, want: "datum 1 needs a label"},
		{name: "duplicate label", edit: func(c *Config) { c.Data[1].Label = "Day 1" }, want: `datum label "Day 1" is duplicated`},
		{name: "open finite", edit: func(c *Config) { c.Data[0].Open = math.NaN() }, want: `datum "Day 1" open must be finite`},
		{name: "high finite", edit: func(c *Config) { c.Data[0].High = math.Inf(1) }, want: `datum "Day 1" high must be finite`},
		{name: "low above open", edit: func(c *Config) { c.Data[0].Low = 101 }, want: "low must be less than or equal"},
		{name: "low above close", edit: func(c *Config) { c.Data[0].Low = 106 }, want: "low must be less than or equal"},
		{name: "high below open", edit: func(c *Config) { c.Data[0].High = 99 }, want: "high must be greater than or equal"},
		{name: "high below close", edit: func(c *Config) { c.Data[0].High = 104 }, want: "high must be greater than or equal"},
		{name: "width", edit: func(c *Config) { c.Width = -1 }, want: "width cannot be negative"},
		{name: "height", edit: func(c *Config) { c.Height = -1 }, want: "height cannot be negative"},
		{name: "title font", edit: func(c *Config) { c.Options.TitleFontSize = math.NaN() }, want: "title font size"},
		{name: "Y unit", edit: func(c *Config) { c.Options.YUnit = -1 }, want: "Y unit"},
		{name: "padding", edit: func(c *Config) { c.Options.Padding.Left = -1 }, want: "padding cannot be negative"},
		{name: "aggregation window", edit: func(c *Config) { c.Aggregation.WindowSize = 1 }, want: "aggregation window size"},
		{name: "aggregation title without window", edit: func(c *Config) { c.Aggregation.Title = "Grouped" }, want: "need a window size"},
		{name: "candle exclusivity", edit: func(c *Config) { c.Options.Increasing = CandleStyle{Color: "red", Class: "up"} }, want: "both color and class"},
		{name: "unsupported trend", edit: func(c *Config) { c.TrendLines = []TrendLine{{Type: "median", Period: 2}} }, want: "unsupported"},
		{name: "trend period low", edit: func(c *Config) { c.TrendLines = []TrendLine{{Type: TrendTypeSimpleMovingAverage, Period: 1}} }, want: "between 2 and datum count"},
		{name: "trend period high", edit: func(c *Config) { c.TrendLines = []TrendLine{{Type: TrendTypeSimpleMovingAverage, Period: 5}} }, want: "between 2 and datum count"},
		{name: "duplicate trend", edit: func(c *Config) {
			c.TrendLines = []TrendLine{{Type: TrendTypeSimpleMovingAverage, Period: 2}, {Type: TrendTypeSimpleMovingAverage, Period: 3}}
		}, want: "duplicated"},
		{name: "conflicting bands", edit: func(c *Config) {
			c.TrendLines = []TrendLine{{Type: TrendTypeBollingerUpper, Period: 2}, {Type: TrendTypeBollingerLower, Period: 3}}
		}, want: "periods conflict"},
		{name: "trend exclusivity", edit: func(c *Config) {
			c.TrendLines = []TrendLine{{Type: TrendTypeSimpleMovingAverage, Period: 2, Color: "red", Class: "middle"}}
		}, want: "both color and class"},
		{name: "root attr", edit: func(c *Config) { c.RootAttrs = templ.Attributes{"ARIA-Label": "override"} }, want: `root attribute "ARIA-Label" is reserved`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := validConfig()
			cfg.Data = append([]Datum(nil), cfg.Data...)
			test.edit(&cfg)
			err := cfg.validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate() error = %v, want %q", err, test.want)
			}
		})
	}
}
