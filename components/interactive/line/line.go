package line

import (
	"fmt"
	"strconv"
	"time"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by Line.
type Instance = chart.Instance

// Config describes an accessible, browser-rendered line chart.
//
// Values must be application-owned because the browser renderer serializes them.
type Config struct {
	Label   string
	Caption string
	XAxis   []string
	// TimeAxis selects a temporal x axis. It is mutually exclusive with XAxis.
	// LiveData remains categorical because CartesianSnapshot carries categories.
	TimeAxis *TimeAxis
	// ValueAxis selects ordered numerical x coordinates. It is mutually
	// exclusive with XAxis and TimeAxis.
	ValueAxis     *ValueAxis
	Series        []Series
	Width         string
	Height        string
	Options       chart.ChartOptions
	SeriesOptions chart.SeriesOptions
	Style         charttheme.Style
	Live          *chart.LiveData
	// VisualScale maps x or y values through a theme-aware piecewise scale.
	VisualScale *VisualScale
}

// TimeAxis defines ordered instants and a required inclusive lower bound.
// Values and Minimum use time.Time so callers never supply renderer values.
type TimeAxis struct {
	Values  []time.Time
	Minimum time.Time
	// SplitNumber recommends readable temporal tick density. Zero uses the
	// responsive default of four segments; the private renderer also hides
	// any remaining overlap while retaining endpoint labels.
	SplitNumber int
}

// ValueAxis defines ordered finite numerical x coordinates.
type ValueAxis struct {
	Values []float64
}

// VisualDimension selects the coordinate used by a piecewise visual scale.
type VisualDimension string

const (
	// VisualDimensionX maps x coordinates.
	VisualDimensionX VisualDimension = "x"
	// VisualDimensionY maps y values.
	VisualDimensionY VisualDimension = "y"
)

// VisualScale defines theme-aware numeric pieces. Bounds are optional and
// exclusive, matching a greater-than and less-than interval.
type VisualScale struct {
	Dimension VisualDimension
	Pieces    []VisualPiece
}

// VisualPiece defines one open numeric interval.
type VisualPiece struct {
	GreaterThan *float64
	LessThan    *float64
}

// Statistic selects a calculated series reference.
type Statistic string

const (
	// StatisticMinimum selects the smallest series value.
	StatisticMinimum Statistic = "minimum"
	// StatisticMaximum selects the largest series value.
	StatisticMaximum Statistic = "maximum"
	// StatisticAverage selects the arithmetic mean of series values.
	StatisticAverage Statistic = "average"
)

// Coordinate identifies one finite x/y point. X is a numeric coordinate
// or a zero-based category index.
type Coordinate struct {
	X float64
	Y float64
}

// PointReference places a calculated point on a series.
type PointReference struct {
	Name      string
	Statistic Statistic
}

// GuideReference draws a calculated, vertical, or two-coordinate guide.
// X is a numeric coordinate or a zero-based category index.
// Exactly one of Statistic, X, or Start and End must be configured.
type GuideReference struct {
	Name      string
	Statistic Statistic
	X         *float64
	Start     *Coordinate
	End       *Coordinate
}

// RangeReference highlights an inclusive numeric x-axis interval or an
// inclusive interval between zero-based category indexes.
type RangeReference struct {
	Name   string
	StartX float64
	EndX   float64
}

// References configures theme-aware points, guides, and ranges.
type References struct {
	Points      []PointReference
	Lines       []GuideReference
	Areas       []RangeReference
	ShowLabels  *bool
	StartSymbol string
	EndSymbol   string
	SymbolSize  int
}

// Series describes one named line series.
type Series struct {
	Name       string
	Data       []Data
	Options    chart.SeriesOptions
	References References
}

// Data describes one finite line value and optional point symbol.
type Data struct {
	Name       string
	Value      float64
	Symbol     string
	SymbolSize int
}

// Line builds a reusable interactive line component.
func Line(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveLine, err)
	}

	chart := charts.NewLine()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{
				Width:  cfg.Width,
				Height: cfg.Height,
			}),
		}, globalOptions...)
	}
	if cfg.TimeAxis != nil {
		globalOptions = append(globalOptions, charts.WithXAxisOpts(opts.XAxis{
			Type:        "time",
			Min:         lineTimeValue(cfg.TimeAxis.Minimum),
			SplitNumber: lineTimeSplitNumber(*cfg.TimeAxis),
			AxisLabel: &opts.AxisLabel{
				HideOverlap:   opts.Bool(true),
				ShowMinLabel:  opts.Bool(true),
				ShowMaxLabel:  opts.Bool(true),
				AlignMinLabel: "left",
				AlignMaxLabel: "right",
			},
		}), charts.WithGridOpts(opts.Grid{Top: "18%", Bottom: "15%", ContainLabel: opts.Bool(true)}))
	} else if cfg.ValueAxis != nil {
		globalOptions = append(globalOptions, charts.WithXAxisOpts(opts.XAxis{Type: "value"}))
	}
	visualScaleReplacements := []internalinteractive.ScriptReplacement(nil)
	if cfg.VisualScale != nil {
		visualScale, replacements := rendererVisualScale(*cfg.VisualScale)
		globalOptions = append(globalOptions, charts.WithVisualMapOpts(visualScale))
		visualScaleReplacements = replacements
	}
	chart.SetGlobalOptions(globalOptions...)
	if cfg.TimeAxis == nil && cfg.ValueAxis == nil {
		chart.SetXAxis(cfg.XAxis)
	}
	for _, series := range cfg.Series {
		data := make([]opts.LineData, len(series.Data))
		for index, point := range series.Data {
			value := any(point.Value)
			if cfg.TimeAxis != nil {
				value = []any{lineTimeValue(cfg.TimeAxis.Values[index]), point.Value}
			} else if cfg.ValueAxis != nil {
				value = []any{cfg.ValueAxis.Values[index], point.Value}
			}
			data[index] = opts.LineData{Name: point.Name, Value: value, Symbol: point.Symbol, SymbolSize: point.SymbolSize}
		}
		seriesOptions := internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)
		seriesOptions = append(seriesOptions, rendererReferences(series.References)...)
		chart.AddSeries(series.Name, data, seriesOptions...)
	}

	return internalinteractive.New(chartcomponents.KindInteractiveLine, internalinteractive.RenderConfig{
		Label:              cfg.Label,
		Caption:            cfg.Caption,
		Chart:              chart,
		ResponsiveWidth:    internalinteractive.ResponsiveWidth(cfg.Width),
		Style:              cfg.Style,
		Live:               internalinteractive.CartesianLiveConfig(cfg.Live),
		Details:            lineDetails(cfg),
		Animation:          cfg.Options.Animation,
		Controls:           cfg.Options.Controls,
		Export:             cfg.Options.Export,
		AxisLabelIntervals: internalinteractive.AxisLabelIntervals(cfg.Options),
		ScriptReplacements: visualScaleReplacements,
	})
}

