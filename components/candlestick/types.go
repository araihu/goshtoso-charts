// Package candlestick renders accessible server-side SVG candlestick charts.
package candlestick

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Datum is one open-high-low-close observation at a named category or time.
type Datum struct {
	Label string
	Open  float64
	High  float64
	Low   float64
	Close float64
}

// Axis controls renderer-neutral axis presentation.
type Axis struct{ Title string }

// TrendType selects a supported close-price trend calculation.
type TrendType string

const (
	// TrendTypeBollingerUpper draws SMA plus two population standard deviations.
	TrendTypeBollingerUpper TrendType = "bollinger-upper"
	// TrendTypeSimpleMovingAverage draws a centered simple moving average.
	TrendTypeSimpleMovingAverage TrendType = "simple-moving-average"
	// TrendTypeBollingerLower draws SMA minus two population standard deviations.
	TrendTypeBollingerLower TrendType = "bollinger-lower"
)

// TrendLine configures one close-price trend. Color and Class are mutually
// exclusive presentation overrides; their zero values use the chart theme.
type TrendLine struct {
	Type   TrendType
	Period int
	Color  string
	Class  string
}

// CandleStyle customizes increasing or decreasing candle marks. Color and
// Class are mutually exclusive; their zero values use semantic theme tokens.
type CandleStyle struct {
	Color string
	Class string
}

// BodyStyle selects how candle bodies are painted. Its zero value uses filled
// bodies. Values describe visual intent and do not expose renderer constants.
type BodyStyle string

const (
	BodyStyleFilled      BodyStyle = "filled"
	BodyStyleTraditional BodyStyle = "traditional"
	BodyStyleOutline     BodyStyle = "outline"
)

// Series is one named, aligned OHLC population. ShowWicks overrides the
// chart-level wick setting when non-nil.
type Series struct {
	Name      string
	Data      []Datum
	BodyStyle BodyStyle
	ShowWicks *bool
}

// PatternType identifies a supported candlestick formation. Values are stable
// renderer-neutral identifiers, not names from a drawing package.
type PatternType string

const (
	PatternTypeDoji             PatternType = "doji"
	PatternTypeHammer           PatternType = "hammer"
	PatternTypeInvertedHammer   PatternType = "inverted-hammer"
	PatternTypeShootingStar     PatternType = "shooting-star"
	PatternTypeGravestoneDoji   PatternType = "gravestone-doji"
	PatternTypeDragonflyDoji    PatternType = "dragonfly-doji"
	PatternTypeBullishMarubozu  PatternType = "bullish-marubozu"
	PatternTypeBearishMarubozu  PatternType = "bearish-marubozu"
	PatternTypeBullishEngulfing PatternType = "bullish-engulfing"
	PatternTypeBearishEngulfing PatternType = "bearish-engulfing"
	PatternTypePiercingLine     PatternType = "piercing-line"
	PatternTypeDarkCloudCover   PatternType = "dark-cloud-cover"
	PatternTypeMorningStar      PatternType = "morning-star"
	PatternTypeEveningStar      PatternType = "evening-star"
)

// PatternSelection selects a deterministic built-in vocabulary subset.
type PatternSelection string

const (
	PatternSelectionAll      PatternSelection = "all"
	PatternSelectionCore     PatternSelection = "core"
	PatternSelectionBullish  PatternSelection = "bullish"
	PatternSelectionBearish  PatternSelection = "bearish"
	PatternSelectionReversal PatternSelection = "reversal"
	PatternSelectionTrend    PatternSelection = "trend"
)

// PatternLabelText selects safe, built-in text formatting for visual labels.
type PatternLabelText string

const (
	PatternLabelTextName          PatternLabelText = "name"
	PatternLabelTextNameWithCount PatternLabelText = "name-with-count"
)

// PatternLabelStyle customizes rendered pattern labels. Empty Color and
// BackgroundColor use chart-theme tokens. Class is added to label text. No
// callbacks are accepted.
type PatternLabelStyle struct {
	Text            PatternLabelText
	Color           string
	Class           string
	BackgroundColor string
	FontSize        float64
	CornerRadius    int
}

