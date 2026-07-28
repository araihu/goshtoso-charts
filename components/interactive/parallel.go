package interactive

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const maxParallelDetailRows = 64

// ParallelScale controls a numeric dimension's scale.
type ParallelScale string

const (
	// ParallelScaleLinear uses a continuous linear scale. It is the default.
	ParallelScaleLinear ParallelScale = ""
	// ParallelScaleLog uses a positive logarithmic scale.
	ParallelScaleLog ParallelScale = "log"
)

// ParallelNameLocation controls where a dimension name sits along its axis.
type ParallelNameLocation string

const (
	ParallelNameEnd    ParallelNameLocation = ""
	ParallelNameStart  ParallelNameLocation = "start"
	ParallelNameMiddle ParallelNameLocation = "middle"
)

// ParallelRange bounds a numeric dimension.
type ParallelRange struct {
	Min *float64
	Max *float64
}

// ParallelAxisLabel configures labels beside one parallel axis.
type ParallelAxisLabel struct {
	Show     *bool
	Rotate   float64
	Margin   float64
	FontSize int
	Color    string
}

// ParallelAxisLine configures one parallel axis line.
type ParallelAxisLine struct {
	Show  *bool
	Color string
	Width float64
	Type  string
}

// ParallelDimension describes one ordered numeric or categorical dimension.
// Categories selects a categorical axis; otherwise the dimension is numeric.
type ParallelDimension struct {
	Name         string
	Range        *ParallelRange
	Categories   []string
	Scale        ParallelScale
	Inverse      bool
	NameLocation ParallelNameLocation
	Label        ParallelAxisLabel
	Line         ParallelAxisLine
}

// ParallelLayout sets percentage insets around the parallel coordinate system.
// Nil values retain renderer defaults.
type ParallelLayout struct {
	LeftPercent   *float64
	RightPercent  *float64
	TopPercent    *float64
	BottomPercent *float64
}

// ParallelLineOptions configures observation paths.
type ParallelLineOptions struct {
	Width   float64
	Type    string
	Opacity *float64
}

// ParallelSeriesOptions configures a complete parallel series.
type ParallelSeriesOptions struct {
	Line            ParallelLineOptions
	Smooth          *bool
	InactiveOpacity *float64
}

type parallelValueKind uint8

const (
	parallelValueInvalid parallelValueKind = iota
	parallelValueNumber
	parallelValueCategory
)

// ParallelValue is one typed numeric or categorical dimension value.
// Construct values with ParallelNumber or ParallelCategory.
type ParallelValue struct {
	kind     parallelValueKind
	number   float64
	category string
}

// ParallelNumber constructs a numeric dimension value.
func ParallelNumber(value float64) ParallelValue {
	return ParallelValue{kind: parallelValueNumber, number: value}
}

// ParallelCategory constructs a categorical dimension value.
func ParallelCategory(value string) ParallelValue {
	return ParallelValue{kind: parallelValueCategory, category: value}
}

// ParallelObservation describes one named row aligned with Dimensions.
// Class provides color-independent semantics. Color optionally overrides its path.
type ParallelObservation struct {
	Name   string
	Values []ParallelValue
	Class  string
	Color  string
}

// ParallelSeries describes one named collection of aligned observations.
// Class provides color-independent semantics. Color optionally overrides all paths.
type ParallelSeries struct {
	Name         string
	Observations []ParallelObservation
	Class        string
	Color        string
	Options      ParallelSeriesOptions
}

// ParallelConfig describes an accessible parallel-coordinates chart.
type ParallelConfig struct {
	Label      string
	Caption    string
	Dimensions []ParallelDimension
	Series     []ParallelSeries
	Layout     ParallelLayout
	Width      string
	Height     string
	Options    ChartOptions
	Style      charttheme.Style
	RootAttrs  templ.Attributes
}

