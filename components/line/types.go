// Package line renders accessible server-side SVG line charts.
package line

import (
	"fmt"
	"math"
	"strings"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Title controls the visible chart title.
type Title struct {
	Text      string
	Subtext   string
	Placement Placement
	FontSize  float64
}

// Placement selects horizontal placement without exposing renderer offsets.
type Placement string

const (
	PlacementDefault Placement = ""
	PlacementCenter  Placement = "center"
	PlacementRight   Placement = "right"
)

// SymbolShape selects the marker drawn at each non-missing point.
type SymbolShape string

const (
	SymbolDefault SymbolShape = ""
	SymbolNone    SymbolShape = "none"
	SymbolCircle  SymbolShape = "circle"
	SymbolDot     SymbolShape = "dot"
	SymbolSquare  SymbolShape = "square"
	SymbolDiamond SymbolShape = "diamond"
)

// Symbol controls point marker shape and radius. Zero size selects renderer defaults.
type Symbol struct {
	Shape SymbolShape
	Size  float64
}

// Point is one aligned value or an explicit missing observation.
// Missing points break the line and are reported as unavailable in exact data.
type Point struct {
	Value   float64
	Missing bool
}

// ValueFormat selects safe built-in number formatting for labels and references.
type ValueFormat string

const (
	ValueFormatDefault   ValueFormat = ""
	ValueFormatHumanized ValueFormat = "humanized"
)

// LabelColorScale selects semantic chart-theme tokens for data-label color.
type LabelColorScale string

const (
	LabelColorScaleNone       LabelColorScale = ""
	LabelColorScaleColdToWarm LabelColorScale = "cold-to-warm"
)

// DataLabelOptions controls labels at individual line points.
type DataLabelOptions struct {
	Show     bool
	Format   ValueFormat
	Decimals int
	// TrailingZeros preserves the requested decimal width for humanized values.
	TrailingZeros bool
	ColorScale    LabelColorScale
	FontSize      float64
}

// ReferenceStyle decorates the adjacent reference-evidence rows.
type ReferenceStyle struct {
	Class string
}

// References selects statistical annotations for one series. Average renders a
// reference line; Minimum and Maximum render reference points.
type References struct {
	Average     bool
	Minimum     bool
	Maximum     bool
	Format      ValueFormat
	Decimals    int
	PointSize   int
	PointPrefix string
	Style       ReferenceStyle
}

// TextAnnotation draws one label at chart-relative pixel coordinates. SeriesIndex
// selects a chart-theme color; Color overrides it and Class decorates the SVG text.
type TextAnnotation struct {
	Text        string
	X           int
	Y           int
	SeriesIndex int
	FontSize    float64
	Color       string
	Class       string
}

// AreaOptions controls optional fill below every line series.
// Opacity uses the renderer-neutral range 0..1. Zero selects the renderer's
// default opacity when Enabled is true.
type AreaOptions struct {
	Enabled bool
	Opacity float64
}

// CategoryAxisOptions controls the horizontal category axis.
// A nil BoundaryGap preserves the Line component's legacy behavior.
type CategoryAxisOptions struct {
	BoundaryGap   *bool
	Unit          float64
	LabelCount    int
	LabelFontSize float64
	// LabelRotation is expressed in degrees.
	LabelRotation float64
}

// LegendOptions controls legend layout.
type LegendOptions struct {
	Hidden  bool
	Padding Padding
}

// Padding is spacing in pixels.
type Padding struct{ Top, Right, Bottom, Left int }

// Axis controls one numeric Y axis. Unit is a suggested positive tick step.
// Min and Max optionally bound the axis and its assigned values. Color and
// Class are mutually exclusive presentation overrides.
type Axis struct {
	Title         string
	TitleFontSize float64
	Unit          float64
	Min           *float64
	Max           *float64
	Hidden        bool
	LabelCount    int
	LabelFontSize float64
	SpineLine     bool
	Color         string
	Class         string
}

// Series is one labeled sequence of values. Values must align with Config.Labels.
type Series struct {
	Name   string
	Values []float64
	// Points and Values are mutually exclusive. Points represent explicit gaps.
	Points     []Point
	YAxisIndex int
	Symbol     Symbol
	Labels     DataLabelOptions
	References References
	Color      string
	Class      string
}

// Config describes an SSR SVG line chart.
//
// Label is required and becomes the accessible name of the rendered figure.
// Caption remains visible below the chart. Keep the data table near the chart
// when users need exact values; charts are summary views, not the only data UI.
type Config struct {
	Label            string
	Caption          string
	Title            Title
	Labels           []string
	Series           []Series
	Area             AreaOptions
	XAxis            CategoryAxisOptions
	Legend           LegendOptions
	Padding          Padding
	Symbol           Symbol
	StrokeWidth      float64
	SmoothingTension float64
	Stacked          bool
	Annotations      []TextAnnotation
	// YAxes optionally configures one or two numeric axes. Its zero value keeps
	// the original single-axis Line rendering byte-for-byte.
	YAxes  []Axis
	Width  int
	Height int
	Style  charttheme.Style
	// Controls configures shared controls; Expand defaults on while fullscreen defaults off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("line chart label is required")
	}
	if len(cfg.Labels) == 0 {
		return fmt.Errorf("line chart needs at least one label")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("line chart needs at least one series")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("line chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("line chart height cannot be negative")
	}
	if !finite(cfg.Area.Opacity) || cfg.Area.Opacity < 0 || cfg.Area.Opacity > 1 {
		return fmt.Errorf("line chart area opacity must be finite and between 0 and 1")
	}
	if !cfg.Area.Enabled && cfg.Area.Opacity != 0 {
		return fmt.Errorf("line chart area opacity requires area fill to be enabled")
	}
	if err := validatePadding("legend", cfg.Legend.Padding); err != nil {
		return err
	}
	if err := validatePadding("plot", cfg.Padding); err != nil {
		return err
	}
	if err := validateSymbol("default", cfg.Symbol); err != nil {
		return err
	}
	if !finite(cfg.StrokeWidth) || cfg.StrokeWidth < 0 {
		return fmt.Errorf("line chart stroke width must be finite and non-negative")
	}
	if !finite(cfg.SmoothingTension) || cfg.SmoothingTension < 0 || cfg.SmoothingTension > 1 {
		return fmt.Errorf("line chart smoothing tension must be finite and between 0 and 1")
	}
	if cfg.Stacked && len(cfg.YAxes) > 1 {
		return fmt.Errorf("line chart stacked treatment supports one Y axis")
	}
	if !finite(cfg.Title.FontSize) || cfg.Title.FontSize < 0 {
		return fmt.Errorf("line chart title font size must be finite and non-negative")
	}
	if cfg.Title.Placement != PlacementDefault && cfg.Title.Placement != PlacementCenter && cfg.Title.Placement != PlacementRight {
		return fmt.Errorf("line chart title placement %q is unsupported", cfg.Title.Placement)
	}
	if err := validateCategoryAxis(cfg.XAxis); err != nil {
		return err
	}
	if len(cfg.YAxes) > 2 {
		return fmt.Errorf("line chart supports at most two Y axes")
	}
	axisCount := len(cfg.YAxes)
	if axisCount == 0 {
		axisCount = 1
	}
	labels := make(map[string]struct{}, len(cfg.Labels))
	for index, label := range cfg.Labels {
		label = strings.TrimSpace(label)
		if label == "" {
			return fmt.Errorf("line chart label %d cannot be empty", index+1)
		}
		if _, exists := labels[label]; exists {
			return fmt.Errorf("line chart label %q is duplicated", label)
		}
		labels[label] = struct{}{}
	}
	seriesNames := make(map[string]struct{}, len(cfg.Series))
	usedAxes := make([]bool, axisCount)
	for index, axis := range cfg.YAxes {
		if !finite(axis.Unit) || axis.Unit < 0 {
			return fmt.Errorf("line chart Y axis %d unit must be finite and non-negative", index)
		}
		if axis.Min != nil && !finite(*axis.Min) {
			return fmt.Errorf("line chart Y axis %d minimum must be finite", index)
		}
		if axis.Max != nil && !finite(*axis.Max) {
			return fmt.Errorf("line chart Y axis %d maximum must be finite", index)
		}
		if axis.Min != nil && axis.Max != nil && *axis.Min >= *axis.Max {
			return fmt.Errorf("line chart Y axis %d minimum must be less than maximum", index)
		}
		if axis.LabelCount < 0 {
			return fmt.Errorf("line chart Y axis %d label count cannot be negative", index)
		}
		if !finite(axis.LabelFontSize) || axis.LabelFontSize < 0 {
			return fmt.Errorf("line chart Y axis %d label font size must be finite and non-negative", index)
		}
		if !finite(axis.TitleFontSize) || axis.TitleFontSize < 0 {
			return fmt.Errorf("line chart Y axis %d title font size must be finite and non-negative", index)
		}
		if strings.TrimSpace(axis.Color) != "" && strings.TrimSpace(axis.Class) != "" {
			return fmt.Errorf("line chart Y axis %d cannot set both color and class", index)
		}
		if unsafeCSS(axis.Color) {
			return fmt.Errorf("line chart Y axis %d color is unsafe", index)
		}
		if unsafeClass(axis.Class) {
			return fmt.Errorf("line chart Y axis %d class is unsafe", index)
		}
	}
	for index, series := range cfg.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("line chart series %d needs a name", index+1)
		}
		if _, exists := seriesNames[name]; exists {
			return fmt.Errorf("line chart series name %q is duplicated", name)
		}
		seriesNames[name] = struct{}{}
		if series.Values != nil && series.Points != nil {
			return fmt.Errorf("line chart series %q cannot use both values and points", series.Name)
		}
		if series.Values == nil && series.Points == nil {
			return fmt.Errorf("line chart series %q needs values or points", series.Name)
		}
		if series.Values != nil && len(series.Values) != len(cfg.Labels) {
			return fmt.Errorf("line chart series %q has %d values; need %d", series.Name, len(series.Values), len(cfg.Labels))
		}
		if series.Points != nil && len(series.Points) != len(cfg.Labels) {
			return fmt.Errorf("line chart series %q has %d points; need %d", series.Name, len(series.Points), len(cfg.Labels))
		}
		if series.YAxisIndex < 0 || series.YAxisIndex >= axisCount {
			return fmt.Errorf("line chart series %q Y axis index %d is out of bounds for %d axes", series.Name, series.YAxisIndex, axisCount)
		}
		usedAxes[series.YAxisIndex] = true
		if strings.TrimSpace(series.Color) != "" && strings.TrimSpace(series.Class) != "" {
			return fmt.Errorf("line chart series %q cannot set both color and class", series.Name)
		}
		if unsafeCSS(series.Color) {
			return fmt.Errorf("line chart series %q color is unsafe", series.Name)
		}
		if unsafeClass(series.Class) {
			return fmt.Errorf("line chart series %q class is unsafe", series.Name)
		}
		if err := validateSymbol(fmt.Sprintf("series %q", series.Name), series.Symbol); err != nil {
			return err
		}
		if err := validateLabels(series.Name, series.Labels); err != nil {
			return err
		}
		if err := validateReferences(series.Name, series.References); err != nil {
			return err
		}
		for valueIndex, value := range resolvedSeriesValues(series) {
			if series.Points != nil && series.Points[valueIndex].Missing {
				continue
			}
			if !finite(value) {
				kind := "value"
				if series.Points != nil {
					kind = "point"
				}
				return fmt.Errorf("line chart series %q %s %d must be finite", series.Name, kind, valueIndex)
			}
			if len(cfg.YAxes) > 0 {
				axis := cfg.YAxes[series.YAxisIndex]
				if axis.Min != nil && value < *axis.Min {
					return fmt.Errorf("line chart series %q value %d is below Y axis %d minimum", series.Name, valueIndex, series.YAxisIndex)
				}
				if axis.Max != nil && value > *axis.Max {
					return fmt.Errorf("line chart series %q value %d is above Y axis %d maximum", series.Name, valueIndex, series.YAxisIndex)
				}
			}
		}
	}
	for index := range cfg.YAxes {
		if !usedAxes[index] {
			return fmt.Errorf("line chart Y axis %d has no assigned series", index)
		}
	}
	for index, annotation := range cfg.Annotations {
		if strings.TrimSpace(annotation.Text) == "" {
			return fmt.Errorf("line chart annotation %d text is required", index+1)
		}
		if annotation.X < 0 || annotation.Y < 0 {
			return fmt.Errorf("line chart annotation %d coordinates cannot be negative", index+1)
		}
		if annotation.SeriesIndex < 0 || annotation.SeriesIndex >= len(cfg.Series) {
			return fmt.Errorf("line chart annotation %d series index is out of bounds", index+1)
		}
		if !finite(annotation.FontSize) || annotation.FontSize < 0 {
			return fmt.Errorf("line chart annotation %d font size must be finite and non-negative", index+1)
		}
		if strings.TrimSpace(annotation.Color) != "" && strings.TrimSpace(annotation.Class) != "" {
			return fmt.Errorf("line chart annotation %d cannot set both color and class", index+1)
		}
		if unsafeCSS(annotation.Color) || unsafeClass(annotation.Class) {
			return fmt.Errorf("line chart annotation %d presentation is unsafe", index+1)
		}
	}
	return nil
}