// CloseReferenceType selects a close-value reference line.
type CloseReferenceType string

const (
	CloseReferenceAverage CloseReferenceType = "average"
	CloseReferenceMinimum CloseReferenceType = "minimum"
)

// PatternOptions configures detection and annotation as one Candlestick
// behavior variant. Its zero value leaves pattern detection disabled.
type PatternOptions struct {
	Selection    PatternSelection
	Enabled      []PatternType
	PreferLabels bool
	Label        PatternLabelStyle
	References   []CloseReferenceType
	// Threshold zero values preserve the pinned renderer defaults: 0.05 for
	// doji bodies, 0.01 for shadow tolerance, 2 for shadow ratio, and 1 for
	// engulfing size. The upstream API treats every non-positive value as its
	// default, so explicit zero is not a distinct supported setting.
	DojiThreshold    float64
	ShadowTolerance  float64
	ShadowRatio      float64
	EngulfingMinSize float64
}

// AggregationOptions compares source candles with candles grouped into fixed
// ordered windows. Its zero value leaves aggregation disabled.
type AggregationOptions struct {
	WindowSize int
	Title      string
	SeriesName string
}

// PatternResult is one detected formation. Results are ordered first by datum
// order and then by the selected vocabulary order when several patterns match.
type PatternResult struct {
	Index int
	Label string
	Type  PatternType
	Name  string
}

// Padding controls chart inset in pixels. Its zero value keeps renderer defaults.
type Padding struct{ Top, Right, Bottom, Left int }

// Geometry controls finite chart-specific candle geometry. Zero values keep
// upstream defaults: 0.8 body width and 1-pixel wicks. Pointer fields retain
// the distinction between an automatic default and an explicit false or zero.
type Geometry struct {
	CandleWidth float64
	WickWidth   float64
	SeriesGap   *float64
	ShowWicks   *bool
}

// Options controls renderer-neutral candlestick presentation.
type Options struct {
	TitleFontSize float64
	YUnit         float64
	LegendHidden  bool
	Padding       Padding
	Increasing    CandleStyle
	Decreasing    CandleStyle
	Geometry      Geometry
}

