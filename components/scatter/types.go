// Package scatter renders accessible server-side SVG scatter charts.
package scatter

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Point is one observation assigned to an ordered category.
type Point struct {
	Category string
	Value    float64
}

// Symbol selects the marker drawn for scatter points.
type Symbol string

const (
	// SymbolDefault inherits chart options or draws a dot when used globally.
	SymbolDefault Symbol = ""
	// SymbolDot draws a compact filled dot.
	SymbolDot Symbol = "dot"
	// SymbolCircle draws a hollow circle.
	SymbolCircle Symbol = "circle"
	// SymbolSquare draws a square.
	SymbolSquare Symbol = "square"
	// SymbolDiamond draws a diamond.
	SymbolDiamond Symbol = "diamond"
)

// Options controls scatter-specific marker presentation.
//
// Size is the marker radius in pixels. Zero selects the renderer default.
type Options struct {
	Symbol        Symbol
	Size          float64
	Trend         TrendLine
	ReferenceLine ReferenceLine
	ValueFormat   ValueFormat
}

// TrendKind selects a renderer-neutral statistical trend.
type TrendKind string

const (
	TrendNone TrendKind = ""
	// TrendSimpleMovingAverage draws a simple moving average over aligned categories.
	TrendSimpleMovingAverage TrendKind = "simple-moving-average"
)

// TrendLine configures an optional series trend.
type TrendLine struct {
	Kind   TrendKind
	Period int
}

// ReferenceLine selects a renderer-neutral statistical reference line.
type ReferenceLine string

const (
	ReferenceLineNone    ReferenceLine = ""
	ReferenceLineMaximum ReferenceLine = "maximum"
)

// ValueFormat selects renderer-neutral numeric label formatting.
type ValueFormat string

const (
	ValueFormatDefault   ValueFormat = ""
	ValueFormatHumanized ValueFormat = "humanized"
)

// CategoryAxisOptions controls dense categorical axis labels.
type CategoryAxisOptions struct {
	BoundaryGap   *bool
	LabelCount    int
	LabelFontSize float64
	// LabelRotation is expressed in degrees.
	LabelRotation float64
}

// ValueAxisOptions controls the numeric axis.
type ValueAxisOptions struct {
	Min           *float64
	Max           *float64
	Unit          float64
	LabelSkip     int
	LabelFontSize float64
}

// LegendOrientation controls legend flow.
type LegendOrientation string

const (
	LegendHorizontal LegendOrientation = ""
	LegendVertical   LegendOrientation = "vertical"
)

// HorizontalPlacement selects horizontal placement for titles and legends.
type HorizontalPlacement string

const (
	PlacementDefault HorizontalPlacement = ""
	PlacementCenter  HorizontalPlacement = "center"
	PlacementRight   HorizontalPlacement = "right"
)

// Alignment selects marker/text alignment inside a legend.
type Alignment string

const (
	AlignmentDefault Alignment = ""
	AlignmentRight   Alignment = "right"
	AlignmentCenter  Alignment = "center"
)

// LegendOptions controls legend layout without exposing renderer types.
type LegendOptions struct {
	Orientation LegendOrientation
	Placement   HorizontalPlacement
	Alignment   Alignment
	FontSize    float64
}

// TitleOptions controls visible chart title placement.
type TitleOptions struct {
	Text      string
	Placement HorizontalPlacement
}

// Padding is plot padding in pixels.
type Padding struct{ Top, Right, Bottom, Left int }

// Series is one named population of categorical points.
//
// Options override Config.Options for this series when their fields are set.
type Series struct {
	Name   string
	Points []Point
	// Values is aligned with Config.Categories. Each category may contain zero,
	// one, or multiple samples. Values and Points are mutually exclusive.
	Values  [][]float64
	Options Options
}

