package interactive

import (
	"fmt"
	"strconv"
	"time"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// LineConfig describes an accessible, browser-rendered line chart.
//
// Values must be application-owned because the browser renderer serializes them.
type LineConfig struct {
	Label   string
	Caption string
	XAxis   []string
	// TimeAxis selects a temporal x axis. It is mutually exclusive with XAxis.
	// LiveData remains categorical because CartesianSnapshot carries categories.
	TimeAxis *LineTimeAxis
	// ValueAxis selects ordered numerical x coordinates. It is mutually
	// exclusive with XAxis and TimeAxis.
	ValueAxis     *LineValueAxis
	Series        []LineSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
	Live          *LiveData
	// VisualScale maps x or y values through a theme-aware piecewise scale.
	VisualScale *LineVisualScale
}

// LineTimeAxis defines ordered instants and a required inclusive lower bound.
// Values and Minimum use time.Time so callers never supply renderer values.
type LineTimeAxis struct {
	Values  []time.Time
	Minimum time.Time
	// SplitNumber recommends readable temporal tick density. Zero uses the
	// responsive default of four segments; the private renderer also hides
	// any remaining overlap while retaining endpoint labels.
	SplitNumber int
}

// LineValueAxis defines ordered finite numerical x coordinates.
type LineValueAxis struct {
	Values []float64
}

// LineVisualDimension selects the coordinate used by a piecewise visual scale.
type LineVisualDimension string

const (
	// LineVisualDimensionX maps x coordinates.
	LineVisualDimensionX LineVisualDimension = "x"
	// LineVisualDimensionY maps y values.
	LineVisualDimensionY LineVisualDimension = "y"
)

// LineVisualScale defines theme-aware numeric pieces. Bounds are optional and
// exclusive, matching a greater-than and less-than interval.
type LineVisualScale struct {
	Dimension LineVisualDimension
	Pieces    []LineVisualPiece
}

// LineVisualPiece defines one open numeric interval.
type LineVisualPiece struct {
	GreaterThan *float64
	LessThan    *float64
}

// LineStatistic selects a calculated series reference.
type LineStatistic string

const (
	// LineStatisticMinimum selects the smallest series value.
	LineStatisticMinimum LineStatistic = "minimum"
	// LineStatisticMaximum selects the largest series value.
	LineStatisticMaximum LineStatistic = "maximum"
	// LineStatisticAverage selects the arithmetic mean of series values.
	LineStatisticAverage LineStatistic = "average"
)

// LineCoordinate identifies one finite x/y point. X is a numeric coordinate
// or a zero-based category index.
type LineCoordinate struct {
	X float64
	Y float64
}

// LinePointReference places a calculated point on a series.
type LinePointReference struct {
	Name      string
	Statistic LineStatistic
}

// LineGuideReference draws a calculated, vertical, or two-coordinate guide.
// X is a numeric coordinate or a zero-based category index.
// Exactly one of Statistic, X, or Start and End must be configured.
type LineGuideReference struct {
	Name      string
	Statistic LineStatistic
	X         *float64
	Start     *LineCoordinate
	End       *LineCoordinate
}

// LineRangeReference highlights an inclusive numeric x-axis interval or an
// inclusive interval between zero-based category indexes.
type LineRangeReference struct {
	Name   string
	StartX float64
	EndX   float64
}

// LineReferences configures theme-aware points, guides, and ranges.
type LineReferences struct {
	Points      []LinePointReference
	Lines       []LineGuideReference
	Areas       []LineRangeReference
	ShowLabels  *bool
	StartSymbol string
	EndSymbol   string
	SymbolSize  int
}

// LineSeries describes one named line series.
type LineSeries struct {
	Name       string
	Data       []LineData
	Options    SeriesOptions
	References LineReferences
}

// LineData describes one finite line value and optional point symbol.
type LineData struct {
	Name       string
	Value      float64
	Symbol     string
	SymbolSize int
}

// Line builds a reusable interactive line component.
func Line(cfg LineConfig) Instance {
	if err := validateLineConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveLine, err)
	}

	chart := charts.NewLine()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
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
	visualScaleReplacements := []scriptReplacement(nil)
	if cfg.VisualScale != nil {
		visualScale, replacements := rendererLineVisualScale(*cfg.VisualScale)
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
		seriesOptions := mergeSeriesOptions(cfg.SeriesOptions, series.Options)
		seriesOptions = append(seriesOptions, rendererLineReferences(series.References)...)
		chart.AddSeries(series.Name, data, seriesOptions...)
	}

	return newInstance(chartcomponents.KindInteractiveLine, renderConfig{
		Label:              cfg.Label,
		Caption:            cfg.Caption,
		Chart:              chart,
		ResponsiveWidth:    responsiveWidth(cfg.Width),
		Style:              cfg.Style,
		Live:               cartesianLiveConfig(cfg.Live),
		Details:            lineDetails(cfg),
		Animation:          cfg.Options.Animation,
		Controls:           cfg.Options.Controls,
		Export:             cfg.Options.Export,
		AxisLabelIntervals: axisLabelIntervals(cfg.Options),
		ScriptReplacements: visualScaleReplacements,
	})
}

