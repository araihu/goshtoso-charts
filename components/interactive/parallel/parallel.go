// Package parallel provides the canonical interactive parallel-coordinates API.
//
// Parallel-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package parallel

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const maxParallelDetailRows = 64

// Instance is the renderer-neutral chart instance returned by Parallel.
type Instance = chart.Instance

// Scale controls a numeric dimension's scale.
type Scale string

const (
	// ScaleLinear uses a continuous linear scale. It is the default.
	ScaleLinear Scale = ""
	// ScaleLog uses a positive logarithmic scale.
	ScaleLog Scale = "log"
)

// NameLocation controls where a dimension name sits along its axis.
type NameLocation string

const (
	NameEnd    NameLocation = ""
	NameStart  NameLocation = "start"
	NameMiddle NameLocation = "middle"
)

// Range bounds a numeric dimension.
type Range struct {
	Min *float64
	Max *float64
}

// AxisLabel configures labels beside one parallel axis.
type AxisLabel struct {
	Show     *bool
	Rotate   float64
	Margin   float64
	FontSize int
	Color    string
}

// AxisLine configures one parallel axis line.
type AxisLine struct {
	Show  *bool
	Color string
	Width float64
	Type  string
}

// Dimension describes one ordered numeric or categorical dimension.
// Categories selects a categorical axis; otherwise the dimension is numeric.
type Dimension struct {
	Name         string
	Range        *Range
	Categories   []string
	Scale        Scale
	Inverse      bool
	NameLocation NameLocation
	Label        AxisLabel
	Line         AxisLine
}

// Layout sets percentage insets around the parallel coordinate system.
// Nil values retain renderer defaults.
type Layout struct {
	LeftPercent   *float64
	RightPercent  *float64
	TopPercent    *float64
	BottomPercent *float64
}

// LineOptions configures observation paths.
type LineOptions struct {
	Width   float64
	Type    string
	Opacity *float64
}

// SeriesOptions configures a complete parallel series.
type SeriesOptions struct {
	Line            LineOptions
	Smooth          *bool
	InactiveOpacity *float64
}

type parallelValueKind uint8

const (
	parallelValueInvalid parallelValueKind = iota
	parallelValueNumber
	parallelValueCategory
)

// Value is one typed numeric or categorical dimension value.
// Construct values with Number or Category.
type Value struct {
	kind     parallelValueKind
	number   float64
	category string
}

// Number constructs a numeric dimension value.
func Number(value float64) Value {
	return Value{kind: parallelValueNumber, number: value}
}

// Category constructs a categorical dimension value.
func Category(value string) Value {
	return Value{kind: parallelValueCategory, category: value}
}

// Observation describes one named row aligned with Dimensions.
// Class provides color-independent semantics. Color optionally overrides its path.
type Observation struct {
	Name   string
	Values []Value
	Class  string
	Color  string
}

// Series describes one named collection of aligned observations.
// Class provides color-independent semantics. Color optionally overrides all paths.
type Series struct {
	Name         string
	Observations []Observation
	Class        string
	Color        string
	Options      SeriesOptions
}

// Config describes an accessible parallel-coordinates chart.
type Config struct {
	Label      string
	Caption    string
	Dimensions []Dimension
	Series     []Series
	Layout     Layout
	Width      string
	Height     string
	Options    chart.ChartOptions
	Style      charttheme.Style
	RootAttrs  templ.Attributes
}