// Config describes an SSR SVG scatter chart.
//
// Label is required and becomes the accessible name of the rendered figure.
// Caption remains visible below the chart. Style controls the theme palette and
// caller root class. RootAttrs accepts additional figure attributes except
// class, role, and aria-label, which remain owned by the component.
type Config struct {
	Label      string
	Caption    string
	Categories []string
	Series     []Series
	Options    Options
	Title      TitleOptions
	Legend     LegendOptions
	XAxis      CategoryAxisOptions
	YAxis      ValueAxisOptions
	Padding    Padding
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
		return fmt.Errorf("scatter chart label is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("scatter chart needs at least one series")
	}
	if len(cfg.Categories) == 0 {
		return fmt.Errorf("scatter chart needs at least one category")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("scatter chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("scatter chart height cannot be negative")
	}
	if err := cfg.Options.validate("scatter chart options"); err != nil {
		return err
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("scatter chart root attribute %q is reserved", attribute)
			}
		}
	}
	categoryIndexes := make(map[string]int, len(cfg.Categories))
	for index, category := range cfg.Categories {
		if strings.TrimSpace(category) == "" {
			return fmt.Errorf("scatter chart category %d cannot be empty", index+1)
		}
		if _, exists := categoryIndexes[category]; exists {
			return fmt.Errorf("scatter chart category %q is duplicated", category)
		}
		categoryIndexes[category] = index
	}
	for index, series := range cfg.Series {
		if strings.TrimSpace(series.Name) == "" {
			return fmt.Errorf("scatter chart series %d needs a name", index+1)
		}
		if series.Points != nil && series.Values != nil {
			return fmt.Errorf("scatter chart series %q cannot use both points and aligned values", series.Name)
		}
		if series.Points == nil && series.Values == nil {
			return fmt.Errorf("scatter chart series %q needs points or aligned values", series.Name)
		}
		if err := series.Options.validate(fmt.Sprintf("scatter chart series %q options", series.Name)); err != nil {
			return err
		}
		resolvedOptions := series.Options.resolved(cfg.Options)
		if resolvedOptions.Trend.Kind != TrendNone && resolvedOptions.Trend.Period > len(cfg.Categories) {
			return fmt.Errorf("scatter chart series %q trend period cannot exceed category count", series.Name)
		}
		for pointIndex, point := range series.Points {
			if _, ok := categoryIndexes[point.Category]; !ok {
				return fmt.Errorf("scatter chart series %q point %d references unknown category %q", series.Name, pointIndex, point.Category)
			}
			if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				return fmt.Errorf("scatter chart series %q point %d must contain a finite value", series.Name, pointIndex)
			}
		}
		if series.Values != nil {
			if len(series.Values) != len(cfg.Categories) {
				return fmt.Errorf("scatter chart series %q aligned values length %d must match %d categories", series.Name, len(series.Values), len(cfg.Categories))
			}
			samples := 0
			for categoryIndex, values := range series.Values {
				for sampleIndex, value := range values {
					if math.IsNaN(value) || math.IsInf(value, 0) {
						return fmt.Errorf("scatter chart series %q category %d sample %d must contain a finite value", series.Name, categoryIndex, sampleIndex)
					}
					samples++
				}
			}
			if samples == 0 {
				return fmt.Errorf("scatter chart series %q needs at least one aligned value", series.Name)
			}
		}
	}
	return cfg.validateLayout()
}

func (options Options) validate(prefix string) error {
	switch options.Symbol {
	case SymbolDefault, SymbolDot, SymbolCircle, SymbolSquare, SymbolDiamond:
	default:
		return fmt.Errorf("%s has unsupported symbol %q", prefix, options.Symbol)
	}
	if math.IsNaN(options.Size) || math.IsInf(options.Size, 0) || options.Size < 0 {
		return fmt.Errorf("%s size must be a finite non-negative number", prefix)
	}
	switch options.Trend.Kind {
	case TrendNone:
		if options.Trend.Period != 0 {
			return fmt.Errorf("%s trend period requires a trend kind", prefix)
		}
	case TrendSimpleMovingAverage:
		if options.Trend.Period <= 0 {
			return fmt.Errorf("%s trend period must be positive", prefix)
		}
	default:
		return fmt.Errorf("%s has unsupported trend %q", prefix, options.Trend.Kind)
	}
	switch options.ReferenceLine {
	case ReferenceLineNone, ReferenceLineMaximum:
	default:
		return fmt.Errorf("%s has unsupported reference line %q", prefix, options.ReferenceLine)
	}
	switch options.ValueFormat {
	case ValueFormatDefault, ValueFormatHumanized:
	default:
		return fmt.Errorf("%s has unsupported value format %q", prefix, options.ValueFormat)
	}
	return nil
}