// Config describes an SSR SVG candlestick chart.
type Config struct {
	Label      string
	Caption    string
	Title      string
	SeriesName string
	Data       []Datum
	// Series provides multiple aligned OHLC populations. It cannot be combined
	// with legacy single-series SeriesName, Data, TrendLines, or Patterns fields.
	Series      []Series
	TrendLines  []TrendLine
	Patterns    PatternOptions
	Aggregation AggregationOptions
	XAxis       Axis
	YAxis       Axis
	Options     Options
	Width       int
	Height      int
	Style       charttheme.Style
	RootAttrs   templ.Attributes
	// Controls configures shared controls; Expand defaults on while fullscreen defaults off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("candlestick chart label is required")
	}
	if len(cfg.Series) > 0 && (strings.TrimSpace(cfg.SeriesName) != "" || len(cfg.Data) > 0 || len(cfg.TrendLines) > 0 || patternOptionsConfigured(cfg.Patterns)) {
		return fmt.Errorf("candlestick chart cannot combine Series with single-series SeriesName, Data, TrendLines, or Patterns")
	}
	series := cfg.resolvedSeries()
	if len(series) == 0 {
		return fmt.Errorf("candlestick chart needs at least one series")
	}
	if len(cfg.Series) == 0 && strings.TrimSpace(cfg.SeriesName) == "" {
		return fmt.Errorf("candlestick chart series name is required")
	}
	if len(cfg.Series) == 0 && len(cfg.Data) == 0 {
		return fmt.Errorf("candlestick chart needs at least one datum")
	}
	if len(series) > 1 && cfg.Aggregation.WindowSize > 0 {
		return fmt.Errorf("candlestick chart aggregation supports a single series")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("candlestick chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("candlestick chart height cannot be negative")
	}
	if !finite(cfg.Options.TitleFontSize) || cfg.Options.TitleFontSize < 0 {
		return fmt.Errorf("candlestick chart title font size must be finite and non-negative")
	}
	if !finite(cfg.Options.YUnit) || cfg.Options.YUnit < 0 {
		return fmt.Errorf("candlestick chart Y unit must be finite and non-negative")
	}
	if negativePadding(cfg.Options.Padding) {
		return fmt.Errorf("candlestick chart padding cannot be negative")
	}
	if err := validateGeometry(cfg.Options.Geometry); err != nil {
		return err
	}
	if cfg.Aggregation.WindowSize == 1 || cfg.Aggregation.WindowSize < 0 {
		return fmt.Errorf("candlestick chart aggregation window size must be zero or at least 2")
	}
	if cfg.Aggregation.WindowSize == 0 && (strings.TrimSpace(cfg.Aggregation.Title) != "" || strings.TrimSpace(cfg.Aggregation.SeriesName) != "") {
		return fmt.Errorf("candlestick chart aggregation options need a window size")
	}
	if err := validateCandleStyle("increasing", cfg.Options.Increasing); err != nil {
		return err
	}
	if err := validateCandleStyle("decreasing", cfg.Options.Decreasing); err != nil {
		return err
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("candlestick chart root attribute %q is reserved", attribute)
			}
		}
	}
	seriesNames := make(map[string]struct{}, len(series))
	for index, candidate := range series {
		name := strings.TrimSpace(candidate.Name)
		if name == "" {
			if len(cfg.Series) == 0 {
				return fmt.Errorf("candlestick chart series name is required")
			}
			return fmt.Errorf("candlestick chart series %d name is required", index+1)
		}
		if _, exists := seriesNames[name]; exists {
			return fmt.Errorf("candlestick chart series name %q is duplicated", name)
		}
		seriesNames[name] = struct{}{}
		if err := validateBodyStyle(index, candidate.BodyStyle); err != nil {
			return err
		}
		if err := validateData(candidate.Data); err != nil {
			return err
		}
		if index > 0 && !labelsAlign(series[0].Data, candidate.Data) {
			return fmt.Errorf("candlestick chart series %d labels must align with series 1", index+1)
		}
	}
	trendTypes := make(map[TrendType]int, len(cfg.TrendLines))
	bandPeriod := 0
	for index, trend := range cfg.TrendLines {
		switch trend.Type {
		case TrendTypeBollingerUpper, TrendTypeSimpleMovingAverage, TrendTypeBollingerLower:
		default:
			return fmt.Errorf("candlestick chart trend line %d type %q is unsupported", index+1, trend.Type)
		}
		if trend.Period < 2 || trend.Period > len(series[0].Data) {
			return fmt.Errorf("candlestick chart trend line %q period must be between 2 and datum count", trend.Type)
		}
		if _, exists := trendTypes[trend.Type]; exists {
			return fmt.Errorf("candlestick chart trend line type %q is duplicated", trend.Type)
		}
		trendTypes[trend.Type] = trend.Period
		if strings.TrimSpace(trend.Color) != "" && strings.TrimSpace(trend.Class) != "" {
			return fmt.Errorf("candlestick chart trend line %q cannot set both color and class", trend.Type)
		}
		if unsafeCSS(trend.Color) {
			return fmt.Errorf("candlestick chart trend line %q color is unsafe", trend.Type)
		}
		if unsafeClass(trend.Class) {
			return fmt.Errorf("candlestick chart trend line %q class is unsafe", trend.Type)
		}
		if trend.Type == TrendTypeBollingerUpper || trend.Type == TrendTypeBollingerLower {
			if bandPeriod != 0 && bandPeriod != trend.Period {
				return fmt.Errorf("candlestick chart Bollinger band periods conflict")
			}
			bandPeriod = trend.Period
		}
	}
	if err := validatePatternOptions(cfg.Patterns); err != nil {
		return err
	}
	return nil
}

