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

// Padding controls chart inset in pixels. Its zero value keeps renderer defaults.
type Padding struct{ Top, Right, Bottom, Left int }

// Options controls renderer-neutral candlestick presentation.
type Options struct {
	TitleFontSize float64
	YUnit         float64
	LegendHidden  bool
	Padding       Padding
	Increasing    CandleStyle
	Decreasing    CandleStyle
}

// Config describes an SSR SVG candlestick chart.
type Config struct {
	Label      string
	Caption    string
	Title      string
	SeriesName string
	Data       []Datum
	TrendLines []TrendLine
	XAxis      Axis
	YAxis      Axis
	Options    Options
	Width      int
	Height     int
	Style      charttheme.Style
	RootAttrs  templ.Attributes
	// Controls configures shared controls; Expand defaults on while fullscreen defaults off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("candlestick chart label is required")
	}
	if strings.TrimSpace(cfg.SeriesName) == "" {
		return fmt.Errorf("candlestick chart series name is required")
	}
	if len(cfg.Data) == 0 {
		return fmt.Errorf("candlestick chart needs at least one datum")
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
	labels := make(map[string]struct{}, len(cfg.Data))
	for index, datum := range cfg.Data {
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
	trendTypes := make(map[TrendType]int, len(cfg.TrendLines))
	bandPeriod := 0
	for index, trend := range cfg.TrendLines {
		switch trend.Type {
		case TrendTypeBollingerUpper, TrendTypeSimpleMovingAverage, TrendTypeBollingerLower:
		default:
			return fmt.Errorf("candlestick chart trend line %d type %q is unsupported", index+1, trend.Type)
		}
		if trend.Period < 2 || trend.Period > len(cfg.Data) {
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
	return nil
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