func (options Options) resolved(fallback Options) Options {
	if options.Symbol == SymbolDefault {
		options.Symbol = fallback.Symbol
	}
	if options.Size == 0 {
		options.Size = fallback.Size
	}
	if options.Trend.Kind == TrendNone {
		options.Trend = fallback.Trend
	}
	if options.ReferenceLine == ReferenceLineNone {
		options.ReferenceLine = fallback.ReferenceLine
	}
	if options.ValueFormat == ValueFormatDefault {
		options.ValueFormat = fallback.ValueFormat
	}
	return options
}

func (cfg Config) validateLayout() error {
	finiteNonNegative := func(name string, value float64) error {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return fmt.Errorf("scatter chart %s must be a finite non-negative number", name)
		}
		return nil
	}
	if err := finiteNonNegative("x-axis label font size", cfg.XAxis.LabelFontSize); err != nil {
		return err
	}
	if math.IsNaN(cfg.XAxis.LabelRotation) || math.IsInf(cfg.XAxis.LabelRotation, 0) || cfg.XAxis.LabelRotation < -360 || cfg.XAxis.LabelRotation > 360 {
		return fmt.Errorf("scatter chart x-axis label rotation must be finite and between -360 and 360 degrees")
	}
	if cfg.XAxis.LabelCount < 0 {
		return fmt.Errorf("scatter chart x-axis label count cannot be negative")
	}
	if err := finiteNonNegative("y-axis label font size", cfg.YAxis.LabelFontSize); err != nil {
		return err
	}
	if err := finiteNonNegative("y-axis unit", cfg.YAxis.Unit); err != nil {
		return err
	}
	if cfg.YAxis.LabelSkip < 0 {
		return fmt.Errorf("scatter chart y-axis label skip cannot be negative")
	}
	if cfg.YAxis.Min != nil && (math.IsNaN(*cfg.YAxis.Min) || math.IsInf(*cfg.YAxis.Min, 0)) {
		return fmt.Errorf("scatter chart y-axis minimum must be finite")
	}
	if cfg.YAxis.Max != nil && (math.IsNaN(*cfg.YAxis.Max) || math.IsInf(*cfg.YAxis.Max, 0)) {
		return fmt.Errorf("scatter chart y-axis maximum must be finite")
	}
	if cfg.YAxis.Min != nil && cfg.YAxis.Max != nil && *cfg.YAxis.Min >= *cfg.YAxis.Max {
		return fmt.Errorf("scatter chart y-axis minimum must be less than maximum")
	}
	if cfg.Title.Placement != PlacementDefault && cfg.Title.Placement != PlacementCenter {
		return fmt.Errorf("scatter chart title placement %q is not supported", cfg.Title.Placement)
	}
	if cfg.Legend.Orientation != LegendHorizontal && cfg.Legend.Orientation != LegendVertical {
		return fmt.Errorf("scatter chart legend orientation %q is not supported", cfg.Legend.Orientation)
	}
	if cfg.Legend.Placement != PlacementDefault && cfg.Legend.Placement != PlacementCenter && cfg.Legend.Placement != PlacementRight {
		return fmt.Errorf("scatter chart legend placement %q is not supported", cfg.Legend.Placement)
	}
	if cfg.Legend.Alignment != AlignmentDefault && cfg.Legend.Alignment != AlignmentRight && cfg.Legend.Alignment != AlignmentCenter {
		return fmt.Errorf("scatter chart legend alignment %q is not supported", cfg.Legend.Alignment)
	}
	if err := finiteNonNegative("legend font size", cfg.Legend.FontSize); err != nil {
		return err
	}
	if cfg.Padding.Top < 0 || cfg.Padding.Right < 0 || cfg.Padding.Bottom < 0 || cfg.Padding.Left < 0 {
		return fmt.Errorf("scatter chart padding cannot be negative")
	}
	return nil
}

func (cfg Config) width() int {
	if cfg.Width > 0 {
		return cfg.Width
	}
	return 720
}

func (cfg Config) height() int {
	if cfg.Height > 0 {
		return cfg.Height
	}
	return 320
}