func validatePatternOptions(options PatternOptions) error {
	if options.Selection == "" && len(options.Enabled) == 0 {
		if patternOptionsConfiguredExceptSelection(options) {
			return fmt.Errorf("candlestick chart pattern options need a selection")
		}
		return nil
	}
	if options.Selection != "" && len(options.Enabled) > 0 {
		return fmt.Errorf("candlestick chart patterns must use either selection or enabled patterns")
	}
	if options.Selection != "" && !validPatternSelection(options.Selection) {
		return fmt.Errorf("candlestick chart pattern selection %q is unsupported", options.Selection)
	}
	seenPatterns := make(map[PatternType]struct{}, len(options.Enabled))
	for _, pattern := range options.Enabled {
		if _, exists := patternNames[pattern]; !exists {
			return fmt.Errorf("candlestick chart pattern %q is unsupported", pattern)
		}
		if _, exists := seenPatterns[pattern]; exists {
			return fmt.Errorf("candlestick chart pattern %q is duplicated", pattern)
		}
		seenPatterns[pattern] = struct{}{}
	}
	if options.Label.Text != "" && options.Label.Text != PatternLabelTextName && options.Label.Text != PatternLabelTextNameWithCount {
		return fmt.Errorf("candlestick chart pattern label text %q is unsupported", options.Label.Text)
	}
	if !finite(options.Label.FontSize) || options.Label.FontSize < 0 {
		return fmt.Errorf("candlestick chart pattern label font size must be finite and non-negative")
	}
	if options.Label.CornerRadius < 0 {
		return fmt.Errorf("candlestick chart pattern label corner radius cannot be negative")
	}
	if !finite(options.DojiThreshold) || options.DojiThreshold < 0 || options.DojiThreshold > 1 {
		return fmt.Errorf("candlestick chart doji threshold must be finite and between 0 and 1")
	}
	if !finite(options.ShadowTolerance) || options.ShadowTolerance < 0 || options.ShadowTolerance > 1 {
		return fmt.Errorf("candlestick chart shadow tolerance must be finite and between 0 and 1")
	}
	if !finite(options.ShadowRatio) || options.ShadowRatio < 0 {
		return fmt.Errorf("candlestick chart shadow ratio must be finite and non-negative")
	}
	if !finite(options.EngulfingMinSize) || options.EngulfingMinSize < 0 {
		return fmt.Errorf("candlestick chart engulfing minimum size must be finite and non-negative")
	}
	for name, value := range map[string]string{"color": options.Label.Color, "class": options.Label.Class, "background color": options.Label.BackgroundColor} {
		if (strings.Contains(name, "color") && unsafeCSS(value)) || (strings.Contains(name, "class") && unsafeClass(value)) {
			return fmt.Errorf("candlestick chart pattern label %s is unsafe", name)
		}
	}
	references := make(map[CloseReferenceType]struct{}, len(options.References))
	for _, reference := range options.References {
		if reference != CloseReferenceAverage && reference != CloseReferenceMinimum {
			return fmt.Errorf("candlestick chart close reference %q is unsupported", reference)
		}
		if _, exists := references[reference]; exists {
			return fmt.Errorf("candlestick chart close reference %q is duplicated", reference)
		}
		references[reference] = struct{}{}
	}
	return nil
}

func validPatternSelection(selection PatternSelection) bool {
	switch selection {
	case PatternSelectionAll, PatternSelectionCore, PatternSelectionBullish, PatternSelectionBearish, PatternSelectionReversal, PatternSelectionTrend:
		return true
	default:
		return false
	}
}

func patternOptionsConfigured(options PatternOptions) bool {
	return options.Selection != "" || len(options.Enabled) > 0 || patternOptionsConfiguredExceptSelection(options)
}

func patternOptionsConfiguredExceptSelection(options PatternOptions) bool {
	return len(options.References) > 0 || options.PreferLabels || options.Label != (PatternLabelStyle{}) ||
		options.DojiThreshold != 0 || options.ShadowTolerance != 0 || options.ShadowRatio != 0 || options.EngulfingMinSize != 0
}