func validateSymbol(name string, symbol Symbol) error {
	switch symbol.Shape {
	case SymbolDefault, SymbolNone, SymbolCircle, SymbolDot, SymbolSquare, SymbolDiamond:
	default:
		return fmt.Errorf("line chart %s symbol shape %q is unsupported", name, symbol.Shape)
	}
	if !finite(symbol.Size) || symbol.Size < 0 {
		return fmt.Errorf("line chart %s symbol size must be finite and non-negative", name)
	}
	return nil
}

func validateCategoryAxis(axis CategoryAxisOptions) error {
	if !finite(axis.Unit) || axis.Unit < 0 {
		return fmt.Errorf("line chart X axis unit must be finite and non-negative")
	}
	if axis.LabelCount < 0 {
		return fmt.Errorf("line chart X axis label count cannot be negative")
	}
	if !finite(axis.LabelFontSize) || axis.LabelFontSize < 0 {
		return fmt.Errorf("line chart X axis label font size must be finite and non-negative")
	}
	if !finite(axis.LabelRotation) || axis.LabelRotation < -360 || axis.LabelRotation > 360 {
		return fmt.Errorf("line chart X axis label rotation must be between -360 and 360 degrees")
	}
	return nil
}

func validateLabels(series string, labels DataLabelOptions) error {
	if labels.Format != ValueFormatDefault && labels.Format != ValueFormatHumanized {
		return fmt.Errorf("line chart series %q label format %q is unsupported", series, labels.Format)
	}
	if labels.Decimals < 0 || labels.Decimals > 12 {
		return fmt.Errorf("line chart series %q label decimals must be between 0 and 12", series)
	}
	if labels.ColorScale != LabelColorScaleNone && labels.ColorScale != LabelColorScaleColdToWarm {
		return fmt.Errorf("line chart series %q label color scale %q is unsupported", series, labels.ColorScale)
	}
	if !finite(labels.FontSize) || labels.FontSize < 0 {
		return fmt.Errorf("line chart series %q label font size must be finite and non-negative", series)
	}
	return nil
}