// Parallel builds a reusable interactive parallel-coordinates component.
func Parallel(cfg Config) Instance {
	if err := validateParallelConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveParallel, err)
	}

	chart := charts.NewParallel()
	privateAxes := make([]opts.ParallelAxis, len(cfg.Dimensions))
	renderedAxes := make([]rendererParallelAxis, len(cfg.Dimensions))
	for index, dimension := range cfg.Dimensions {
		privateAxes[index] = opts.ParallelAxis{
			Dim: index, Name: dimension.Name, Inverse: opts.Bool(dimension.Inverse),
			NameLocation: string(dimension.NameLocation),
		}
		renderedAxes[index] = renderParallelAxis(index, dimension)
		if len(dimension.Categories) > 0 {
			privateAxes[index].Type = "category"
			privateAxes[index].Data = dimension.Categories
		} else {
			privateAxes[index].Type = string(dimension.Scale)
			if dimension.Range != nil && dimension.Range.Max != nil {
				privateAxes[index].Max = *dimension.Range.Max
			}
		}
	}

	globalOptions := []charts.GlobalOpts{
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
		charts.WithParallelAxisList(privateAxes),
	}
	if layout := renderParallelLayout(cfg.Layout); layout != (opts.ParallelComponent{}) {
		globalOptions = append(globalOptions, charts.WithParallelComponentOpts(layout))
	}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)

	replacements := make([]internalinteractive.ScriptReplacement, 0, len(cfg.Series)+1)
	privateAxisJSON, _ := json.Marshal(privateAxes)
	renderedAxisJSON, _ := json.Marshal(renderedAxes)
	replacements = append(replacements, internalinteractive.ScriptReplacement{
		Old: `"parallelAxis":` + string(privateAxisJSON),
		New: `"parallelAxis":` + string(renderedAxisJSON),
	})
	for seriesIndex, series := range cfg.Series {
		data := make([]opts.ParallelData, len(series.Observations))
		seriesOptions := parallelSeriesOptions(series)
		chart.AddSeries(series.Name, data, seriesOptions...)
		rendered := make([]rendererParallelObservation, len(series.Observations))
		for observationIndex, observation := range series.Observations {
			rendered[observationIndex] = renderParallelObservation(observation)
		}
		chart.MultiSeries[seriesIndex].Data = rendered
		if series.Class != "" || series.Options.InactiveOpacity != nil {
			quotedName, _ := json.Marshal(series.Name)
			old := `"name":` + string(quotedName) + `,"type":"parallel"`
			addition := ""
			if series.Class != "" {
				quotedClass, _ := json.Marshal(series.Class)
				addition += `,"className":` + string(quotedClass)
			}
			if series.Options.InactiveOpacity != nil {
				addition += `,"inactiveOpacity":` + strconv.FormatFloat(*series.Options.InactiveOpacity, 'f', -1, 64)
			}
			replacements = append(replacements, internalinteractive.ScriptReplacement{Old: old, New: old + addition})
		}
	}

	return internalinteractive.New(chartcomponents.KindInteractiveParallel, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		RootAttrs: cfg.RootAttrs, Details: parallelExactValues(cfg.Dimensions, flattenParallelRows(cfg.Series, maxParallelDetailRows)),
		ScriptReplacements: replacements,
	})
}

type rendererParallelAxis struct {
	Dim          int                        `json:"dim"`
	Name         string                     `json:"name"`
	Min          *float64                   `json:"min,omitempty"`
	Max          *float64                   `json:"max,omitempty"`
	Inverse      bool                       `json:"inverse,omitempty"`
	NameLocation string                     `json:"nameLocation,omitempty"`
	Type         string                     `json:"type,omitempty"`
	Data         []string                   `json:"data,omitempty"`
	AxisLabel    *rendererParallelAxisLabel `json:"axisLabel,omitempty"`
	AxisLine     *rendererParallelAxisLine  `json:"axisLine,omitempty"`
}

type rendererParallelAxisLabel struct {
	Show     *bool   `json:"show,omitempty"`
	Rotate   float64 `json:"rotate,omitempty"`
	Margin   float64 `json:"margin,omitempty"`
	FontSize int     `json:"fontSize,omitempty"`
	Color    string  `json:"color,omitempty"`
}

type rendererParallelAxisLine struct {
	Show      *bool                     `json:"show,omitempty"`
	LineStyle rendererParallelLineStyle `json:"lineStyle"`
}

type rendererParallelLineStyle struct {
	Color   string   `json:"color,omitempty"`
	Width   float64  `json:"width,omitempty"`
	Type    string   `json:"type,omitempty"`
	Opacity *float64 `json:"opacity,omitempty"`
}

type rendererParallelObservation struct {
	Name      string                     `json:"name"`
	Value     []any                      `json:"value"`
	ClassName string                     `json:"className,omitempty"`
	LineStyle *rendererParallelLineStyle `json:"lineStyle,omitempty"`
}