func validateGeometry(geometry Geometry) error {
	if !finite(geometry.CandleWidth) || geometry.CandleWidth < 0 || geometry.CandleWidth > 1 {
		return fmt.Errorf("candlestick chart candle width must be finite and between 0 and 1")
	}
	if !finite(geometry.WickWidth) || geometry.WickWidth < 0 {
		return fmt.Errorf("candlestick chart wick width must be finite and non-negative")
	}
	if geometry.SeriesGap != nil && (!finite(*geometry.SeriesGap) || *geometry.SeriesGap < 0 || *geometry.SeriesGap > 1) {
		return fmt.Errorf("candlestick chart series gap must be finite and between 0 and 1")
	}
	return nil
}

func validateBodyStyle(index int, style BodyStyle) error {
	switch style {
	case "", BodyStyleFilled, BodyStyleTraditional, BodyStyleOutline:
		return nil
	default:
		return fmt.Errorf("candlestick chart series %d body style %q is unsupported", index+1, style)
	}
}

func validateData(data []Datum) error {
	if len(data) == 0 {
		return fmt.Errorf("candlestick chart needs at least one datum")
	}
	labels := make(map[string]struct{}, len(data))
	for index, datum := range data {
		if strings.TrimSpace(datum.Label) == "" {
			return fmt.Errorf("candlestick chart datum %d needs a label", index+1)
		}
		if _, exists := labels[datum.Label]; exists {
			return fmt.Errorf("candlestick chart datum label %q is duplicated", datum.Label)
		}
		labels[datum.Label] = struct{}{}
		for name, value := range map[string]float64{"open": datum.Open, "high": datum.High, "low": datum.Low, "close": datum.Close} {
			if !finite(value) {
				return fmt.Errorf("candlestick chart datum %q %s must be finite", datum.Label, name)
			}
		}
		if datum.Low > datum.Open || datum.Low > datum.Close {
			return fmt.Errorf("candlestick chart datum %q low must be less than or equal to open and close", datum.Label)
		}
		if datum.High < datum.Open || datum.High < datum.Close {
			return fmt.Errorf("candlestick chart datum %q high must be greater than or equal to open and close", datum.Label)
		}
	}
	return nil
}

func labelsAlign(reference, candidate []Datum) bool {
	if len(reference) != len(candidate) {
		return false
	}
	for index := range reference {
		if reference[index].Label != candidate[index].Label {
			return false
		}
	}
	return true
}

func (cfg Config) resolvedSeries() []Series {
	if len(cfg.Series) > 0 {
		return cfg.Series
	}
	if strings.TrimSpace(cfg.SeriesName) == "" && len(cfg.Data) == 0 {
		return nil
	}
	return []Series{{Name: cfg.SeriesName, Data: cfg.Data}}
}

func validateCandleStyle(name string, style CandleStyle) error {
	if strings.TrimSpace(style.Color) != "" && strings.TrimSpace(style.Class) != "" {
		return fmt.Errorf("candlestick chart %s candles cannot set both color and class", name)
	}
	if unsafeCSS(style.Color) {
		return fmt.Errorf("candlestick chart %s candle color is unsafe", name)
	}
	if unsafeClass(style.Class) {
		return fmt.Errorf("candlestick chart %s candle class is unsafe", name)
	}
	return nil
}

func negativePadding(padding Padding) bool {
	return padding.Top < 0 || padding.Right < 0 || padding.Bottom < 0 || padding.Left < 0
}

func unsafeCSS(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ContainsAny(value, ";{}<>\\\"") || strings.Contains(value, "url(") || strings.Contains(value, "expression(")
}

func unsafeClass(value string) bool { return strings.ContainsAny(value, "\"'<>;") }

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
func (cfg Config) width() int {
	if cfg.Width > 0 {
		return cfg.Width
	}
	return 600
}
func (cfg Config) height() int {
	if cfg.Height > 0 {
		return cfg.Height
	}
	return 400
}