func validateReferences(series string, references References) error {
	if references.Format != ValueFormatDefault && references.Format != ValueFormatHumanized {
		return fmt.Errorf("line chart series %q reference format %q is unsupported", series, references.Format)
	}
	if references.Decimals < 0 || references.Decimals > 12 {
		return fmt.Errorf("line chart series %q reference decimals must be between 0 and 12", series)
	}
	if references.PointSize < 0 {
		return fmt.Errorf("line chart series %q reference point size cannot be negative", series)
	}
	if unsafeClass(references.Style.Class) {
		return fmt.Errorf("line chart series %q reference presentation is unsafe", series)
	}
	return nil
}

func resolvedSeriesValues(series Series) []float64 {
	if series.Points == nil {
		return series.Values
	}
	values := make([]float64, len(series.Points))
	for index, point := range series.Points {
		values[index] = point.Value
	}
	return values
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func unsafeCSS(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ContainsAny(value, ";{}<>\\\"") || strings.Contains(value, "url(") || strings.Contains(value, "expression(")
}

func unsafeClass(value string) bool { return strings.ContainsAny(value, "\"'<>;") }

func validatePadding(name string, padding Padding) error {
	if padding.Top < 0 || padding.Right < 0 || padding.Bottom < 0 || padding.Left < 0 {
		return fmt.Errorf("line chart %s padding cannot be negative", name)
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