func renderParallelAxis(index int, value Dimension) rendererParallelAxis {
	result := rendererParallelAxis{
		Dim: index, Name: value.Name, Inverse: value.Inverse,
		NameLocation: string(value.NameLocation),
	}
	if len(value.Categories) > 0 {
		result.Type, result.Data = "category", value.Categories
	} else {
		result.Type = string(value.Scale)
		if value.Range != nil {
			result.Min, result.Max = value.Range.Min, value.Range.Max
		}
	}
	if value.Label.Show != nil || value.Label.Rotate != 0 || value.Label.Margin != 0 || value.Label.FontSize != 0 || value.Label.Color != "" {
		result.AxisLabel = &rendererParallelAxisLabel{
			Show: value.Label.Show, Rotate: value.Label.Rotate, Margin: value.Label.Margin,
			FontSize: value.Label.FontSize, Color: value.Label.Color,
		}
	}
	if value.Line.Show != nil || value.Line.Color != "" || value.Line.Width != 0 || value.Line.Type != "" {
		result.AxisLine = &rendererParallelAxisLine{
			Show:      value.Line.Show,
			LineStyle: rendererParallelLineStyle{Color: value.Line.Color, Width: value.Line.Width, Type: value.Line.Type},
		}
	}
	return result
}

func renderParallelLayout(value Layout) opts.ParallelComponent {
	return opts.ParallelComponent{
		Left: percentPosition(value.LeftPercent), Right: percentPosition(value.RightPercent),
		Top: percentPosition(value.TopPercent), Bottom: percentPosition(value.BottomPercent),
	}
}

func percentPosition(value *float64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(*value, 'f', -1, 64) + "%"
}

func parallelSeriesOptions(series Series) []charts.SeriesOpts {
	result := make([]charts.SeriesOpts, 0, 2)
	line := series.Options.Line
	if series.Color != "" || line.Width != 0 || line.Type != "" || line.Opacity != nil {
		style := opts.LineStyle{Color: series.Color, Width: float32(line.Width), Type: line.Type}
		if line.Opacity != nil {
			style.Opacity = opts.Float(float32(*line.Opacity))
		}
		result = append(result, charts.WithLineStyleOpts(style))
	}
	if series.Options.Smooth != nil {
		result = append(result, charts.WithSeriesOpts(func(value *charts.SingleSeries) {
			value.Smooth = opts.Bool(*series.Options.Smooth)
		}))
	}
	return result
}

func renderParallelObservation(value Observation) rendererParallelObservation {
	result := rendererParallelObservation{Name: value.Name, ClassName: value.Class, Value: make([]any, len(value.Values))}
	for index, item := range value.Values {
		if item.kind == parallelValueCategory {
			result.Value[index] = item.category
		} else {
			result.Value[index] = item.number
		}
	}
	if value.Color != "" {
		result.LineStyle = &rendererParallelLineStyle{Color: value.Color}
	}
	return result
}

func validateParallelConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("parallel chart label is required")
	}
	if len(cfg.Dimensions) < 2 {
		return fmt.Errorf("parallel chart requires at least two dimensions")
	}
	dimensionNames := make(map[string]bool, len(cfg.Dimensions))
	for index, dimension := range cfg.Dimensions {
		name := strings.TrimSpace(dimension.Name)
		if name == "" {
			return fmt.Errorf("parallel chart dimension %d name is required", index)
		}
		if dimensionNames[name] {
			return fmt.Errorf("parallel chart dimension name %q is duplicated", name)
		}
		dimensionNames[name] = true
		if err := validateParallelDimension(dimension); err != nil {
			return fmt.Errorf("parallel chart dimension %q: %w", name, err)
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("parallel chart series is required")
	}
	seriesNames := make(map[string]bool, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("parallel chart series %d name is required", seriesIndex)
		}
		if seriesNames[name] {
			return fmt.Errorf("parallel chart series name %q is duplicated", name)
		}
		seriesNames[name] = true
		if len(series.Observations) == 0 {
			return fmt.Errorf("parallel chart series %q observations are required", name)
		}
		if err := validateParallelSeriesOptions(name, series.Options); err != nil {
			return err
		}
		observationNames := make(map[string]bool, len(series.Observations))
		for observationIndex, observation := range series.Observations {
			observationName := strings.TrimSpace(observation.Name)
			if observationName == "" {
				return fmt.Errorf("parallel chart series %q observation %d name is required", name, observationIndex)
			}
			if observationNames[observationName] {
				return fmt.Errorf("parallel chart series %q observation name %q is duplicated", name, observationName)
			}
			observationNames[observationName] = true
			if len(observation.Values) != len(cfg.Dimensions) {
				return fmt.Errorf("parallel chart series %q observation %q has %d values for %d dimensions", name, observationName, len(observation.Values), len(cfg.Dimensions))
			}
			for valueIndex, value := range observation.Values {
				if err := validateParallelValue(cfg.Dimensions[valueIndex], value); err != nil {
					return fmt.Errorf("parallel chart series %q observation %q dimension %q: %w", name, observationName, cfg.Dimensions[valueIndex].Name, err)
				}
			}
		}
	}
	if err := validateParallelLayout(cfg.Layout); err != nil {
		return err
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("parallel chart root attribute %q is reserved", attribute)
			}
		}
	}
	return internalinteractive.ValidateChartOptions(cfg.Options)
}

func validateParallelDimension(value Dimension) error {
	if value.Scale != ScaleLinear && value.Scale != ScaleLog {
		return fmt.Errorf("scale %q is not supported", value.Scale)
	}
	if value.NameLocation != NameEnd && value.NameLocation != NameStart && value.NameLocation != NameMiddle {
		return fmt.Errorf("name location %q is not supported", value.NameLocation)
	}
	if len(value.Categories) > 0 {
		if value.Range != nil {
			return fmt.Errorf("categorical dimensions cannot define a numeric range")
		}
		if value.Scale == ScaleLog {
			return fmt.Errorf("categorical dimensions cannot use a log scale")
		}
		seen := make(map[string]bool, len(value.Categories))
		for index, category := range value.Categories {
			category = strings.TrimSpace(category)
			if category == "" {
				return fmt.Errorf("category %d is empty", index)
			}
			if seen[category] {
				return fmt.Errorf("category %q is duplicated", category)
			}
			seen[category] = true
		}
	}
	if value.Range != nil {
		if value.Range.Min == nil && value.Range.Max == nil {
			return fmt.Errorf("numeric range must set min or max")
		}
		if value.Range.Min != nil && !internalinteractive.FiniteNumber(*value.Range.Min) || value.Range.Max != nil && !internalinteractive.FiniteNumber(*value.Range.Max) {
			return fmt.Errorf("numeric range bounds must be finite")
		}
		if value.Range.Min != nil && value.Range.Max != nil && *value.Range.Min >= *value.Range.Max {
			return fmt.Errorf("numeric range min must be less than max")
		}
		if value.Scale == ScaleLog && value.Range.Min != nil && *value.Range.Min <= 0 {
			return fmt.Errorf("log range minimum must be positive")
		}
	}
	if !internalinteractive.FiniteNumber(value.Label.Rotate) || value.Label.Rotate < -90 || value.Label.Rotate > 90 {
		return fmt.Errorf("label rotation must be between -90 and 90")
	}
	if !internalinteractive.FiniteNumber(value.Label.Margin) || value.Label.Margin < 0 {
		return fmt.Errorf("label margin must be finite and nonnegative")
	}
	if value.Label.FontSize < 0 {
		return fmt.Errorf("label font size must be nonnegative")
	}
	if !internalinteractive.FiniteNumber(value.Line.Width) || value.Line.Width < 0 {
		return fmt.Errorf("axis line width must be finite and nonnegative")
	}
	if value.Line.Type != "" && value.Line.Type != "solid" && value.Line.Type != "dashed" && value.Line.Type != "dotted" {
		return fmt.Errorf("axis line type %q is not supported", value.Line.Type)
	}
	return nil
}