func validateConfig(cfg Config) error {
	if err := internalinteractive.ValidateChartOptions(cfg.Options); err != nil {
		return err
	}
	if err := internalinteractive.ValidateLiveData(cfg.Live); err != nil {
		return err
	}
	axisCount := 0
	if len(cfg.XAxis) != 0 {
		axisCount++
	}
	if cfg.TimeAxis != nil {
		axisCount++
	}
	if cfg.ValueAxis != nil {
		axisCount++
	}
	if axisCount > 1 {
		return fmt.Errorf("line chart category, time, and value axes are mutually exclusive")
	}
	if axisCount == 0 {
		return fmt.Errorf("line chart x axis is required")
	}
	if cfg.TimeAxis != nil {
		if cfg.Live != nil {
			return fmt.Errorf("line chart live data supports categorical x axis only")
		}
		if err := validateTimeAxis(*cfg.TimeAxis); err != nil {
			return err
		}
	}
	if cfg.ValueAxis != nil {
		if cfg.Live != nil {
			return fmt.Errorf("line chart live data supports categorical x axis only")
		}
		if err := validateValueAxis(*cfg.ValueAxis); err != nil {
			return err
		}
	}
	if cfg.VisualScale != nil {
		if err := validateVisualScale(*cfg.VisualScale); err != nil {
			return err
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("line chart series is required")
	}
	for index, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("line chart series %d name is required", index)
		}
		axisLen := len(cfg.XAxis)
		if cfg.TimeAxis != nil {
			axisLen = len(cfg.TimeAxis.Values)
		} else if cfg.ValueAxis != nil {
			axisLen = len(cfg.ValueAxis.Values)
		}
		if len(series.Data) != axisLen {
			return fmt.Errorf("line chart series %q has %d data points for %d x-axis values", series.Name, len(series.Data), axisLen)
		}
		for dataIndex, point := range series.Data {
			if !internalinteractive.FiniteNumber(point.Value) {
				return fmt.Errorf("line chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
			if err := validateLineSymbol(point.Symbol); err != nil {
				return fmt.Errorf("line chart series %q data point %d %w", series.Name, dataIndex, err)
			}
		}
		if err := validateSeriesOptions(series.Options); err != nil {
			return fmt.Errorf("line chart series %q: %w", series.Name, err)
		}
		if err := validateReferences(series.References); err != nil {
			return fmt.Errorf("line chart series %q: %w", series.Name, err)
		}
	}
	if err := validateSeriesOptions(cfg.SeriesOptions); err != nil {
		return fmt.Errorf("line chart shared series options: %w", err)
	}
	return nil
}

func validateValueAxis(axis ValueAxis) error {
	if len(axis.Values) == 0 {
		return fmt.Errorf("line chart value axis values are required")
	}
	for index, value := range axis.Values {
		if !internalinteractive.FiniteNumber(value) {
			return fmt.Errorf("line chart value axis value %d must be finite", index)
		}
		if index > 0 && value <= axis.Values[index-1] {
			return fmt.Errorf("line chart value axis values must be strictly increasing")
		}
	}
	return nil
}

func validateVisualScale(scale VisualScale) error {
	if scale.Dimension != VisualDimensionX && scale.Dimension != VisualDimensionY {
		return fmt.Errorf("line chart visual scale dimension must be x or y")
	}
	if len(scale.Pieces) == 0 {
		return fmt.Errorf("line chart visual scale pieces are required")
	}
	for index, piece := range scale.Pieces {
		if piece.GreaterThan == nil && piece.LessThan == nil {
			return fmt.Errorf("line chart visual scale piece %d requires a bound", index)
		}
		if piece.GreaterThan != nil && !internalinteractive.FiniteNumber(*piece.GreaterThan) {
			return fmt.Errorf("line chart visual scale piece %d lower bound must be finite", index)
		}
		if piece.LessThan != nil && !internalinteractive.FiniteNumber(*piece.LessThan) {
			return fmt.Errorf("line chart visual scale piece %d upper bound must be finite", index)
		}
		if piece.GreaterThan != nil && piece.LessThan != nil && *piece.GreaterThan >= *piece.LessThan {
			return fmt.Errorf("line chart visual scale piece %d lower bound must be less than upper bound", index)
		}
	}
	return nil
}

func validateSeriesOptions(options chart.SeriesOptions) error {
	if options.Step != "" && options.Step != "start" && options.Step != "middle" && options.Step != "end" {
		return fmt.Errorf("step must be start, middle, or end")
	}
	for _, symbol := range []string{options.Symbol} {
		if err := validateLineSymbol(symbol); err != nil {
			return err
		}
	}
	return nil
}

func validateLineSymbol(symbol string) error {
	switch symbol {
	case "", "circle", "rect", "roundRect", "triangle", "diamond", "pin", "arrow", "none":
		return nil
	default:
		return fmt.Errorf("symbol %q is not supported", symbol)
	}
}

func validateStatistic(statistic Statistic) error {
	switch statistic {
	case StatisticMinimum, StatisticMaximum, StatisticAverage:
		return nil
	default:
		return fmt.Errorf("statistic %q is not supported", statistic)
	}
}

func validateReferences(references References) error {
	if references.SymbolSize < 0 {
		return fmt.Errorf("reference symbol size must be nonnegative")
	}
	for _, symbol := range []string{references.StartSymbol, references.EndSymbol} {
		if err := validateLineReferenceSymbol(symbol); err != nil {
			return fmt.Errorf("reference %w", err)
		}
	}
	for index, point := range references.Points {
		if point.Name == "" {
			return fmt.Errorf("point reference %d name is required", index)
		}
		if err := validateStatistic(point.Statistic); err != nil {
			return fmt.Errorf("point reference %d %w", index, err)
		}
	}
	for index, line := range references.Lines {
		if line.Name == "" {
			return fmt.Errorf("guide reference %d name is required", index)
		}
		modes := 0
		if line.Statistic != "" {
			modes++
			if err := validateStatistic(line.Statistic); err != nil {
				return fmt.Errorf("guide reference %d %w", index, err)
			}
		}
		if line.X != nil {
			modes++
			if !internalinteractive.FiniteNumber(*line.X) {
				return fmt.Errorf("guide reference %d x value must be finite", index)
			}
		}
		if line.Start != nil || line.End != nil {
			modes++
			if line.Start == nil || line.End == nil {
				return fmt.Errorf("guide reference %d requires both coordinates", index)
			}
			if !internalinteractive.FiniteNumber(line.Start.X) || !internalinteractive.FiniteNumber(line.Start.Y) || !internalinteractive.FiniteNumber(line.End.X) || !internalinteractive.FiniteNumber(line.End.Y) {
				return fmt.Errorf("guide reference %d coordinates must be finite", index)
			}
		}
		if modes != 1 {
			return fmt.Errorf("guide reference %d requires exactly one reference mode", index)
		}
	}
	for index, area := range references.Areas {
		if area.Name == "" {
			return fmt.Errorf("range reference %d name is required", index)
		}
		if !internalinteractive.FiniteNumber(area.StartX) || !internalinteractive.FiniteNumber(area.EndX) || area.StartX >= area.EndX {
			return fmt.Errorf("range reference %d requires an increasing finite interval", index)
		}
	}
	return nil
}

func validateLineReferenceSymbol(symbol string) error {
	if symbol == "square" {
		return nil
	}
	return validateLineSymbol(symbol)
}

const (
	lineVisualZeroGreaterThanSentinel = float32(-3.4028212e+38)
	lineVisualZeroLessThanSentinel    = float32(-3.4028222e+38)
)

func rendererVisualScale(scale VisualScale) (opts.VisualMap, []internalinteractive.ScriptReplacement) {
	dimension := "0"
	if scale.Dimension == VisualDimensionY {
		dimension = "1"
	}
	pieces := make([]opts.Piece, len(scale.Pieces))
	replacements := make([]internalinteractive.ScriptReplacement, 0, 2)
	for index, piece := range scale.Pieces {
		if piece.GreaterThan != nil {
			pieces[index].Gt = float32(*piece.GreaterThan)
			if *piece.GreaterThan == 0 {
				pieces[index].Gt = lineVisualZeroGreaterThanSentinel
				replacements = appendLineVisualZeroReplacement(replacements, "gt", lineVisualZeroGreaterThanSentinel)
			}
		}
		if piece.LessThan != nil {
			pieces[index].Lt = float32(*piece.LessThan)
			if *piece.LessThan == 0 {
				pieces[index].Lt = lineVisualZeroLessThanSentinel
				replacements = appendLineVisualZeroReplacement(replacements, "lt", lineVisualZeroLessThanSentinel)
			}
		}
	}
	return opts.VisualMap{Type: "piecewise", Dimension: dimension, Pieces: pieces, Show: opts.Bool(false)}, replacements
}

func appendLineVisualZeroReplacement(replacements []internalinteractive.ScriptReplacement, field string, sentinel float32) []internalinteractive.ScriptReplacement {
	old := `"` + field + `":` + strconv.FormatFloat(float64(sentinel), 'g', -1, 32)
	for _, replacement := range replacements {
		if replacement.Old == old {
			return replacements
		}
	}
	return append(replacements, internalinteractive.ScriptReplacement{Old: old, New: `"` + field + `":0`})
}

func rendererReferences(references References) []charts.SeriesOpts {
	result := make([]charts.SeriesOpts, 0, 6)
	if len(references.Points) > 0 {
		points := make([]opts.MarkPointNameTypeItem, len(references.Points))
		for index, point := range references.Points {
			points[index] = opts.MarkPointNameTypeItem{Name: point.Name, Type: rendererStatistic(point.Statistic)}
		}
		result = append(result, charts.WithMarkPointNameTypeItemOpts(points...))
		if references.ShowLabels != nil {
			result = append(result, charts.WithMarkPointStyleOpts(opts.MarkPointStyle{Label: &opts.Label{Show: opts.Bool(*references.ShowLabels)}}))
		}
	}
	statistics := make([]opts.MarkLineNameTypeItem, 0, len(references.Lines))
	coordinates := make([]opts.MarkLineNameCoordItem, 0, len(references.Lines))
	xValues := make([]opts.MarkLineNameXAxisItem, 0, len(references.Lines))
	for _, line := range references.Lines {
		switch {
		case line.Statistic != "":
			statistics = append(statistics, opts.MarkLineNameTypeItem{Name: line.Name, Type: rendererStatistic(line.Statistic)})
		case line.X != nil:
			xValues = append(xValues, opts.MarkLineNameXAxisItem{Name: line.Name, XAxis: *line.X})
		default:
			coordinates = append(coordinates, opts.MarkLineNameCoordItem{Name: line.Name, Coordinate0: []any{line.Start.X, line.Start.Y}, Coordinate1: []any{line.End.X, line.End.Y}})
		}
	}
	if len(statistics) > 0 {
		result = append(result, charts.WithMarkLineNameTypeItemOpts(statistics...))
	}
	if len(coordinates) > 0 {
		result = append(result, charts.WithMarkLineNameCoordItemOpts(coordinates...))
	}
	if len(xValues) > 0 {
		result = append(result, charts.WithMarkLineNameXAxisItemOpts(xValues...))
	}
	if len(references.Lines) > 0 && (references.ShowLabels != nil || references.StartSymbol != "" || references.EndSymbol != "" || references.SymbolSize != 0) {
		style := opts.MarkLineStyle{SymbolSize: float32(references.SymbolSize)}
		if references.StartSymbol != "" || references.EndSymbol != "" {
			style.Symbol = []string{references.StartSymbol, references.EndSymbol}
		}
		if references.ShowLabels != nil {
			style.Label = &opts.Label{Show: opts.Bool(*references.ShowLabels)}
		}
		result = append(result, charts.WithMarkLineStyleOpts(style))
	}
	if len(references.Areas) > 0 {
		areas := make([][]opts.MarkAreaData, len(references.Areas))
		for index, area := range references.Areas {
			areas[index] = []opts.MarkAreaData{{Name: area.Name, XAxis: area.StartX}, {XAxis: area.EndX}}
		}
		result = append(result, charts.WithMarkAreaData(areas...))
		if references.ShowLabels != nil {
			result = append(result, charts.WithMarkAreaStyleOpts(opts.MarkAreaStyle{Label: &opts.Label{Show: opts.Bool(*references.ShowLabels), Position: "middle"}}))
		}
	}
	return result
}

func rendererStatistic(statistic Statistic) string {
	switch statistic {
	case StatisticMinimum:
		return "min"
	case StatisticMaximum:
		return "max"
	default:
		return "average"
	}
}

func validateTimeAxis(axis TimeAxis) error {
	if axis.Minimum.IsZero() {
		return fmt.Errorf("line chart time axis minimum is required")
	}
	if len(axis.Values) == 0 {
		return fmt.Errorf("line chart time axis values are required")
	}
	if axis.SplitNumber < 0 {
		return fmt.Errorf("line chart time axis split number must be nonnegative")
	}
	minimum := axis.Minimum.UTC()
	previous := time.Time{}
	for index, value := range axis.Values {
		if value.IsZero() {
			return fmt.Errorf("line chart time axis value %d is required", index)
		}
		value = value.UTC()
		if value.Before(minimum) {
			return fmt.Errorf("line chart time axis value %d precedes minimum", index)
		}
		if index > 0 && !value.After(previous) {
			return fmt.Errorf("line chart time axis values must be strictly chronological")
		}
		previous = value
	}
	return nil
}

func lineTimeSplitNumber(axis TimeAxis) int {
	if axis.SplitNumber == 0 {
		return 4
	}
	return axis.SplitNumber
}

func lineTimeValue(value time.Time) string { return value.UTC().Format(time.RFC3339) }

func lineExactValuesTitle(cfg Config) string {
	if cfg.TimeAxis != nil {
		return "Exact time and values"
	}
	if cfg.ValueAxis != nil {
		return "Exact x and values"
	}
	return "Exact category and values"
}

func lineDetails(cfg Config) templ.Component {
	if cfg.Live != nil {
		return nil
	}
	return lineExactValues(cfg)
}

func lineExactValuesAttributes(cfg Config) templ.Attributes {
	attributes := templ.Attributes{"data-line-exact-values": ""}
	if cfg.TimeAxis != nil {
		attributes["data-line-time-exact-values"] = ""
	}
	return attributes
}

func lineXAxisHeading(cfg Config) string {
	if cfg.TimeAxis != nil {
		return "Time"
	}
	if cfg.ValueAxis != nil {
		return "X"
	}
	return "Category"
}

func lineXAxisLength(cfg Config) int {
	if cfg.TimeAxis != nil {
		return len(cfg.TimeAxis.Values)
	}
	if cfg.ValueAxis != nil {
		return len(cfg.ValueAxis.Values)
	}
	return len(cfg.XAxis)
}

func lineXAxisValue(cfg Config, index int) string {
	if cfg.TimeAxis != nil {
		return lineTimeValue(cfg.TimeAxis.Values[index])
	}
	if cfg.ValueAxis != nil {
		return strconv.FormatFloat(cfg.ValueAxis.Values[index], 'g', -1, 64)
	}
	return cfg.XAxis[index]
}