func validateLineConfig(cfg LineConfig) error {
	if err := validateChartOptions(cfg.Options); err != nil {
		return err
	}
	if err := validateLiveData(cfg.Live); err != nil {
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
		if err := validateLineTimeAxis(*cfg.TimeAxis); err != nil {
			return err
		}
	}
	if cfg.ValueAxis != nil {
		if cfg.Live != nil {
			return fmt.Errorf("line chart live data supports categorical x axis only")
		}
		if err := validateLineValueAxis(*cfg.ValueAxis); err != nil {
			return err
		}
	}
	if cfg.VisualScale != nil {
		if err := validateLineVisualScale(*cfg.VisualScale); err != nil {
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
			if !finiteNumber(point.Value) {
				return fmt.Errorf("line chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
			if err := validateLineSymbol(point.Symbol); err != nil {
				return fmt.Errorf("line chart series %q data point %d %w", series.Name, dataIndex, err)
			}
		}
		if err := validateLineSeriesOptions(series.Options); err != nil {
			return fmt.Errorf("line chart series %q: %w", series.Name, err)
		}
		if err := validateLineReferences(series.References); err != nil {
			return fmt.Errorf("line chart series %q: %w", series.Name, err)
		}
	}
	if err := validateLineSeriesOptions(cfg.SeriesOptions); err != nil {
		return fmt.Errorf("line chart shared series options: %w", err)
	}
	return nil
}

func validateLineValueAxis(axis LineValueAxis) error {
	if len(axis.Values) == 0 {
		return fmt.Errorf("line chart value axis values are required")
	}
	for index, value := range axis.Values {
		if !finiteNumber(value) {
			return fmt.Errorf("line chart value axis value %d must be finite", index)
		}
		if index > 0 && value <= axis.Values[index-1] {
			return fmt.Errorf("line chart value axis values must be strictly increasing")
		}
	}
	return nil
}

func validateLineVisualScale(scale LineVisualScale) error {
	if scale.Dimension != LineVisualDimensionX && scale.Dimension != LineVisualDimensionY {
		return fmt.Errorf("line chart visual scale dimension must be x or y")
	}
	if len(scale.Pieces) == 0 {
		return fmt.Errorf("line chart visual scale pieces are required")
	}
	for index, piece := range scale.Pieces {
		if piece.GreaterThan == nil && piece.LessThan == nil {
			return fmt.Errorf("line chart visual scale piece %d requires a bound", index)
		}
		if piece.GreaterThan != nil && !finiteNumber(*piece.GreaterThan) {
			return fmt.Errorf("line chart visual scale piece %d lower bound must be finite", index)
		}
		if piece.LessThan != nil && !finiteNumber(*piece.LessThan) {
			return fmt.Errorf("line chart visual scale piece %d upper bound must be finite", index)
		}
		if piece.GreaterThan != nil && piece.LessThan != nil && *piece.GreaterThan >= *piece.LessThan {
			return fmt.Errorf("line chart visual scale piece %d lower bound must be less than upper bound", index)
		}
	}
	return nil
}

func validateLineSeriesOptions(options SeriesOptions) error {
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

func validateLineStatistic(statistic LineStatistic) error {
	switch statistic {
	case LineStatisticMinimum, LineStatisticMaximum, LineStatisticAverage:
		return nil
	default:
		return fmt.Errorf("statistic %q is not supported", statistic)
	}
}

func validateLineReferences(references LineReferences) error {
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
		if err := validateLineStatistic(point.Statistic); err != nil {
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
			if err := validateLineStatistic(line.Statistic); err != nil {
				return fmt.Errorf("guide reference %d %w", index, err)
			}
		}
		if line.X != nil {
			modes++
			if !finiteNumber(*line.X) {
				return fmt.Errorf("guide reference %d x value must be finite", index)
			}
		}
		if line.Start != nil || line.End != nil {
			modes++
			if line.Start == nil || line.End == nil {
				return fmt.Errorf("guide reference %d requires both coordinates", index)
			}
			if !finiteNumber(line.Start.X) || !finiteNumber(line.Start.Y) || !finiteNumber(line.End.X) || !finiteNumber(line.End.Y) {
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
		if !finiteNumber(area.StartX) || !finiteNumber(area.EndX) || area.StartX >= area.EndX {
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

func rendererLineVisualScale(scale LineVisualScale) (opts.VisualMap, []scriptReplacement) {
	dimension := "0"
	if scale.Dimension == LineVisualDimensionY {
		dimension = "1"
	}
	pieces := make([]opts.Piece, len(scale.Pieces))
	replacements := make([]scriptReplacement, 0, 2)
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

func appendLineVisualZeroReplacement(replacements []scriptReplacement, field string, sentinel float32) []scriptReplacement {
	old := `"` + field + `":` + strconv.FormatFloat(float64(sentinel), 'g', -1, 32)
	for _, replacement := range replacements {
		if replacement.Old == old {
			return replacements
		}
	}
	return append(replacements, scriptReplacement{Old: old, New: `"` + field + `":0`})
}

func rendererLineReferences(references LineReferences) []charts.SeriesOpts {
	result := make([]charts.SeriesOpts, 0, 6)
	if len(references.Points) > 0 {
		points := make([]opts.MarkPointNameTypeItem, len(references.Points))
		for index, point := range references.Points {
			points[index] = opts.MarkPointNameTypeItem{Name: point.Name, Type: rendererLineStatistic(point.Statistic)}
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
			statistics = append(statistics, opts.MarkLineNameTypeItem{Name: line.Name, Type: rendererLineStatistic(line.Statistic)})
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

func rendererLineStatistic(statistic LineStatistic) string {
	switch statistic {
	case LineStatisticMinimum:
		return "min"
	case LineStatisticMaximum:
		return "max"
	default:
		return "average"
	}
}

func validateLineTimeAxis(axis LineTimeAxis) error {
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

func lineTimeSplitNumber(axis LineTimeAxis) int {
	if axis.SplitNumber == 0 {
		return 4
	}
	return axis.SplitNumber
}

func lineTimeValue(value time.Time) string { return value.UTC().Format(time.RFC3339) }

func lineExactValuesTitle(cfg LineConfig) string {
	if cfg.TimeAxis != nil {
		return "Exact time and values"
	}
	if cfg.ValueAxis != nil {
		return "Exact x and values"
	}
	return "Exact category and values"
}

func lineDetails(cfg LineConfig) templ.Component {
	if cfg.Live != nil {
		return nil
	}
	return lineExactValues(cfg)
}

func lineExactValuesAttributes(cfg LineConfig) templ.Attributes {
	attributes := templ.Attributes{"data-line-exact-values": ""}
	if cfg.TimeAxis != nil {
		attributes["data-line-time-exact-values"] = ""
	}
	return attributes
}

func lineXAxisHeading(cfg LineConfig) string {
	if cfg.TimeAxis != nil {
		return "Time"
	}
	if cfg.ValueAxis != nil {
		return "X"
	}
	return "Category"
}

func lineXAxisLength(cfg LineConfig) int {
	if cfg.TimeAxis != nil {
		return len(cfg.TimeAxis.Values)
	}
	if cfg.ValueAxis != nil {
		return len(cfg.ValueAxis.Values)
	}
	return len(cfg.XAxis)
}

func lineXAxisValue(cfg LineConfig, index int) string {
	if cfg.TimeAxis != nil {
		return lineTimeValue(cfg.TimeAxis.Values[index])
	}
	if cfg.ValueAxis != nil {
		return strconv.FormatFloat(cfg.ValueAxis.Values[index], 'g', -1, 64)
	}
	return cfg.XAxis[index]
}