func validateParallelValue(dimension Dimension, value Value) error {
	if len(dimension.Categories) > 0 {
		if value.kind != parallelValueCategory {
			return fmt.Errorf("value must be categorical")
		}
		for _, category := range dimension.Categories {
			if value.category == category {
				return nil
			}
		}
		return fmt.Errorf("category %q is not declared", value.category)
	}
	if value.kind != parallelValueNumber {
		return fmt.Errorf("value must be numeric")
	}
	if !internalinteractive.FiniteNumber(value.number) {
		return fmt.Errorf("numeric value must be finite")
	}
	if dimension.Scale == ScaleLog && value.number <= 0 {
		return fmt.Errorf("log value must be positive")
	}
	if dimension.Range != nil && dimension.Range.Min != nil && value.number < *dimension.Range.Min {
		return fmt.Errorf("value %v is below minimum %v", value.number, *dimension.Range.Min)
	}
	if dimension.Range != nil && dimension.Range.Max != nil && value.number > *dimension.Range.Max {
		return fmt.Errorf("value %v exceeds maximum %v", value.number, *dimension.Range.Max)
	}
	return nil
}

func validateParallelSeriesOptions(name string, value SeriesOptions) error {
	if !internalinteractive.FiniteNumber(value.Line.Width) || value.Line.Width < 0 {
		return fmt.Errorf("parallel chart series %q line width must be finite and nonnegative", name)
	}
	if value.Line.Type != "" && value.Line.Type != "solid" && value.Line.Type != "dashed" && value.Line.Type != "dotted" {
		return fmt.Errorf("parallel chart series %q line type %q is not supported", name, value.Line.Type)
	}
	if value.Line.Opacity != nil && (!internalinteractive.FiniteNumber(*value.Line.Opacity) || *value.Line.Opacity < 0 || *value.Line.Opacity > 1) {
		return fmt.Errorf("parallel chart series %q line opacity must be between 0 and 1", name)
	}
	if value.InactiveOpacity != nil && (!internalinteractive.FiniteNumber(*value.InactiveOpacity) || *value.InactiveOpacity < 0 || *value.InactiveOpacity > 1) {
		return fmt.Errorf("parallel chart series %q inactive opacity must be between 0 and 1", name)
	}
	return nil
}

func validateParallelLayout(value Layout) error {
	for name, inset := range map[string]*float64{
		"left": value.LeftPercent, "right": value.RightPercent,
		"top": value.TopPercent, "bottom": value.BottomPercent,
	} {
		if inset != nil && (!internalinteractive.FiniteNumber(*inset) || *inset < 0 || *inset > 100) {
			return fmt.Errorf("parallel chart %s inset must be between 0 and 100 percent", name)
		}
	}
	if value.LeftPercent != nil && value.RightPercent != nil && *value.LeftPercent+*value.RightPercent >= 100 {
		return fmt.Errorf("parallel chart horizontal insets must total less than 100 percent")
	}
	if value.TopPercent != nil && value.BottomPercent != nil && *value.TopPercent+*value.BottomPercent >= 100 {
		return fmt.Errorf("parallel chart vertical insets must total less than 100 percent")
	}
	return nil
}

type parallelValueRow struct {
	Series      string
	Observation string
	Values      []string
	Class       string
}

type parallelValueRows struct {
	Rows    []parallelValueRow
	Omitted int
}

func flattenParallelRows(series []Series, limit int) parallelValueRows {
	result := parallelValueRows{Rows: make([]parallelValueRow, 0, limit)}
	for _, group := range series {
		for _, observation := range group.Observations {
			if len(result.Rows) >= limit {
				result.Omitted++
				continue
			}
			values := make([]string, len(observation.Values))
			for index, value := range observation.Values {
				if value.kind == parallelValueCategory {
					values[index] = value.category
				} else {
					values[index] = strconv.FormatFloat(value.number, 'f', -1, 64)
				}
			}
			class := strings.TrimSpace(strings.Join([]string{group.Class, observation.Class}, " "))
			result.Rows = append(result.Rows, parallelValueRow{Series: group.Name, Observation: observation.Name, Values: values, Class: class})
		}
	}
	return result
}