// Parallel builds a reusable interactive parallel-coordinates component.
func Parallel(cfg ParallelConfig) Instance {
	if err := validateParallelConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveParallel, err)
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
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)

	replacements := make([]scriptReplacement, 0, len(cfg.Series)+1)
	privateAxisJSON, _ := json.Marshal(privateAxes)
	renderedAxisJSON, _ := json.Marshal(renderedAxes)
	replacements = append(replacements, scriptReplacement{
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
			replacements = append(replacements, scriptReplacement{Old: old, New: old + addition})
		}
	}

	return newInstance(chartcomponents.KindInteractiveParallel, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style,
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

func renderParallelAxis(index int, value ParallelDimension) rendererParallelAxis {
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

func renderParallelLayout(value ParallelLayout) opts.ParallelComponent {
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

func parallelSeriesOptions(series ParallelSeries) []charts.SeriesOpts {
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

func renderParallelObservation(value ParallelObservation) rendererParallelObservation {
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

func validateParallelConfig(cfg ParallelConfig) error {
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
	return validateChartOptions(cfg.Options)
}

func validateParallelDimension(value ParallelDimension) error {
	if value.Scale != ParallelScaleLinear && value.Scale != ParallelScaleLog {
		return fmt.Errorf("scale %q is not supported", value.Scale)
	}
	if value.NameLocation != ParallelNameEnd && value.NameLocation != ParallelNameStart && value.NameLocation != ParallelNameMiddle {
		return fmt.Errorf("name location %q is not supported", value.NameLocation)
	}
	if len(value.Categories) > 0 {
		if value.Range != nil {
			return fmt.Errorf("categorical dimensions cannot define a numeric range")
		}
		if value.Scale == ParallelScaleLog {
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
		if value.Range.Min != nil && !finiteNumber(*value.Range.Min) || value.Range.Max != nil && !finiteNumber(*value.Range.Max) {
			return fmt.Errorf("numeric range bounds must be finite")
		}
		if value.Range.Min != nil && value.Range.Max != nil && *value.Range.Min >= *value.Range.Max {
			return fmt.Errorf("numeric range min must be less than max")
		}
		if value.Scale == ParallelScaleLog && value.Range.Min != nil && *value.Range.Min <= 0 {
			return fmt.Errorf("log range minimum must be positive")
		}
	}
	if !finiteNumber(value.Label.Rotate) || value.Label.Rotate < -90 || value.Label.Rotate > 90 {
		return fmt.Errorf("label rotation must be between -90 and 90")
	}
	if !finiteNumber(value.Label.Margin) || value.Label.Margin < 0 {
		return fmt.Errorf("label margin must be finite and nonnegative")
	}
	if value.Label.FontSize < 0 {
		return fmt.Errorf("label font size must be nonnegative")
	}
	if !finiteNumber(value.Line.Width) || value.Line.Width < 0 {
		return fmt.Errorf("axis line width must be finite and nonnegative")
	}
	if value.Line.Type != "" && value.Line.Type != "solid" && value.Line.Type != "dashed" && value.Line.Type != "dotted" {
		return fmt.Errorf("axis line type %q is not supported", value.Line.Type)
	}
	return nil
}

func validateParallelValue(dimension ParallelDimension, value ParallelValue) error {
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
	if !finiteNumber(value.number) {
		return fmt.Errorf("numeric value must be finite")
	}
	if dimension.Scale == ParallelScaleLog && value.number <= 0 {
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

func validateParallelSeriesOptions(name string, value ParallelSeriesOptions) error {
	if !finiteNumber(value.Line.Width) || value.Line.Width < 0 {
		return fmt.Errorf("parallel chart series %q line width must be finite and nonnegative", name)
	}
	if value.Line.Type != "" && value.Line.Type != "solid" && value.Line.Type != "dashed" && value.Line.Type != "dotted" {
		return fmt.Errorf("parallel chart series %q line type %q is not supported", name, value.Line.Type)
	}
	if value.Line.Opacity != nil && (!finiteNumber(*value.Line.Opacity) || *value.Line.Opacity < 0 || *value.Line.Opacity > 1) {
		return fmt.Errorf("parallel chart series %q line opacity must be between 0 and 1", name)
	}
	if value.InactiveOpacity != nil && (!finiteNumber(*value.InactiveOpacity) || *value.InactiveOpacity < 0 || *value.InactiveOpacity > 1) {
		return fmt.Errorf("parallel chart series %q inactive opacity must be between 0 and 1", name)
	}
	return nil
}

func validateParallelLayout(value ParallelLayout) error {
	for name, inset := range map[string]*float64{
		"left": value.LeftPercent, "right": value.RightPercent,
		"top": value.TopPercent, "bottom": value.BottomPercent,
	} {
		if inset != nil && (!finiteNumber(*inset) || *inset < 0 || *inset > 100) {
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

func flattenParallelRows(series []ParallelSeries, limit int) parallelValueRows {
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
